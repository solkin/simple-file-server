package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
)

var baseDir string

func main() {
	log.Println("start")

	var port int
	flag.IntVar(&port, "port", 8080, "port for incoming connections")
	flag.StringVar(&baseDir, "dir", "", "directory for static files serving")
	flag.Parse()

	baseDirInfo, err := os.Stat(baseDir)
	if os.IsNotExist(err) || !baseDirInfo.IsDir() {
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
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/", ListFilesWeb)
}
