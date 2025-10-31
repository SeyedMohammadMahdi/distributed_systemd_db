//   ____ _                 ____  ____
//  / ___(_)_ __           |  _ \| __ )
// | |  _| | '_ \   _____  | | | |  _ \
// | |_| | | | | | |_____| | |_| | |_) |
//  \____|_|_| |_|         |____/|____/
// this project is going to implement a simple database using gin framework as the webserver backbone

package main

import (
	"encoding/json"
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
			respByte, err := item.ValueCopy(nil)

			if err != nil {
				return err
			}

			if err:= json.Unmarshal(respByte, &response); err != nil {
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
				if err := json.Unmarshal(valCopy ,&decode); err != nil {
					return err
				}

				result = append(result, gin.H{
					"key": string(key),
					"value": decode,
				})
			}
			return nil
		})

		if err != nil {
			c.Status(http.StatusBadRequest)
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

	router.Run()
}
