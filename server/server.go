package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"crypto/sha256"
	"os"
)

const MAX_FILE_SIZE = 1 << 20 // 1MB

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got /upload request\n")

	err := r.ParseMultipartForm(MAX_FILE_SIZE)
	if err != nil {
		http.Error(w, "Couldn't parse form data", http.StatusBadRequest)
		return
	}

	// Retrieve the file from the form data
	fileHash := r.FormValue("hash")
	f, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file content from request", http.StatusInternalServerError)
		return
	}
	fileContent, _ := ioutil.ReadAll(f)
	defer f.Close()

	//check hash of the file
	h := sha256.New()
	h.Write(fileContent)

  	if fileHash != fmt.Sprintf("%x", h.Sum(nil)) {
		fmt.Printf("%x", h.Sum(nil))
		http.Error(w, "an error occurred while computing the hash", http.StatusBadRequest)
		return
	}

	// Create the uploads folder if it doesn't
	// already exist
	err = os.MkdirAll("./uploads", os.ModePerm)
	if err != nil {
		http.Error(w, "Error creating directory for storing file", http.StatusInternalServerError)
		return
	}

	// Create a new file in the uploads directory
	err = os.WriteFile(fmt.Sprintf("./uploads/%s", fileHash), fileContent, 0644)
	if err != nil {
		http.Error(w, "Error writing to file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, fmt.Sprintf("{\"fileHash\": %s}", fileHash))
}

func main() {
	http.HandleFunc("/upload", uploadHandler)
	err := http.ListenAndServe(":3333", nil)
  	
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}