package utils

import (
	"path/filepath"
	"strings"
)

func AllowedFileExtension(filename string, fileType string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch fileType {
	case "image":
		switch ext {
		case ".jpg", ".jpeg", ".png":
			return true
		}
	case "pdf":
		switch ext {
		case ".pdf", ".doc":
			return true
		}
	}
	return false
}
