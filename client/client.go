package main

import (
	"fmt"
//    "log"
	"net"
	"net/rpc"
	"math/rand"
	"errors"
)

type Client struct {
	id string
}

func (c *Client) init() {
	id_bytes := make([]byte, 32)
    rand.Read(id_bytes)
	c.id = fmt.Sprintf("%x", id_bytes)
}

func (c *Client) BookServerBudget(nodeIP string) (e error) {

	args := UploadRequestArgs{1, c.id}
    var reply UploadRequestReply
	err := call(nodeIP, "Node.UploadRequest", &args, &reply)

	if err != nil {
		return err
	} else if !reply.granted {
		return errors.New(fmt.Sprintf("Only %d files can be uploaded", reply.available))
	}

	return nil
}

func (c *Client) uploadCohort(nodeIP string, files []string) (e error, uploaded []string) {

	args := UploadFilesArgs{c.id, files}
    var reply UploadFilesReply
	err := call(nodeIP, "Node.UploadFiles", &args, &reply)

	if err != nil {
		return err, nil
	} else if !reply.success || reply.numUploads != len(files) {
		return errors.New(reply.message), nil
	}

	return nil, reply.uploaded
}

func (c *Client) UploadFilesInCohort(nodeIP string, filePaths []string, cohortSize int) (e error) {
	

	return nil
}