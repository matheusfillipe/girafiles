package api

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"log/slog"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"github.com/gabriel-vasile/mimetype"
)

const FILEDIR = "data"

var storageLock = &sync.Mutex{}

type Node struct {
	name      string
	shortname string
	extension string
	ip        string
	timestamp int64
	reader    io.Reader
}

type fileResponse struct {
	name     string
	shortname string
	mimetype string
	content  []byte
}

func getFileHash(reader io.Reader) (string, error) {
	md5Hash := md5.New()
	_, err := io.Copy(md5Hash, reader)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", md5Hash.Sum(nil)), nil
}

func newNode(file io.Reader, extension string, ip string) (*Node, error) {
	hash, err := getFileHash(file)
	if err != nil {
		return nil, err
	}
	return &Node{
		name:      hash + extension,
		extension: extension,
		ip:        ip,
		timestamp: time.Now().UTC().Unix(),
		reader:    file,
	}, nil
}

func saveToDisk(src io.Reader, filename string, ip string) (string, *Node, error) {
	var settings = GetSettings()

	// Create buffer to read multiple times from memory
	buf, err := io.ReadAll(src)
	if err != nil {
		return "", nil, err
	}
	if len(buf) == 0 {
		return "", nil, fmt.Errorf("File is empty")
	}
	if len(buf)/1024/1024 > settings.FileSizeLimit {
		return "", nil, fmt.Errorf("File size limit exceeded. Limit is %dMB", settings.FileSizeLimit)
	}

	// Parse file and create node
	node, err := newNode(bytes.NewReader(buf), filepath.Ext(filename), ip)
	if err != nil {
		return "", nil, err
	}

	// Create data directory if it doesn't exist
	dir := filepath.Join(settings.StorePath, FILEDIR)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if os.MkdirAll(dir, os.ModePerm) != nil {
			return "", node, err
		}
	}

	// Write file to disk
	dst := filepath.Join(dir, node.name)
	out, err := os.Create(dst)
	if err != nil {
		return "", node, err
	}
	defer out.Close()
	_, err = io.Copy(out, bytes.NewReader(buf))
	if err != nil {
		return "", node, err
	}
	return dst, node, nil
}

func handleDbUploadErr(err error, dst string, node *Node) (string, error) {
	db := GetDB()

	// If it fails we also delete the file
	if err.Error() == DUP_ENTRY_ERROR {
		shortname, errdb := db.getShortnameForFilename(node.name)
		if errdb != nil {
			return "", fmt.Errorf("Failed to get filename! %s", errdb.Error())
		}
		return shortname, err
	}
	if err.Error() == DUP_ALIAS_ERROR {
		os.Remove(dst)
		return node.shortname, fmt.Errorf("This bucket/alias is already in use")
	}

	os.Remove(dst)
	return node.shortname, err
}

func Upload(file *multipart.FileHeader, ip string) (string, error) {
	storageLock.Lock()
	defer storageLock.Unlock()
	db := GetDB()

	if err := db.CheckRateLimit(ip); err != nil {
		return "", err
	}

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Upload file to disk
	dst, node, err := saveToDisk(src, file.Filename, ip)
	if err != nil {
		return "", err
	}

	// Write node to database
	err = db.insertNode(node)
	if err != nil {
		return handleDbUploadErr(err, dst, node)
	}
	return node.shortname, err
}

func UploadToBucket(src io.Reader, ip string, bucket string, name string) (string, error) {
	storageLock.Lock()
	defer storageLock.Unlock()
	db := GetDB()

	if err := db.CheckRateLimit(ip); err != nil {
		return "", err
	}

	// Upload file to disk
	dst, node, err := saveToDisk(src, name, ip)
	if err != nil {
		return "", err
	}

	// Write node to database
	err = db.insertAlias(bucket, name, node)
	if err != nil {
		return handleDbUploadErr(err, dst, node)
	}
	return node.shortname, err
}

