package main

import (
	"fmt"
	"github.com/gorilla/handlers"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const UPLOAD_PATH="/storage"

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		log.Println("Error: Method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 32 MB is the default used by FormFile
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

    maxUploadSize, err := strconv.ParseInt(os.Getenv("MAX_UPLOAD_SIZE"), 10, 64)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// get a reference to the fileHeaders
	files := r.MultipartForm.File["file"]

	for _, fileHeader := range files {
		if fileHeader.Size > maxUploadSize {
			http.Error(w, fmt.Sprintf("The uploaded file is too large: %s.", fileHeader.Filename), http.StatusBadRequest)
			return
		}

		srcFile, err := fileHeader.Open()
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer srcFile.Close()

		_, err = srcFile.Seek(0, io.SeekStart)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = os.MkdirAll(UPLOAD_PATH, os.ModePerm)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		destFile, err := os.Create(fmt.Sprintf("%s/%d%s", UPLOAD_PATH, time.Now().UnixNano(), filepath.Ext(fileHeader.Filename)))
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		defer destFile.Close()

		_, err = io.Copy(destFile, srcFile)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		log.Printf("Upload complete to %s", UPLOAD_PATH)
	}

	fmt.Fprintf(w, "Upload successful")
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/upload", uploadHandler)

    log.Println("Server started")
    log.Printf("Upload size limited to %s bytes", os.Getenv("MAX_UPLOAD_SIZE"))
	if err := http.ListenAndServe(":4500", handlers.LoggingHandler(os.Stdout, mux)); err != nil {
		log.Fatal(err)
	}
}
