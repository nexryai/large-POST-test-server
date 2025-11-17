package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

const saveDir = "saved"

var tmpl = template.Must(template.ParseFiles("index.gohtml"))

func handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	err := tmpl.Execute(w, nil)
	if err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(10 << 20) // 10MB
	if err != nil {
		log.Printf("Failed to parse multipart form: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		log.Printf("Failed to get file from form: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.Printf("Received file: %s (Size: %d bytes)", handler.Filename, handler.Size)

	saveName := uuid.Must(uuid.NewRandom()).String()
	savePath := filepath.Join(saveDir, filepath.Base(saveName))

	dst, err := os.Create(savePath)
	if err != nil {
		log.Printf("Failed to create file: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		log.Printf("Failed to copy file content: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully saved file to: %s", savePath)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "File '%s' uploaded successfully.", handler.Filename)
}


func main() {
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		log.Fatalf("Failed to create save directory: %v", err)
	}
	log.Printf("Serving files to '%s' directory", saveDir)


	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/upload", handleUpload)

	// サーバー起動
	port := "8080"
	log.Printf("Starting server on http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}