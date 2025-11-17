package main

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
)

//go:embed index.gohtml
var content embed.FS

var tmpl = template.Must(template.ParseFS(content, "index.gohtml"))

func handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	err := tmpl.ExecuteTemplate(w, "index.gohtml", nil)
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

	// Just simulating saving the file without actually writing it to disk
	dst, err := os.Create("/dev/null")
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

	log.Printf("Successfully saved file to: %s", "/dev/null")

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "File '%s' uploaded successfully.", handler.Filename)
}


func main() {
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/upload", handleUpload)

	port := "8080"
	log.Printf("Starting server on http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}