//   ____ _                 ____  ____
//  / ___(_)_ __           |  _ \| __ )
// | |  _| | '_ \   _____  | | | |  _ \
// | |_| | | | | | |_____| | |_| | |_) |
//  \____|_|_| |_|         |____/|____/
// this project is going to implement a simple database using gin framework as the webserver backbone

package main

import (
	"log"
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
