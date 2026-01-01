//   ____ _                 ____  ____
//  / ___(_)_ __           |  _ \| __ )
// | |  _| | '_ \   _____  | | | |  _ \
// | |_| | | | | | |_____| | |_| | |_) |
//  \____|_|_| |_|         |____/|____/
// this project is going to implement a simple database using gin framework as the webserver backbone

package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	grpc_util "simple_db/grpc"
	"strconv"
	"strings"

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
	if !role {
		go BackupServerLogRecServer()
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

	router.GET("/objects/:id", func(c *gin.Context) {

		id := c.Param("id")
		c.Status(200)
		var response any
		err := db.View(func(txn *badger.Txn) error {
			item, err := txn.Get([]byte(id))

			if err != nil {
				return err
			}

			// var err error
			respByte, err := item.ValueCopy(nil)

			if err != nil {
				return err
			}

			if err := json.Unmarshal(respByte, &response); err != nil {
				return err
			}

			return nil
		})

		if err == badger.ErrKeyNotFound {
			// c.Status(http.StatusNotFound)
			// c.JSON(http.StatusNotFound, gin.H{
			// 	"message": "not found",
			// 	"id": id,
			// })
			c.Status(http.StatusNotFound)
			return
		}
		c.JSON(http.StatusOK, response)
	})

	router.GET("/objects", func(c *gin.Context) {
		var result []gin.H
		err := db.View(func(txn *badger.Txn) error {
			it := txn.NewIterator(badger.DefaultIteratorOptions)
			defer it.Close()

			for it.Rewind(); it.Valid(); it.Next() {
				item := it.Item()
				key := item.Key()

				var valCopy []byte

				err := item.Value(func(val []byte) error {
					valCopy = append(valCopy, val...)
					return nil
				})

				if err != nil {
					return err
				}

				var decode any
				if err := json.Unmarshal(valCopy, &decode); err != nil {
					return err
				}

				result = append(result, gin.H{
					"key":   string(key),
					"value": decode,
				})
			}
			return nil
		})

		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		if len(result) == 0 {
			c.Status(http.StatusNotFound)
			return
		}

		c.JSON(http.StatusOK, result)
	})

	router.PUT("/objects", func(c *gin.Context) {
		var data Data
		if err := c.ShouldBindJSON(&data); err != nil {
			c.Status(400)
			return
		}

		d, err := json.Marshal(data.Value)

		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		res1, err := c1.PutOperation(context.Background(), &grpc_util.Operation{
			Key:   data.Key,
			Value: string(d),
		})
		if err != nil {
			c.Status(http.StatusConflict)
			return
		}

		if res1.GetStatus() != 0 {
			c.Status(http.StatusConflict)
			return
		}

		res2, err := c2.PutOperation(context.Background(), &grpc_util.Operation{
			Key:   data.Key,
			Value: string(d),
		})

		if err != nil {
			c.Status(http.StatusConflict)
			return
		}

		if res2.GetStatus() != 0 {
			c.Status(http.StatusConflict)
			return
		}

		err = db.Update(func(txn *badger.Txn) error {

			err := txn.Set([]byte(data.Key), d)
			if err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		c.Status(http.StatusOK)

	})

	//check the role of the node if it is master or backup
	// if the node is master then we have

	router.Run()
}

type server struct {
	grpc_util.UnimplementedPutLogServer
}

func (s *server) PutOperation(ctx context.Context, in *grpc_util.Operation) (*grpc_util.Status, error) {
	// to test if the setup is working in syncronous way uncomment the following line
	// what you should expect is that until the backup server do not write the data in the data base and respond with success the master won't write it
	// return &grpc_util.Status{Status: 1}, nil
	d, err := json.Marshal(in.GetValue())
	if err != nil {
		return &grpc_util.Status{Status: 1}, nil
	}
	err = db.Update(func(txn *badger.Txn) error {
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

func BackupServerLogRecServer() {
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
