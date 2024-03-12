package main

import (
	"fmt"
	"flag"
	"net"
	"net/rpc"
	"math/rand"
	"github.com/schollz/peerdiscovery"
	"time"
)

type Peer struct {
	address string
	isPrimary bool
	maritalStatus bool
	lastHeartBeat time.Duration
}

type Node struct {
	id string
	address string

	isPrimary bool
	maritalStatus bool
	marriedTo string

	discoveredAddresses map[string]struct{}
	peerTable map[string]*Peer

	fileBudget int

	fileBookings map[string]int
	fileStatusTable map[string]int
}

func (n *Node) init(address string, isPrimary bool, fileBudget int) {
	// Intitialize all maps
	n.fileBookings = make(map[string]int)
	n.fileStatusTable = make(map[string]int)
	n.peerTable = make(map[string]*Peer)
	n.discoveredAddresses = make(map[string]struct{})

	// set passed arguments
	n.address = address
	n.isPrimary = isPrimary
	n.fileBudget = fileBudget

	// set id
	id_bytes := make([]byte, 32)
    rand.Read(id_bytes)
	n.id = fmt.Sprintf("%x", id_bytes)

	// set everything else to default
	n.maritalStatus = false
	n.marriedTo = ""
}

func (n *Node) HeartBeat(args *HeartBeatArgs, reply *HeartBeatReply) {
	fmt.Printf("Got HeartBeat from %s -> address: %s, isPrimary: %t, maritalStatus: %t",
		args.sender,
		args.address,
		args.isPrimary,
		args.maritalStatus
	)

	if _, ok := n.peerTable[args.sender]; !ok {
		n.peerTable[args.sender] = &Peer{
			address: args.address
			isPrimary: args.isPrimary
			maritalStatus: args.maritalStatus
			lastHeartBeat: time.Now()
		}
	} else {
		n.peerTable[args.sender].lastHeartBeat = time.Now()
		n.peerTable[args.sender].maritalStatus = args.maritalStatus
		n.peerTable[args.sender].isPrimary = args.isPrimary
	}

	reply.receiver = n.id
	reply.isPrimary = n.isPrimary
	reply.maritalStatus = n.maritalStatus
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

func (n *Node) sendFirstHeartBeat(address string) (e error) {
	args := HeartBeatArgs{
		sender: n.id,
		address: n.address
		isPrimary: n.isPrimary
		maritalStatus: n.maritalStatus
	}
	
	var reply HeartBeatReply

	fmt.Printf("Sending first heart beat to: %s\n", address)
	
	if err := call(address, "Node.HeartBeat", &args, &reply); e != nil {
		return err
	}

	fmt.Printf("%s-> id: %s, isPrimary: %t, maritalStatus: %t\n", 
		address, 
		reply.receiver, 
		reply.isPrimary, 
		reply.maritalStatus,
	)

	n.peerTable[reply.receiver] = &Peer{
		address: address,
		isPrimary: reply.isPrimary,
		maritalStatus: reply.maritalStatus,
		lastHeartBeat: time.Now(),
	}

	return nil
}

func (n *Node) checkHeartBeats() (e error) {

	for id, peer := range n.peerTable {
		if (time.Now() - peer.lastHeartBeat) > time.Second {
			// send heartbeat
			args := HeartBeatArgs{
				sender: n.id,
				address: n.address
				isPrimary: n.isPrimary
				maritalStatus: n.maritalStatus
			}

			var reply HeartBeatReply

			fmt.Printf("Sending heart beat to: %s\n", id)
			
			if err := call(peer.address, "Node.HeartBeat", &args, &reply); e != nil {
				fmt.Printf("Missed heartbeat to %s\n", id)
				return err
			}

			fmt.Printf("%s-> ip: %s, isPrimary: %t, maritalStatus: %t\n", 
				reply.receiver, 
				peer.address, 
				reply.isPrimary, 
				reply.maritalStatus,
			)

			// update peer information
			peer.lastHeartBeat = time.Now()
			peer.isPrimary = reply.isPrimary
			peer.maritalStatus = reply.maritalStatus
		}
	}

	return nil
}

func (n *Node) discoverNewPeers(limit int) {
	discoveries, _ := peerdiscovery.Discover(peerdiscovery.Settings{Limit: limit})
	for _, d := range discoveries {
		if _, ok := n.discoveredAddresses[address]; !ok {
			fmt.Printf("Discovered new peer: %s", address)
			n.discoveredAddresses[address] = struct{}{}
			go n.sendFirstHeartBeat(address)
		}
	}
}

func (n *Node) start() {
	rpc.Register(n)

	// Listen on a TCP address and port
	listener, err := net.Listen("tcp", n.address + ":8080")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer listener.Close()

	fmt.Println("RPC server listening on", listener.Addr())

	// we could use the same quit channel but seperate control is reserverd for future improvements
	discoveryTicker := time.NewTicker(30 * time.Second)
	discoveryQuit := make(chan struct{})
	heartBeatTicker := time.NewTicker(1 * time.Second)
	heartBeatQuit := make(chan struct{})

	go func() {
		for {
			select {
				case <- discoveryTicker.C:
					n.discoverNewPeers(5)
				case <- discoveryQuit:
					discoveryTicker.Stop()
				case <- heartBeatTicker.C:
					n.checkHeartBeats()
				case <- heartBeatQuit:
					heartBeatTicker.Stop()
					return
			}
		}
	}()
	defer discoveryQuit<-struct{}{}
	defer heartBeatQuit<-struct{}{}

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
	isPrimary := flag.Bool("primary", false, "is this node a primary node")
	fileBudget := flag.Int("budget", false, "how many 1MB files can this node manage")

	// create a new instance
	n := Node{}
	
	// initialize
	n.init(GetLocalIP(), isPrimary, fileBudget)
	
	// start server
	n.start()
}
