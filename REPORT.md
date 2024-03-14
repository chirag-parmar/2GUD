# Report - Programming Task

> Problem: Implement a file server that use merkle trees to provide proof to the client that the file is corrupted. Use networking and make the server design as ready for production as possible.

## Assumptions

1. The file size is less than or eual to 1MB
2. A standard server machine has 1TB of disk space leading to a approximate maximum of 1 million files.
3. Many servers of the above capacity are available
4. A single client will only upload less than 1 million files under one merkle tree.

## Server Design - 2Gether Until Death (2GUD *too good*)

The first and foremost thing we require to make a file server production ready is to be able to harness the storage capacity of multiple machines. At the same time we need to ensure a basic amount of data availability. To achieve both, Existing solutions use distributed networking protocols with provisions for data redundancy. 2Gether Until Death (2GUD) is one such protocol which is simple and intuitive. The main idea of the protocol is that every file server has one associated replica server with it. The two don't associate with any other servers (a.k.a nodes) until one of them dies, like a traditional marriage.

### Node discovery and heartbeats

As it is clear every node in the network is associated with only one other node. To achieve this first a node must have the capability of identifying other nodes in the network supporting 2GUD functionality. This is achieved using the `mDNS` protocol. The `"github.com/schollz/peerdiscovery` library is used as a ready solution for `mDNS` discovery. Once a node discoveres other nodes in the network it uses their IPv4 addresses for all further communications.

The first piece of communication that happens between nodes are heartbeats. A discovered node is sent a heartbeat as a RPC call (`Node.HeartBeat`) in which the sender node communicates its ID, IPv4 address and other information as arguments. A reply to this RPC call is considered a heartbeat from  receiver and it send the same information except the IPv4 address (since it was discovered). After the first heartbeats are exchange between nodes, they both continue to send heartbeats to each other periodically. Since even a reply is considered a heartbeat any one of the two can initiate the heartbeat RPC call. The period for the demo is set to one second. All gathered information from heartbeats is stored in a peer table by every node.

Apart from the ID and IPv4 address the nodes exchange status information. Since the protocol defines a marriage like scenario the status information includes every nodes marital status as defined below. Along with the marital status the information also includes if the server is a primary or a replica.

```go
type HeartBeatArgs struct {
	Sender string //sender ID
	Address string
	IsPrimary bool
	MaritalStatus bool
	MarriedTo string
}

type HeartBeatReply struct {
	Receiver string // receiver ID
	IsPrimary bool
	MaritalStatus bool
	MarriedTo string
}
```

### Proposals and Roles

Proposals can only be sent by primary nodes (using `Node.Propose`) and hence can only be accepted or rejected by replica nodes. The implementation must ensure that a primary node does not marry two or more replica nodes. If not ensured the network will have replica nodes that *assume* a marriage with a primary without any replication of data (classic case of a love-less extra-marital affair). This is regarded as an inefficiency in the netowrk to utilize storage space.

If a node misses to reply back with a heartbeat for an extended period of time it is assumed dead. A death of the replica node resets the marital status of the primary node which can then send proposals for marriage. A death of the primary node intiates a role switch of the replica to a primary in addition to the resetting of its marital status. After which it starts looking for another replica node in the network.

### Uploading Files

To facilitate chunked uploads of a large number of files the upload process is broken into three, booking, uploading and commiting. In an initial RPC call to the the client *books* a specified number of files. In subsequent RPC calls the client uploads cohorts of files to the server. In the final RPC call the client commits all these files to the server. The resepect RPC methods are `Node.UploadReuqest`, `Node.UploadFiles` and `Node.CommitFiles`. 

Since the booking happens based on number of files (and not the storage space required) every node has a file budget associated with it. With the assumptions made above the maximum file budget of a node is 1 million (1 million 1MB files in a 1TB node). A booking is rejected if the node does not have enough budget left to accomodate the client. 

```go
type UploadRequestArgs struct {
	RequiredBudget int
	RequesterID string
}

type UploadRequestReply struct {
	Granted bool
	Available int
}
```

Once the booking is made the client can choose to uploads all the files in cohorts of any size. Moreover, every call to the `UploadFiles` method is replied with hashes of the successfully uploaded files and the total number of uploads that happened in the call. Therefore, if any one cohort fails, the client can skip uploading the already uploaded files and reupload the other ones in that cohort. This is the main benefit of the broken down upload process. Every uploaded file decrements the booked file budget.

```go
type UploadFilesArgs struct {
	RequesterID string
	Files map[string]string
}

type UploadFilesReply struct {
	NumUploads int
	Uploaded []string
}
 
type CommitFilesArgs struct {
	Hashes []string
	RequesterID string
}

type CommitFilesReply struct {
	Merkle string
	IndexMap map[string]int
}
```

> :warning: The current demo implementation is not tested for edge cases and will fail in one particular edge case where a file is uploaded twice. In this case, it will eat off the booked budget but store the same file again. 

Once all files are uploaded (can be verified with the hashes returned to the `UploadFiles` request) they must be committed. Commiting the files initiates construction of a merkle tree. Once the tree has been constructed the node replies back with an index map, a map between the file hashes and their indexes in the merkle tree. The status is maintained using a table mapping the hashes of the file to its status. Uploaded is `0`, Commited is `1` and Replicated is `2`.

An implementation should routinely replicate files to its replica by reading the file status table and using the same RPC calls to book and upload files to the replica (note these RPC calls happen between the primary and the replica node whereas the above RPC calls happen between a client and a node). Files committed on a replica node do not initiate a tree construction (because it is assumed that the order of the leaf nodes a.k.a file hashes will not be or expected not to be preserved).

