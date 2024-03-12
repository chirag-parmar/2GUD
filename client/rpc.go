package main

type HeartBeatArgs struct {
	Sender string
	Address string
	IsPrimary bool
	MaritalStatus bool
}

type HeartBeatReply struct {
	Receiver string
	IsPrimary bool
	MaritalStatus bool
}

type UploadRequestArgs struct {
	RequiredBudget int
	RequesterID string
}

type UploadRequestReply struct {
	Granted bool
	Available int
}

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

type DownloadFileArgs struct {
	Merkle string
	Index int
}

type DownloadFileReply struct {
	Proof []string
	Content string
}

type ReplicateArgs struct {
	requesterID string
	files map[string]string
}

type ReplicateReply struct {
	success bool
	numReplicated int
	message string
	replicated []string
}

type ProposeReplicationArgs struct {
	proposer string
}

type ProposeReplicationReply struct {
	replicaFor string
	granted bool
}