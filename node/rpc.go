package main

type HeartBeatArgs struct {
	Sender string
	Address string
	IsPrimary bool
	MaritalStatus bool
	MarriedTo string
}

type HeartBeatReply struct {
	Receiver string
	IsPrimary bool
	MaritalStatus bool
	MarriedTo string
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

type ReplicateMerkleArgs struct {
	RequesterID string
	IndexMap map[string]int
	Merkle string
}

type ReplicateMerkleReply struct {
	Success bool
}

type ProposeArgs struct {
	Proposer string
}

type ProposeReply struct {
	Granted bool
}