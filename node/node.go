package main

import (
	"fmt"
	"flag"
	"net"
	"net/rpc"
	"net/http"
	"math/rand"
	"github.com/schollz/peerdiscovery"
	"time"
	"os"
    "os/signal"
    "syscall"
	"errors"
)

type Peer struct {
	address string
	isPrimary bool
	maritalStatus bool
	lastHeartBeat time.Time
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

	trees map[string]*MerkleTree
	treesStatus map[string]int
}

func (n *Node) init(address string, isPrimary bool, fileBudget int) {
	// Intitialize all maps
	n.fileBookings = make(map[string]int)
	n.fileStatusTable = make(map[string]int)
	n.peerTable = make(map[string]*Peer)
	n.discoveredAddresses = make(map[string]struct{})
	n.trees = make(map[string]*MerkleTree)
	n.treesStatus = make(map[string]int)

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

func (n *Node) HeartBeat(args *HeartBeatArgs, reply *HeartBeatReply) error {
	fmt.Printf("Got HeartBeat from %s -> address: %s, isPrimary: %t, maritalStatus: %t\n",
		args.Sender,
		args.Address,
		args.IsPrimary,
		args.MaritalStatus,
	)

	if _, ok := n.peerTable[args.Sender]; !ok {
		n.peerTable[args.Sender] = &Peer{
			address: args.Address,
			isPrimary: args.IsPrimary,
			maritalStatus: args.MaritalStatus,
			lastHeartBeat: time.Now(),
		}
	} else {
		n.peerTable[args.Sender].lastHeartBeat = time.Now()
		n.peerTable[args.Sender].maritalStatus = args.MaritalStatus
		n.peerTable[args.Sender].isPrimary = args.IsPrimary
	}

	reply.Receiver = n.id
	reply.IsPrimary = n.isPrimary
	reply.MaritalStatus = n.maritalStatus

	return nil
}

func (n *Node) UploadRequest(args *UploadRequestArgs, reply *UploadRequestReply) error {
	// check if storage is available
	if n.fileBudget < args.RequiredBudget{
		reply.Granted = false
		reply.Available = n.fileBudget

		return nil
	}

	n.fileBudget -= args.RequiredBudget
	n.fileBookings[args.RequesterID] = args.RequiredBudget

	reply.Granted = true
	reply.Available = n.fileBudget

	return nil
}

func (n *Node) UploadFiles(args *UploadFilesArgs, reply *UploadFilesReply) error {
	if _, ok := n.fileBookings[args.RequesterID]; !ok {
		return errors.New("no bookings made!")
	}

	if n.fileBookings[args.RequesterID] < len(args.Files) {
		return errors.New("fileBudget crossed!")
	}

	reply.NumUploads = 0
	for hash, content := range args.Files {
		if hash != ComputeHash(content) {
			return errors.New("computed hash does not match with provided hash!")
		}
		storeFile(n.id, hash, content)

		// add to temp table
		n.fileStatusTable[hash] = 0

		reply.Uploaded = append(reply.Uploaded, hash)
		reply.NumUploads++
		n.fileBookings[args.RequesterID]--
	}

	return nil
}

func (n *Node) CommitFiles(args *CommitFilesArgs, reply *CommitFilesReply) error {
	if _, ok := n.fileBookings[args.RequesterID]; !ok {
		return errors.New("Bookings not made!")
	}

	t := MerkleTree{}
	isFirst := true
	for _, hash := range args.Hashes {
		if _, ok := n.fileStatusTable[hash]; !ok {
			return errors.New("Hash was not uploaded!")		
		}
		
		n.fileStatusTable[hash] = 1
		if isFirst {
			t.Init(hash)
			isFirst = false
		} else {
			t.AddLeaf(hash)
		}
	}

	n.trees[t.root.hash] = &t
	n.treesStatus[t.root.hash] = 0
	reply.Merkle = t.root.hash
	reply.IndexMap = t.hashToIndex

	// reclaim file budget
	if n.fileBookings[args.RequesterID] > 0 {
		n.fileBudget += n.fileBookings[args.RequesterID];
	}

	// remove the booking entry made
	delete(n.fileBookings, args.RequesterID)

	return nil
}

func (n *Node) DownloadFile(args *DownloadFileArgs, reply *DownloadFileReply) error {
	if _, ok := n.trees[args.Merkle]; !ok {
		return errors.New("Merkle hash provided doesn't exist on this node")
	}

	err, content := readFile(n.id, n.trees[args.Merkle].indexToHash[args.Index])
	if err != nil {
		return err
	}

	// FIXME: skipping below check to make the designed system more meaningful for demo
	// ideally it is assumed that a corrupted file will also result in a corrupted merkle
	// tree. ex. a databse hack
	// if ComputeHash(content) != n.trees[args.Merkle].indexToHash[args.Index] {
	// 	return errors.New("File corrupted on server!")
	// }

	reply.Proof = n.trees[args.Merkle].GetProofByIndex(args.Index)
	reply.Content = content

	return nil
}

func (n *Node) ReplicateMerkle(args *ReplicateMerkleArgs, reply *ReplicateMerkleReply) error {
	// TODO:  this check should technically go in on every method
	// thereby selectively opening up RPC API based on roles
	if args.RequesterID != n.marriedTo {
		return errors.New("Not my partner!")
	}

	orderedHashes := make([]string, len(args.Hashes))
	for _, hash := range args.Hashes {
		// TODO: Analyze if this part is redundant or not. Should we really do another hash check?
		// if hash != ComputeHash(content) {
		// 	reply.success = false
		// 	reply.message = "computed hash does not match with provided hash"

		// 	return nil
		// }

		// add to temp table
		n.fileStatusTable[hash] = 1
		orderedHashes[args.IndexMap[hash]] = hash 
		n.fileBudget--
	}

	t := MerkleTree{}
	for i, mh := range orderedHashes {
		if i == 0 {
			t.Init(mh)
		} else {
			t.AddLeaf(mh)
		}
	}

	if t.root.hash != args.Merkle {
		return errors.New("The replication does not match the original!")
	}

	n.trees[t.root.hash] = &t
	n.treesStatus[t.root.hash] = 0
	reply.Success = true

	return nil
}

func (n *Node) Propose(args *ProposeArgs, reply *ProposeReply) error {
	if n.maritalStatus {
		return errors.New("Already married!")
	}

	n.marriedTo = args.Proposer
	n.maritalStatus = true
	reply.Granted = true

	return nil
}

func (n *Node) sendFirstHeartBeat(address string) error {
	args := HeartBeatArgs{
		Sender: n.id,
		Address: n.address,
		IsPrimary: n.isPrimary,
		MaritalStatus: n.maritalStatus,
	}
	
	var reply HeartBeatReply

	fmt.Printf("Sending first heart beat to: %s\n", address)
	
	if err := call(address, "Node.HeartBeat", &args, &reply); err != nil {
		fmt.Printf("Missed first heartbeat to %s\n", address)
		return err
	}

	fmt.Printf("%s-> id: %s, isPrimary: %t, maritalStatus: %t\n", 
		address, 
		reply.Receiver, 
		reply.IsPrimary, 
		reply.MaritalStatus,
	)

	n.peerTable[reply.Receiver] = &Peer{
		address: address,
		isPrimary: reply.IsPrimary,
		maritalStatus: reply.MaritalStatus,
		lastHeartBeat: time.Now(),
	}

	return nil
}

func (n *Node) checkHeartBeats() error {

	for id, peer := range n.peerTable {
		if time.Now().Sub(peer.lastHeartBeat) > time.Second {
			// send heartbeat
			args := HeartBeatArgs{
				Sender: n.id,
				Address: n.address,
				IsPrimary: n.isPrimary,
				MaritalStatus: n.maritalStatus,
			}

			var reply HeartBeatReply

			fmt.Printf("Sending heart beat to: %s\n", id)
			
			if err := call(peer.address, "Node.HeartBeat", &args, &reply); err != nil {
				fmt.Printf("Missed heartbeat to %s\n", id)
				return err
			}

			fmt.Printf("%s-> ip: %s, isPrimary: %t, maritalStatus: %t\n", 
				reply.Receiver, 
				peer.address, 
				reply.IsPrimary, 
				reply.MaritalStatus,
			)

			// update peer information
			peer.lastHeartBeat = time.Now()
			peer.isPrimary = reply.IsPrimary
			peer.maritalStatus = reply.MaritalStatus
		}
	}

	return nil
}

func (n *Node) discoverNewPeers(limit int) {
	//  TODO: Handle Error
	discoveries, _ := peerdiscovery.Discover(peerdiscovery.Settings{Limit: limit})
	for _, d := range discoveries {
		if _, ok := n.discoveredAddresses[d.Address]; !ok {
			fmt.Printf("Discovered new peer: %s\n", d.Address)
			n.discoveredAddresses[d.Address] = struct{}{}
			go n.sendFirstHeartBeat(d.Address)
		}
	}
}

func main() {
	isPrimary := flag.Bool("primary", false, "is this node a primary node")
	fileBudget := flag.Int("budget", 1000, "how many 1MB files can this node manage")
	flag.Parse()

	gracefulShutDown := make(chan os.Signal, 1)
	signal.Notify(gracefulShutDown, syscall.SIGINT, syscall.SIGTERM)

	// create a new instance
	n := new(Node)
	
	// initialize
	n.init(GetLocalIP(), *isPrimary, *fileBudget)
	
	rpc.Register(n)
	rpc.HandleHTTP()

	// Listen on a TCP address and port
	listener, err := net.Listen("tcp", n.address + ":8080")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer listener.Close()
	go http.Serve(listener, nil)

	fmt.Println("RPC server listening on", listener.Addr())

	// we could use the same quit channel but seperate control is reserverd for future improvements
	// discoveryTicker should technically tick at a much lower frequency.
	discoveryTicker := time.NewTicker(1 * time.Second)
	discoveryQuit := make(chan struct{})
	heartBeatTicker := time.NewTicker(1 * time.Second)
	heartBeatQuit := make(chan struct{})

	go func() {
		for {
			select {
				case <-discoveryTicker.C:
					n.discoverNewPeers(5)
				case <-discoveryQuit:
					discoveryTicker.Stop()
				case <-heartBeatTicker.C:
					n.checkHeartBeats()
				case <-heartBeatQuit:
					heartBeatTicker.Stop()
					return
			}
		}
	}()
	// defer (discoveryQuit <- struct{}{})
	// defer (heartBeatQuit <- struct{}{})

	<-gracefulShutDown
	fmt.Printf("Gracefully shutting down!")
}
