package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	grpc_util "simple_db/grpc"
	"simple_db/operationlog"
	"strconv"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
)

func GetAllHandler(c *gin.Context) {
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
}

func GetObjectHandler(c *gin.Context) {

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
}
func PutObjectHandler(c1 *grpc_util.PutLogClient, c2 *grpc_util.PutLogClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data Data
		if err := c.ShouldBindJSON(&data); err != nil {
			c.Status(400)
			return
		}

		opId := len(operationlog.OperationLogs)

		opLog := operationlog.OperationLog{
			Id:     strconv.Itoa(opId),
			Key:    data.Key,
			Value:  data.Value,
			Status: 0,
		}

		operationlog.OperationLogs = append(operationlog.OperationLogs, &opLog)
		d, err := json.Marshal(data.Value)

		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// readyctx1, readycancle1 := context.WithTimeout(context.Background(), 5*time.Second)
		// defer readycancle1()
		// readyctx2, readycancle2 := context.WithTimeout(context.Background(), 5*time.Second)
		// defer readycancle2()

		// readiness1, err := (*c1).Ready(readyctx1, &grpc_util.M{})

		// if err != nil {
		// 	c.Status(http.StatusConflict)
		// 	log.Println(readiness1)
		// 	return
		// }

		// readiness2, err := (*c2).Ready(readyctx2, &grpc_util.M{})

		// if err != nil {
		// 	c.Status(http.StatusConflict)
		// 	log.Println(readiness2)
		// 	return
		// }

		ready := check_readiness(c1, c2)

		if !ready {
			c.Status(http.StatusConflict)
			return
		}

		// ctx1, cancle1 := context.WithTimeout(context.Background(), 5*time.Second)
		// defer cancle1()

		// ctx2, cancle2 := context.WithTimeout(context.Background(), 5*time.Second)
		// defer cancle2()

		// res1, err := (*c1).PutOperation(ctx1, &grpc_util.Operation{
		// 	Key:   data.Key,
		// 	Value: string(d),
		// })
		// if err != nil {
		// 	c.Status(http.StatusConflict)
		// 	return
		// }

		// if res1.GetStatus() != 0 {
		// 	c.Status(http.StatusConflict)
		// 	return
		// }

		// res2, err := (*c2).PutOperation(ctx2, &grpc_util.Operation{
		// 	Key:   data.Key,
		// 	Value: string(d),
		// })

		// if err != nil {
		// 	c.Status(http.StatusConflict)
		// 	return
		// }

		// if res2.GetStatus() != 0 {
		// 	c.Status(http.StatusConflict)
		// 	return
		// }

		replicated := replicate_operation(c1, c2, data.Key, string(d))

		if !replicated {
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
		opLog.Id = "1000"
		for _, item := range operationlog.OperationLogs {
			log.Println(item.Id)
		}
	}
}

func check_readiness(c1 *grpc_util.PutLogClient, c2 *grpc_util.PutLogClient) bool {
	readyctx1, readycancle1 := context.WithTimeout(context.Background(), 5*time.Second)
	defer readycancle1()
	readyctx2, readycancle2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer readycancle2()

	readiness1, err := (*c1).Ready(readyctx1, &grpc_util.M{})

	if err != nil {
		log.Println(readiness1)
		return false
	}

	readiness2, err := (*c2).Ready(readyctx2, &grpc_util.M{})

	if err != nil {
		log.Println(readiness2)
		return false
	}

	if readiness1.GetStatus() == 0 && readiness2.GetStatus() == 0 {
		return true
	}

	return false
}

func replicate_operation(c1 *grpc_util.PutLogClient, c2 *grpc_util.PutLogClient, key string, value string) bool {
	ctx1, cancle1 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancle1()

	ctx2, cancle2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancle2()

	res1, err := (*c1).PutOperation(ctx1, &grpc_util.Operation{
		Key:   key,
		Value: value,
	})
	if err != nil {
		// c.Status(http.StatusConflict)
		return false
	}

	if res1.GetStatus() != 0 {
		// c.Status(http.StatusConflict)
		return false
	}

	res2, err := (*c2).PutOperation(ctx2, &grpc_util.Operation{
		Key:   key,
		Value: value,
	})

	if err != nil {
		return false
	}

	if res2.GetStatus() != 0 {
		return false
	}
	return true
}
