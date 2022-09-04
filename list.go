package main

import (
	"html/template"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
)

type FileData struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Type string `json:"type"`
	Size string `json:"size"`
	Icon string `json:"icon"`
}

type FilesListData struct {
	Path  string     `json:"path"`
	Files []FileData `json:"files"`
}

func ListFilesWeb(w http.ResponseWriter, r *http.Request) {
	uri := r.RequestURI
	decodedUri, _ := url.QueryUnescape(uri)
	path := filepath.Join(baseDir, decodedUri)

	info, err := os.Stat(path)
	if err != nil {
		Error(http.StatusInternalServerError, "Unable to list files", w)
		return
	}
	if info.IsDir() {
		data, err := GetListFilesData(decodedUri)
		if err != nil {
			Error(err.Status, err.Description, w)
			return
		}
		t, _ := template.ParseFiles("templates/list.html")
		_ = t.Execute(w, data)
	} else {
		name := filepath.Base(path)
		file, err := os.Open(path)
		if err != nil {
			Error(http.StatusInternalServerError, "Unable to serve file", w)
			return
		}
		http.ServeContent(w, r, name, info.ModTime(), file)
	}
}

func GetListFilesData(path string) (*FilesListData, *ErrorResult) {
	dir := filepath.Join(baseDir, path)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, &ErrorResult{http.StatusInternalServerError, "Unable to list files"}
	}
	files := make([]FileData, len(entries))
	sort.Slice(entries, func(i, j int) bool { return entries[i].IsDir() })
	for i, entry := range entries {
		var contentType string
		var size string
		var icon string
		abs := filepath.Join(dir, entry.Name())
		rel := filepath.Join(path, entry.Name())
		if entry.IsDir() {
			contentType = "directory"
			size = ""
			icon = "dir"
		} else {
			contentType, _ = detectContentType(abs)
			info, err := entry.Info()
			if err != nil {
				return nil, &ErrorResult{http.StatusInternalServerError, "Unable to get file info"}
			}
			size = ByteCountIEC(info.Size())
			icon = "file"
		}

		files[i] = FileData{
			Name: entry.Name(),
			Path: rel,
			Type: contentType,
			Size: size,
			Icon: icon,
		}
	}
	return &FilesListData{
		Path:  path,
		Files: files,
	}, nil
}

func detectContentType(name string) (string, error) {
	file, err := os.Open(name)
	if err != nil {
		return "", err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer file.Close()
	// Get the file content
	contentType, err := GetFileContentType(file)
	if err != nil {
		return "", err
	}
	return contentType, nil
}

func GetFileContentType(file *os.File) (string, error) {
	// to sniff the content type only the first
	// 512 bytes are used.

	buf := make([]byte, 512)

	_, err := file.Read(buf)

	if err != nil {
		return "", err
	}

	// the function that actually does the trick
	contentType := http.DetectContentType(buf)

	return contentType, nil
}