func loadFromDisk(name string, shortname string) (fileResponse, error) {
	settings := GetSettings()

	name = strings.SplitN(name, "@", 2)[0]

	dst := filepath.Join(settings.GetFileStoragePath(), name)
	b, err := os.ReadFile(dst)

	if err != nil {
		log.Println(err)
		return fileResponse{}, err
	}
	m := mimetype.Detect(b).String()

	return fileResponse{
	  shortname: shortname,
		name:     name,
		mimetype: m,
		content:  b,
	}, nil
}

func Download(n string) (fileResponse, error) {
	storageLock.Lock()
	defer storageLock.Unlock()

	name, err := GetDB().checkShortName(n)
	if err != nil {
		log.Println(err)
		return fileResponse{}, err
	}

	return loadFromDisk(name,n)
}

func DownloadFromBucket(bucket string, alias string) (fileResponse, error) {
	storageLock.Lock()
	defer storageLock.Unlock()

	name, err := GetDB().checkAlias(bucket, alias)
	if err != nil {
		log.Println(err)
		return fileResponse{}, err
	}
	return loadFromDisk(name,alias)
}

func isStorageLimitExceeded() bool {
	settings := GetSettings()

	if settings.StorePathSizeLimit == 0 {
		return false
	}
	var size int64 = 0
	err := filepath.Walk(settings.GetFileStoragePath(), func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		size += info.Size()
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	slog.Debug(fmt.Sprintf("Storage size: %dMB > %dMB", size/1024/1024, settings.StorePathSizeLimit))
	return size/1024/1024 > int64(settings.StorePathSizeLimit)
}

// Handle storage size after upload requests
func cleanup() {
	storageLock.Lock()
	defer storageLock.Unlock()

	settings := GetSettings()
	cdb := GetDB()
	slog.Info("Checking what we can delete")

	// Delete expired files
	namesToDelete, err := cdb.deleteExpiredFiles()
	if err != nil {
		slog.Error(fmt.Sprintf("Error deleting expired files: %s", err))
	}
	if len(namesToDelete) == 0 {
		slog.Debug("No expired files to delete")
	}
	for _, name := range namesToDelete {
		err := os.Remove(filepath.Join(settings.GetFileStoragePath(), name))
		if err != nil {
			slog.Error(fmt.Sprintf("Error deleting file %s: %s", name, err))
			continue
		}
		slog.Info(fmt.Sprintf("Deleted file %s because it expired", name))
	}

	// Delete oldest file if storage limit is exceeded
	if isStorageLimitExceeded() {
		namesToDelete, err := cdb.deleteOldestFiles(1)
		if err != nil {
			slog.Error(fmt.Sprintf("Error deleting oldest files from database: %s", err))
			return
		}
		if len(namesToDelete) == 0 {
			slog.Debug("No old files to delete")
		}
		for _, name := range namesToDelete {
			err := os.Remove(filepath.Join(settings.GetFileStoragePath(), name))
			if err != nil {
				slog.Error(fmt.Sprintf("Error deleting file %s: %s", name, err))
				continue
			}
			slog.Info(fmt.Sprintf("Deleted file %s because storage limit was exceeded", name))
		}
	} else {
		slog.Debug("Storage limit not exceeded")
	}
}

func humanReadableSize(size int) string {
    if size < 1000 {
        return fmt.Sprintf("%d B", size)
    }

    // Define units
    units := []string{"KB", "MB", "GB", "TB", "PB", "EB"}

    // Calculate the index for units
    sizeFloat := float64(size)
    i := 0

    // Determine the appropriate unit
    for sizeFloat >= 1000 && i < len(units) {
        sizeFloat /= 1000
        i++
    }

    // Round to two decimal places
    sizeFloat = (sizeFloat*100 + 0.5) / 100 // This ensures proper rounding
    return fmt.Sprintf("%.2f %s", sizeFloat, units[i-1])
}