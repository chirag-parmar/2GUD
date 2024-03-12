package main

import (
	"errors"
	"fmt"
	"crypto/sha256"
	"os"
	"net/rpc"
)

func ComputeHash(content string) string {
	//check hash of the file
	h := sha256.New()
	h.Write([]byte(content))

  	return fmt.Sprintf("%x", h.Sum(nil))
}

func storeFile(node_id string, databank string, hash string, content string) (err error) {
	// Create the uploads folder if it doesn't already exist
	err = os.MkdirAll(fmt.Sprintf("./%s", node_id), os.ModePerm)
	if err != nil {
		return errors.New("Error creating directory for storing file")
	}

	if databank != "replica" {
		databank = "primary"
	}

	// Create the uploads folder if it doesn't already exist
	err = os.MkdirAll(fmt.Sprintf("./%s/%s", node_id, databank), os.ModePerm)
	if err != nil {
		return errors.New("Error creating directory for storing file")
	}

	// Create a new file in the uploads directory
	err = os.WriteFile(fmt.Sprintf("./%s/%s/%s", node_id, databank, hash), []byte(content), 0644)
	if err != nil {
		return errors.New("Error writing to file")
	}

	return nil
}

// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
//
func call(rpcname string, args interface{}, reply interface{}) (e error) {
	c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":8080")
	if err != nil {
		return err
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return nil
	}

	return err
}