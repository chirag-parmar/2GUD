import (
	"errors"
	"fmt"
	"crypto/sha256"
	"os"
)

func ComputeHash(content []byte) []byte {
	//check hash of the file
	h := sha256.New()
	h.Write(content)

  	return fmt.Sprintf("%x", h.Sum(nil))
}

func storeFile(databank string, hash string, content []byte) (e error) {
	// Create the uploads folder if it doesn't already exist
	err = os.MkdirAll(fmt.Sprintf("./%s", NODE_ID), os.ModePerm)
	if err != nil {
		return error.New("Error creating directory for storing file")
	}

	if databank != "replica" {
		databank = "primary"
	}

	// Create the uploads folder if it doesn't already exist
	err = os.MkdirAll(fmt.Sprintf("./%s/%s", NODE_ID, databank), os.ModePerm)
	if err != nil {
		return error.New("Error creating directory for storing file")
	}

	// Create a new file in the uploads directory
	err = os.WriteFile(fmt.Sprintf("./%s/%s/%s", NODE_ID, databank, hash), content, 0644)
	if err != nil {
		return error.New("Error writing to file")
	}
}