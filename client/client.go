package main

import (
	"fmt"
	"os"
	"math/rand"
	"errors"
	"strconv"
)

type Client struct {
	id string
}

func (c *Client) init() {
	id_bytes := make([]byte, 32)
    rand.Read(id_bytes)
	c.id = fmt.Sprintf("%x", id_bytes)
}

func (c *Client) BookServerBudget(address string, budget int) error {

	args := UploadRequestArgs{
		RequiredBudget: budget, 
		RequesterID: c.id,
	}

    var reply UploadRequestReply
	
	err := call(address, "Node.UploadRequest", &args, &reply)

	if err != nil {
		return err
	} else if !reply.Granted {
		return errors.New(fmt.Sprintf("Only %d files can be uploaded", reply.Available))
	}

	return nil
}

func (c *Client) uploadCohort(address string, files map[string]string) (err error, uploaded []string) {

	args := UploadFilesArgs{
		RequesterID: c.id, 
		Files: files,
	}

    var reply UploadFilesReply

	err = call(address, "Node.UploadFiles", &args, &reply)

	if err != nil {
		return err, nil
	} else if reply.NumUploads != len(files) {
		return errors.New("Few files failed to upload"), nil
	}

	return nil, reply.Uploaded
}

func (c *Client) UploadFiles(address string, filePaths []string, cohortSize int) (err error, uploadedHashes []string) {

	files := make(map[string]string)

	for i, filePath := range filePaths {
		dat, err := os.ReadFile(filePath)

		if err != nil {
			return err, uploadedHashes
		}

		files[ComputeHash(string(dat))] = string(dat)

		if (i+1)%cohortSize == 0 {
			err, uploaded := c.uploadCohort(address, files)

			if err != nil {
				return err, uploadedHashes
			}

			//check if all uploaded hashes exist in the files cohort
			for _, h := range uploaded {
				if _, ok := files[h]; !ok {
					return errors.New("uploaded file hash doesn't match"), uploadedHashes
				}
			}

			uploadedHashes = append(uploadedHashes, uploaded...)
			files = nil
			files = make(map[string]string)
		}
	}

	// upload the last cohort
	if len(filePaths) < len(uploadedHashes) && files != nil {
		err, uploaded := c.uploadCohort(address, files)
		if err != nil {
			return err, uploadedHashes
		}

		//check if all uploaded hashes exist in the files cohort
		for _, h := range uploaded {
			if _, ok := files[h]; !ok {
				return errors.New("uploaded file hash doesn't match"), uploadedHashes
			}
		}

		uploadedHashes = append(uploadedHashes, uploaded...)
		files = nil
	}

	return nil, uploadedHashes
}

func (c *Client) CommitFiles(address string, uploadedHashes []string) (err error) {
	
	commitArgs := CommitFilesArgs{
		Hashes: uploadedHashes,
		RequesterID: c.id,
	}
	var commitReply CommitFilesReply

	// Commit files on server
	err = call(address, "Node.CommitFiles", &commitArgs, &commitReply)
	if err != nil {
		return err
	}

	// create a merkle tree
	t := MerkleTree{}
	orderedHashes := make([]string, len(uploadedHashes))

	for uh, index := range commitReply.IndexMap {
		orderedHashes[index] = uh
	}

	t.Init(orderedHashes[0])

	for k := 1; k < len(orderedHashes); k++ {
		t.AddLeaf(orderedHashes[k])
	}

	if t.root.hash != commitReply.Merkle {
		return errors.New("Merkle root doesn't match")
	}

	return nil
}

func main() {
	// FIXME: the entire code block below is hack, it is dirty and hardcoded
	// this must be done dynamixally by reading directory as command line arguments
	basePath := "uploadables/"
	fileBatch1 := make([]string, 300)
	fileBatch2 := make([]string, 300)
	fileBatch3 := make([]string, 400)

	for i := 0; i < 1000; i++ {
		if i < 300 {
			fileBatch1[i] = basePath + strconv.Itoa(i) + ".txt"
		} else if i < 600 {
			fileBatch2[i-300] = basePath + strconv.Itoa(i) + ".txt"
		} else {
			fileBatch3[i-600] = basePath + strconv.Itoa(i) + ".txt"
		}
	}

	// addresses := []string{"172.10.0.2", "172.10.0.3", "172.10.0.4"}
	addresses := []string{"127.0.0.1", "127.0.0.1", "127.0.0.1"}
	// Dirty code ends here

	client := new(Client)
	client.init()

	err := client.BookServerBudget(addresses[0], 300)
	if err != nil {
		panic(err)
	}

	err, uploadedHashes1 := client.UploadFiles(addresses[0], fileBatch1, 50)
	if err != nil {
		panic(err)
	}

	err = client.CommitFiles(addresses[0], uploadedHashes1)
	if err != nil {
		panic(err)
	}

	err = client.BookServerBudget(addresses[1], 300)
	if err != nil {
		panic(err)
	}

	err, uploadedHashes2 := client.UploadFiles(addresses[1], fileBatch2, 50)
	if err != nil {
		panic(err)
	}

	err = client.CommitFiles(addresses[1], uploadedHashes2)
	if err != nil {
		panic(err)
	}

	err = client.BookServerBudget(addresses[2], 400)
	if err != nil {
		panic(err)
	}

	err, uploadedHashes3 := client.UploadFiles(addresses[2], fileBatch3, 50)
	if err != nil {
		panic(err)
	}

	err = client.CommitFiles(addresses[2], uploadedHashes3)
	if err != nil {
		panic(err)
	}

}