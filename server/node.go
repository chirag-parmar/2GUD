package main

import (
   "fmt"
   "log"
   "net"
   "net/rpc"
   "errors"
)

type Node struct {
	replicaFor string
	missedHeartBeats int

	fileBudget int
	replicaBudget int

	bookings map[string]int
	statusTable map[string]int
}

func (n *Node) init() {
	n.bookings = make(map[string]int)
	n.statusTable = make(map[string]int)

	n.missedHeartBeats = 0
	n.replicaFor = ""

	// TODO: can go bigger, assuming 1MB max limit on files this is a maximum of 1GB for primary and 1GB for replica
	n.fileBudget = 1000
	n.replicaBudget = 1000
	
	
}

func (n *Node) ReportHeartBeat(args *HeartBeatArgs, reply *HeartBeatReply) (e error) {
	// TODO: should we check actual caller instead of args object? requires analysis of trust model
	if args.proposer == n.replicaFor {
		reply.heartBeat = true
		return nil
	}

	reply.heartBeat = false
	return nil
}

func (n *Node) UploadRequest(args *UploadRequestArgs, reply *UploadFileReply) (e error) {
	// check if storage is available
	if n.fileBudget < args.requiredBudget{
		reply.granted = false
		reply.available = n.fileBudget

		return nil
	}

	n.fileBudget -= args.requiredBudget
	n.bookings[args.requesterID] = args.requiredBudget

	reply.granted = true
	reply.available = n.fileBudget

	return nil
}

func (n *Node) UploadFiles(args *UploadFilesArgs, reply *UploadFilesReply) (e error) {
	if _, ok := n.bookings[args.requesterID]; !ok {
		reply.numUploads = 0
		reply.success = false
		reply.uploaded = nil
		reply.message = "no bookings made!"

		return nil
	}

	if n.bookings[args.requesterID] < len(args.files) {
		reply.numUploads = 0
		reply.success = false
		reply.uploaded = nil
		reply.message = "fileBudget crossed!"

		return nil
	}

	reply.numUploads = 0
	reply.uploaded = []
	for hash, content in args.files {
		if hash != computeHash(content) {
			reply.success = false
			reply.message = "computed hash does not match with provided hash"

			return nil
		}
		storeFile("primary", hash, content)

		// add to temp table
		n.statusTable[hash] = 0

		reply.uploaded.append(hash)
		reply.numUploads++
		n.bookings[args.requesterID]--
	}

	reply.success = true
	return nil
}

func (n *Node) CommitFiles(args *CommitFilesArgs, reply *CommitFilesReply) (e error) {
	if _, ok := n.bookings[args.requesterID]; !ok {
		reply.success = false
		reply.message = "bookings not made"
		reply.merkle = ""

		return nil
	}

	for hash in args.hashes {
		n.statusTable[hash] = 1
	}

	// TODO: create merkle tree here
	reply.merkle = ""
	reply.success = true

	// reclaim file budget
	if n.bookings[args.requesterID] > 0 {
		n.fileBudget += n.bookings[args.requesterID];
		delete(n.bookings, args.requesterID)
	}

	return nil
}

func (n *Node) Replicate(args *ReplicateArgs, reply *ReplicateReply) (e error) {
	if args.requesterID != n.replicaFor {
		reply.success = false
		reply.message = "Not your replica"
		reply.numReplicated = 0
		reply.replicated = nil

		return nil
	}

	reply.numReplicated = 0
	reply.replicated = []
	for hash, content in args.files {
		if hash != computeHash(content) {
			reply.success = false
			reply.message = "computed hash does not match with provided hash"

			return nil
		}
		storeFile("primary", hash, content)

		// add to temp table
		n.statusTable[hash] = 0

		reply.replicated.append(hash)
		reply.numReplicated++
		n.replicaBudget--
	}

	reply.success = true
	return nil
}

func (n *Node) ProposeReplication(args *ProposeReplicationArgs, reply *ProposeReplicationReply) (e error) {
	if n.replicaFor != "" {
		reply.replicaFor = n.replicaFor
		reply.answer = false

		return nil
	}

	n.replicaFor = args.proposer
	reply.replicaFor = n.replicaFor
	reply.answer = true

	return nil
}

//
// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
//
func call(rpcname string, args interface{}, reply interface{}) (e error) {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := coordinatorSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}

func (n *Node) start() {
	rpc.Register(n)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}


