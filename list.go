package main

import (
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
)

type FileData struct {
	Name string `json:"name"`
	Link string `json:"link"`
	Type string `json:"type"`
	Size string `json:"size"`
	Icon string `json:"icon"`
}

type FilesListData struct {
	Path  string     `json:"path"`
	Files []FileData `json:"files"`
}

func ListFilesWeb(w http.ResponseWriter, r *http.Request) {
	// Use r.URL.Path, not r.RequestURI: the latter is raw and carries the query
	// string, which the mux does not clean. Cleaning against "/" strips any ".."
	// segments, so the joined path can never escape baseDir.
	urlPath := path.Clean("/" + r.URL.Path)
	fullPath := filepath.Join(baseDir, filepath.FromSlash(urlPath))

	info, err := os.Stat(fullPath)
	if err != nil {
		Error(http.StatusInternalServerError, "Unable to list files", w)
		return
	}
	if info.IsDir() {
		data, err := GetListFilesData(urlPath)
		if err != nil {
			Error(err.Status, err.Description, w)
			return
		}
		t, _ := template.ParseFiles("templates/list.html")
		_ = t.Execute(w, data)
	} else {
		name := filepath.Base(fullPath)
		file, err := os.Open(fullPath)
		if err != nil {
			Error(http.StatusInternalServerError, "Unable to serve file", w)
			return
		}
		//goland:noinspection GoUnhandledErrorResult
		defer file.Close()
		http.ServeContent(w, r, name, info.ModTime(), file)
	}
}

func GetListFilesData(urlPath string) (*FilesListData, *ErrorResult) {
	dir := filepath.Join(baseDir, filepath.FromSlash(urlPath))
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
		// Names may contain characters that are structural in a URL ("?", "#",
		// space), so escape the link instead of pasting the raw name into href.
		link := (&url.URL{Path: path.Join(urlPath, entry.Name())}).String()
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
			Link: link,
			Type: contentType,
			Size: size,
			Icon: icon,
		}
	}
	return &FilesListData{
		Path:  urlPath,
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

	n, err := file.Read(buf)

	// A short file is not an error: io.EOF just means there is less than 512
	// bytes to sniff. Only the bytes actually read may be handed to the
	// detector, or the zero padding makes every small text file look binary.
	if err != nil && err != io.EOF {
		return "", err
	}

	// the function that actually does the trick
	contentType := http.DetectContentType(buf[:n])

	return contentType, nil
}
