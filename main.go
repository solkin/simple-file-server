package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
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

	srv := &http.Server{Addr: fmt.Sprintf(":%d", port)}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	sig := make(chan os.Signal, 1)

	signal.Notify(sig, os.Interrupt)

	<-sig

	log.Println("shutting down")

	// Give in-flight requests a moment to finish instead of dropping them.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Println(err)
	}
}

func registerApiHandlers() {
	http.Handle("/static/", http.FileServer(http.FS(staticFS)))
	http.HandleFunc("/", ListFilesWeb)
}
