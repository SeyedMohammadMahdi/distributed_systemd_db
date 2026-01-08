package operationlog

// status:
// 0 operation received
// 1 accepted
// 2 rejected
// 3 commit
// 4 abort
// if the operation is not in commit or uncommit status then the replicas negotiate with each other
type OperationLog struct {
	Id     string
	Key    string
	Value  any
	Status int
}

var OperationLogs []*OperationLog
