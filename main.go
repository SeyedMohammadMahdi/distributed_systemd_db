//   ____ _                 ____  ____
//  / ___(_)_ __           |  _ \| __ )
// | |  _| | '_ \   _____  | | | |  _ \
// | |_| | | | | | |_____| | |_| | |_) |
//  \____|_|_| |_|         |____/|____/
// this project is going to implement a simple database using gin framework as the webserver backbone

package main

import (
	"context"
	"log"
	"net"
	"os"
	grpc_util "simple_db/grpc"
	"strconv"
	"strings"
	"sync"

	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Data struct {
	Key   string `json:"key" binding:"required"`
	Value any    `json:"value" binding:"required"`
}

var (
	address []string
	role    bool
	db      *badger.DB
	test    bool = false
)

func init() {
	addr, ok := os.LookupEnv("NODES")
	if !ok || addr == "" {
		log.Fatalln("could not initiate the NODES")
	}

	address = strings.Split(addr, ",")
	rl, ok := os.LookupEnv("ROLE")

	if !ok || rl == "" {
		log.Fatalln("could not get role from environment")
	}

	var err error
	t1, _ := os.LookupEnv("TEST")
	test, err = strconv.ParseBool(t1)
	if err != nil {
		test = false
	}
	role, err = strconv.ParseBool(rl)
	if err != nil {
		log.Fatalln(err)
	}

	db, err = badger.Open(badger.DefaultOptions("./data"))
	if err != nil {
		log.Fatalln(err)
	}

}

func main() {

	defer db.Close()

	var c1 grpc_util.PutLogClient
	var c2 grpc_util.PutLogClient
	var wg sync.WaitGroup = sync.WaitGroup{}

	if !role {
		wg.Add(1)
		go BackupServerLogRecServer(&wg)
	} else {
		conn1, err := grpc.NewClient(address[0], grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalln("could not connect to backup server grpc")
		}

		defer conn1.Close()

		c1 = grpc_util.NewPutLogClient(conn1)

		conn2, err := grpc.NewClient(address[1], grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalln("could not connect to backup server grpc")
		}

		defer conn2.Close()

		c2 = grpc_util.NewPutLogClient(conn2)
	}

	router := gin.Default()

	//check the role of the node if it is master or backup
	// if the node is master then we have
	if role {
		router.GET("/objects/:id", GetObjectHandler)
		router.GET("/objects", GetAllHandler)
		router.PUT("/objects", PutObjectHandler(&c1, &c2))
		router.Run()
	} else {
		router.GET("/objects/:id", GetObjectHandler)
		router.GET("/objects", GetAllHandler)
		router.Run()
	}

	if !role {
		wg.Wait()
	}
}

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
