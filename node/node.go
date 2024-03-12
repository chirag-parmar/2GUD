package main

import (
	"fmt"
//    "log"
	"net"
	"net/rpc"
	"math/rand"
	"github.com/schollz/peerdiscovery"
	// "time"
)

type Node struct {
	id string

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
	
	id_bytes := make([]byte, 32)
    rand.Read(id_bytes)
	n.id = fmt.Sprintf("%x", id_bytes)
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

func (n *Node) UploadRequest(args *UploadRequestArgs, reply *UploadRequestReply) (e error) {
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
	for hash, content := range args.files {
		if hash != ComputeHash(content) {
			reply.success = false
			reply.message = "computed hash does not match with provided hash"

			return nil
		}
		storeFile(n.id, "primary", hash, content)

		// add to temp table
		n.statusTable[hash] = 0

		reply.uploaded = append(reply.uploaded, hash)
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

	for _, hash := range args.hashes {
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
	for hash, content := range args.files {
		if hash != ComputeHash(content) {
			reply.success = false
			reply.message = "computed hash does not match with provided hash"

			return nil
		}
		storeFile(n.id, "replica", hash, content)

		// add to temp table
		n.statusTable[hash] = 2

		reply.replicated = append(reply.replicated, hash)
		reply.numReplicated++
		n.replicaBudget--
	}

	reply.success = true
	return nil
}

func (n *Node) ProposeReplication(args *ProposeReplicationArgs, reply *ProposeReplicationReply) (e error) {
	if n.replicaFor != "" {
		reply.replicaFor = n.replicaFor
		reply.granted = false

		return nil
	}

	n.replicaFor = args.proposer
	reply.replicaFor = n.replicaFor
	reply.granted = true

	return nil
}

func (n *Node) start() {
	n.init()

	rpc.Register(n)

	localIP := GetLocalIP()

	// Listen on a TCP address and port
	listener, err := net.Listen("tcp", localIP + ":8080")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer listener.Close()

	fmt.Println("RPC server listening on", listener.Addr())

	discoveries, _ := peerdiscovery.Discover(peerdiscovery.Settings{Limit: 5})
	for _, d := range discoveries {
		fmt.Printf("discovered '%s'\n", d.Address)
	}

	// Accept incoming connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		// Handle the connection in a separate goroutine using rpc.ServeConn
		go rpc.ServeConn(conn)
	}
	
}

func main() {
	n := Node{}
	n.start()
}