A special RPC method `Node.ReplicateMerkle` is used to replicate merkle trees between a primary and replica. In the call to this method the index map is shared to maintain the order. Similar to file statuses, statuses of trees are also maintinaed using a map between the merkle root hash of the tree and its status. While routinely replicating the files the implementation is also required to replicate the trees with it.

```go
type ReplicateMerkleArgs struct {
	RequesterID string
	IndexMap map[string]int
	Merkle string
}

type ReplicateMerkleReply struct {
	Success bool
}
```

If a replica node dies, its primary resets the status of files back to `1` (Commited but not replicated) and all of it s trees back to `0`(not replicated). In the case, a primary node dies, the file status and trees status on the replica node does not have to be changed because they were never further replicated.

### Downloading files

Files are downloaded by the client using the `Node.DownloadFile` RPC method. Files are references using merkle root hashes and their indexes in the tree. A call to this method initiates a proof construction from the merkle tree which is returned with the contents of the file. The client can then verify the contents of the file using the proof.

```go
type DownloadFileArgs struct {
	Merkle string
	Index int
}

type DownloadFileReply struct {
	Proof []string
	Content string
}
```

## Merkle Tree Implementation

### Tree Construction
Every node, including the root, utilizes a linked list like approach and is of the type below,
```go
type node struct {
	left *node
	right *node
	weight int
	hash string
}
```
while `left`, `right` and `hash` are intuitive and self explanatory, the `weight` is the sum of weights of the `left` and `right` nodes. Every leaf node has a weight of `1` and the all other weights are calculated as the tree is updated.

A tree is just a node (root) linked to various other nodes on the left and the right. A tree also maintians the map from indexes to hashes and hashes to indexes for efficient proof constructions. A tree is of the type below,
```go
type MerkleTree struct {
	root *node
	numLeaves int
	indexToHash map[int]string
	hashToIndex map[string]int
}
```
The first leaf is itself a root node. For every leaf added after that a recursive approach is taken to build the merkle tree using `addLeaf` method.

1. We start from the root node as the current node
2. if the current node is a leaf or its children have equal weight, the subtree/leaf at the current node is copied into a new node and shifted to the left and a new leaf node is added to the right. the weight of the newly added node is calculated as `weight of sidetree + weight of leaf`
3. if the current node is not a leaf or does not have equal weight children, the tree is walked down to the right and step 2 is called in recursion.
4. When the current recursion returns the weight of the parent node after walking down the right is also updated.

(refer `tree.go`)

> The implemented merkle tree only supports appending leaves one by one starting from the first to the last. This was done to initially support tree construction as files are uploaded. But due to scarcity of time and ease of implementation, the commit process was isolated from the upload process.

### Proof Construction

Proof construction follows the textbook method of converting the index of the leaf to a binary number and walking down the tree based on the bit (`0` left, `1` right). As it traverses down, it includes the adjacent node hash to the proof array. This implementation however also adds an instruction to every hash in the proof (a `0` or `1`) of whether the proof hash in the array should be concatenated ot the left or the right. So that verifiers don't have to repeat the process of converting the index to binary.

#### Some rough calculations

* one merkle node consumes = 8 + 8 + 8 + 32 = 56 bytes (roughly) (left pointer + right pointer + weight + hash)
* one entry in the hash table consumes about 40bytes (32 bytes hash + 8 bytes int)
* Total space 56*(2n - 1) + 40n + 40n= 192n - 56 ~ 192n
* if we have 1TB of space, we can store 1M files = 1M merkle leaf nodes => 192 * 10^6 bytes about 192MB for one server

## Drawbacks and Improvements

### Drawbacks

1. In the current implementation and protocol definition, a client independently contacts a primary node and has to remember the primary node it contacted. Moreover, in case of death, it will also have to find the associated replica node. This is not desirable.
2. A primary node replicates all its data to its replica node, the two must have equal storage capacities to avoid losing data or unused storage space.
3. Storage space is measured in number of files than actual storage space. This is a problem because 1 million 1KB files will only consume 1GB of disk space which is only 1% utilization of the server. Therefore a file budget type metric system is not ideal.
4. One clients request is only handled by one node pair. This limits the maximum files under one upload process to be 1 million. Moreover, this will leave small residue budget on servers that cannot be utilized by any clients.

### Improvement - Abstracting out the Merkling

Exsisting P2P tech like IPFS define a content addressable network, hence routing to specific nodes is not a problem when a piece of content is queried. But such a system beats the purpose of a merkled storage because querying will require the client to hold in memory the unique identifiers of individual files.

In 2GUD, we can define a special node that abstracts out the merkling from the storage nodes. All storage nodes independently form pairs but bookings for them are made by the merkler node. And, while uploads are made directly (by the client) on storage nodes, the commiting process is done by the merkler node at the request of the client. The merkler will keep track of all nodes in the network, their deaths and marriages.

The biggest benefit of using such a node would be that the merkler can distribute the storage load evenly to the entire network. This will help utilize residue budget on nodes that is too small for a single client. The added benefit is that the routing problem is solved because the merkler node maintains node information of the entire network anyways. 

### Other Improvments and Pending implementations in Code

1. The garbage cleaning of non commited files is still pending and would have been implemented if more time was available
2. The errors returned and handling of the errors
3. gRPC for streamed RPC calls for file uploading
4. Proper utilization of Go's concurrency tools, right now go routines are just fired without proper thought.