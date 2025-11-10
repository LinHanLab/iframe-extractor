package main

import (
	"path/filepath"
	"strings"
)

func ConvertURLToFilename(url string) string {
	parts := strings.Split(url, "/")
	filename := parts[len(parts)-1]

	if filename == "" {
		if len(parts) > 1 {
			filename = parts[len(parts)-2]
		} else {
			filename = "index"
		}
	}

	if queryStart := strings.Index(filename, "?"); queryStart != -1 {
		filename = filename[:queryStart]
	}

	// Remove existing file extension to avoid double extensions like .asp.html
	// For example: default.asp -> default (will become default.html later)
	if ext := filepath.Ext(filename); ext != "" {
		filename = strings.TrimSuffix(filename, ext)
	}

	return filename
}
