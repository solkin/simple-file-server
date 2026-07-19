package main

import (
	"embed"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
)

//go:embed templates/list.html
var templateFS embed.FS

//go:embed static
var staticFS embed.FS

// Parsed once at startup and embedded into the binary, so the server no longer
// depends on templates/ and static/ sitting next to the working directory.
var listTemplate = template.Must(template.ParseFS(templateFS, "templates/list.html"))

var baseDir string

func main() {
	log.Println("start")

	var port int
	flag.IntVar(&port, "port", 8080, "port for incoming connections")
	flag.StringVar(&baseDir, "dir", "", "directory for static files serving")
	flag.Parse()

	// Check err first: on any error (missing path, permission denied, empty
	// -dir) baseDirInfo is nil, so calling IsDir on it would panic.
	baseDirInfo, err := os.Stat(baseDir)
	if err != nil || !baseDirInfo.IsDir() {
		log.Fatal("Directory does not exist or is not a directory.")
	}

	log.Println("register handlers")
	registerApiHandlers()

	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
			log.Println(err)
		}
	}()

	sig := make(chan os.Signal, 1)

	signal.Notify(sig, os.Interrupt)

	<-sig

	log.Println("shutting down")
	os.Exit(0)
}

func registerApiHandlers() {
	http.Handle("/static/", http.FileServer(http.FS(staticFS)))
	http.HandleFunc("/", ListFilesWeb)
}
