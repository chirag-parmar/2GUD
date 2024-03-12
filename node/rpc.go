package main

type HeartBeatArgs struct {
	sender string
	address string
	isPrimary bool
	maritalStatus bool
}

type HeartBeatReply struct {
	receiver string
	isPrimary bool
	maritalStatus bool
}

type UploadRequestArgs struct {
	requiredBudget int
	requesterID string
}

type UploadRequestReply struct {
	granted bool
	available int
}

type UploadFilesArgs struct {
	requesterID string
	files map[string]string
}

type UploadFilesReply struct {
	success bool
	numUploads int
	message string
	uploaded []string
}

type CommitFilesArgs struct {
	hashes []string
	requesterID string
}

type CommitFilesReply struct {
	success bool
	message string
	merkle string
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