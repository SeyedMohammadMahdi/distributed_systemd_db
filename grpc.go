package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	grpc_util "simple_db/grpc"
	"simple_db/operationlog"
	"sync"

	"github.com/dgraph-io/badger/v4"
	"google.golang.org/grpc"
)

type server struct {
	grpc_util.UnimplementedPutLogServer
}

func (s *server) PutOperation(ctx context.Context, in *grpc_util.Operation) (*grpc_util.Status, error) {
	// to test if the setup is working in syncronous way uncomment the following line
	// what you should expect is that until the backup server do not write the data in the data base and respond with success the master won't write it
	// return &grpc_util.Status{Status: 1}, nil
	var d []byte = []byte(in.GetValue())
	var op *operationlog.OperationLog = nil
	var id = in.GetId()
	for _, item := range operationlog.OperationLogs {
		if item.Id == id {
			op = item
		}
	}
	op.Key = in.GetKey()
	if err := json.Unmarshal(d, &op.Value); err != nil {
		log.Println(err)
		return &grpc_util.Status{
			Status: 1,
		}, nil
	}
	
	err := db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(in.GetKey()), d)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return &grpc_util.Status{Status: 1}, nil
	}
	log.Println("done")
	op.Status = 3
	return &grpc_util.Status{Status: 0}, nil
}

func (s *server) Ready(ctx context.Context, in *grpc_util.Id) (*grpc_util.Status, error) {
	log.Println("replica_1 is ready...")
	opLog := operationlog.OperationLog{
		Id:     in.GetId(),
		Status: 1,
	}
	// mutex lock needed
	operationlog.OperationLogs = append(operationlog.OperationLogs, &opLog)
	return &grpc_util.Status{
		Status: 0,
	}, nil
}

func (s *server) Abort(ctx context.Context, in *grpc_util.Id) (*grpc_util.Status, error) {

	var op *operationlog.OperationLog = nil
	id := in.GetId()
	for _, item := range operationlog.OperationLogs {
		if item.Id == id {
			op = item
			break
		}
	}

	if op == nil || op.Status == 4 { //if operation is already aborted no need to do it again
		return &grpc_util.Status{
			Status: 0,
		}, nil
	}

	err := db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(op.Key))
	})

	if err != nil {
		return &grpc_util.Status{
			Status: 1,
		}, nil
	}
	op.Status = 4
	return &grpc_util.Status{
		Status: 0,
	}, nil
}

func BackupServerLogRecServer(wg *sync.WaitGroup) {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen to 50051 for grpc: %v", err)
	}
	defer lis.Close()
	s := grpc.NewServer()
	grpc_util.RegisterPutLogServer(s, &server{})
	log.Println("listening for grpc requests...")
	if err := s.Serve(lis); err != nil {
		log.Fatal("failed to serve")
	}
}
