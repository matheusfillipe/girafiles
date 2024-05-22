package api

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const FILEDIR = "data"

var storageLock = &sync.Mutex{}

type Node struct {
	name      string
	ip        string
	timestamp int64
	reader    io.Reader
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
		ip:        ip,
		timestamp: time.Now().UTC().Unix(),
		reader:    file,
	}, nil
}

func Upload(file *multipart.FileHeader, ip string) (string, error) {
	storageLock.Lock()
	defer storageLock.Unlock()

	var settings = GetSettings()

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Create buffer to read multiple times from memory
	buf, err := io.ReadAll(src)
	if err != nil {
		return "", err
	}

	if len(buf)/1024/1024 > settings.FileSizeLimit {
		return "", fmt.Errorf("File size limit exceeded. Limit is %dMB", settings.FileSizeLimit)
	}

	if len(buf) == 0 {
		return "", fmt.Errorf("File is empty")
	}

	// Parse file and create node
	node, err := newNode(bytes.NewReader(buf), filepath.Ext(file.Filename), ip)
	if err != nil {
		return "", err
	}

	// Create data directory if it doesn't exist
	dir := filepath.Join(settings.StorePath, FILEDIR)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if os.MkdirAll(dir, os.ModePerm) != nil {
			return node.name, err
		}
	}

	// Write file to disk
	dst := filepath.Join(dir, node.name)
	out, err := os.Create(dst)
	if err != nil {
		return node.name, err
	}
	defer out.Close()
	_, err = io.Copy(out, bytes.NewReader(buf))
	if err != nil {
		return node.name, err
	}

	// Write node to database
	err = GetDB().insertNode(node)
	if err != nil {
		// If it failes we also delete the file
		if err.Error() != DUP_ENTRY_ERROR {
			os.Remove(dst)
		}
		return node.name, err
	}

	return node.name, err
}
func Download(n string) (string, []byte, error) {
	storageLock.Lock()
	defer storageLock.Unlock()

	var settings = GetSettings()
	err, name := GetDB().checkFileName(n)
	if err != nil {
		log.Println(err)
		return "", nil, err
	}

	dst := filepath.Join(settings.GetFileStoragePath(), name)
	b, err := os.ReadFile(dst)

	if err != nil {
		log.Println(err)
		return "", nil, err
	}
	m := http.DetectContentType(b[:512])

	return m, b, nil
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

func touchFile(path string) error {
	storageLock.Lock()
	defer storageLock.Unlock()

	cdb := GetDB()

	// Get the current time
	now := time.Now()
	if err := cdb.updateTimestamp(path, now.UTC().Unix()); err != nil {
		return err
	}

	// Update the modification time of the file to the current time
	err := os.Chtimes(path, now, now)
	if err != nil {
		panic(err)
	}
	return nil
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
