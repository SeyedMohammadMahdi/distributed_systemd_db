package main

import (
	"context"
	"log"
	"net"
	grpc_util "simple_db/grpc"
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
	return &grpc_util.Status{Status: 0}, nil
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
