//   ____ ___ _   _       ____  ____
//  / ___|_ _| \ | |     |  _ \| __ )
// | |  _ | ||  \| |_____| | | |  _ \
// | |_| || || |\  |_____| |_| | |_) |
//  \____|___|_| \_|     |____/|____/

// this project is going to implement a simple database using gin framework as the webserver backbone
package main

import (
	// "bytes"
	// "encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
)

type Data struct {
	Key   string          `json:"key" binding:"required"`
	Value any `json:"value" binding:"required"`
}

func main() {

	db, err := badger.Open(badger.DefaultOptions("./data"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

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
			response, err = item.ValueCopy(nil)

			if err != nil {
				return err
			}

			return nil
		})

		if err == badger.ErrKeyNotFound {
			// c.Status(http.StatusNotFound)
			c.JSON(http.StatusNotFound, gin.H{
				"message": "not found",
				"id": id,
			})
			return
		}
		c.JSON(http.StatusOK, response)
	})

	
	router.PUT("/objects", func(c *gin.Context) {
		var data Data
		if err := c.ShouldBindJSON(&data); err != nil {
			c.Status(400)
			return
		}
		fmt.Println(data.Value)

		// var pretty bytes.Buffer
		// if err := json.Indent(&pretty, data.Value, "", "  "); err != nil {
		// 	// fallback if raw JSON is invalid
		// 	fmt.Printf("Key: %s, Value (raw): %s\n", data.Key, string(data.Value))
		// } else {
		// 	fmt.Printf("Key: %s\nValue:\n%s\n\n", data.Key, pretty.String())
		// }

	})

	router.Run()
}
