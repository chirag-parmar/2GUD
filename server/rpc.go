type HeartBeatArgs struct {
	heartBeat bool
}

type HeartBeatReply struct {
	heartBeat bool
}

type UploadRequestArgs struct {
	requiredBudget int
	requesterID string
}

type UploadReqeustReply struct {
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
	numUploads int
	message string
	uploaded []string
}

type ProposeReplicationArgs struct {
	proposer string
}

type ProposeReplicationReply struct {
	replicaFor string
	granted bool
}