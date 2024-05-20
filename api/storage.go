package api

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const FILEDIR = "data"

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
		timestamp: time.Now().Unix(),
		reader:    file,
	}, nil
}

func Upload(file *multipart.FileHeader, ip string) (string, error) {
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

	// Parse file and create node
	node, err := newNode(bytes.NewReader(buf), filepath.Ext(file.Filename), ip)
	if err != nil {
		return "", err
	}

	// Create data directory if it doesn't exist
	dir := filepath.Join(settings.StorePath, FILEDIR)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if os.MkdirAll(dir, os.ModePerm) != nil {
			return "", err
		}
	}

	// Write file to disk
	dst := filepath.Join(dir, node.name)
	out, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer out.Close()
	_, err = io.Copy(out, bytes.NewReader(buf))
	if err != nil {
		return "", err
	}

	// Write node to database
	err = GetDB().insertNode(node)
	if err != nil {
		// If it failes we also delete the file
		os.Remove(dst)
		return "", err
	}

	return node.name, err
}
func Download(n string) (string, []byte, error) {
	var settings = GetSettings()
	err, name := GetDB().checkFileName(n)
	if err != nil {
		log.Println(err)
		return "", nil, err
	}

	dst := filepath.Join(settings.StorePath, FILEDIR, name)
	b, err := os.ReadFile(dst)

	if err != nil {
		log.Println(err)
		return "", nil, err
	}
	m := http.DetectContentType(b[:512])

	return m, b, nil
}
