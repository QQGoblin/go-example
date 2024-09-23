package main

import (
	"bytes"
	"flag"
	"fmt"
	bolt "go.etcd.io/bbolt"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"os"
)

func ReadOffline(db *bolt.DB, prefix string) (map[string]string, error) {

	bucket := "key"
	tmp := make(map[string]*mvccpb.KeyValue)
	result := make(map[string]string)
	prefixByte := []byte(prefix)

	if err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("got nil bucket for %s", bucket)
		}

		return b.ForEach(func(_, v []byte) error {

			var kv mvccpb.KeyValue
			if err := kv.Unmarshal(v); err != nil {
				return err
			}

			if !bytes.HasPrefix(kv.Key, prefixByte) {
				return nil
			}

			cur, isExit := tmp[string(kv.Key)]
			if !isExit || cur.Version < kv.Version {
				tmp[string(kv.Key)] = &kv
				result[string(kv.Key)] = string(kv.Value)
			}
			return nil
		})
	}); err != nil {
		return nil, err
	}
	return result, nil
}

var (
	prefix string
	path   string
)

func init() {
	flag.StringVar(&path, "path", "etcd-snapshot.db", "snapshot path")
	flag.StringVar(&prefix, "prefix", "", "etcd member name")
}

func main() {

	/*
		参考：
		- https://etcd.io/docs/v3.6/learning/persistent-storage-files/
		- https://etcd.io/docs/v3.6/learning/data_model/
	*/
	flag.Parse()
	if len(prefix) == 0 {
		fmt.Printf("please set etcd key prefix\n")
		os.Exit(-1)
	}

	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		os.Exit(-1)
	}
	defer db.Close()
	re, err := ReadOffline(db, prefix)
	if err != nil {
		fmt.Printf("read failed: %v\n", err)
		os.Exit(-1)
	}
	if len(re) == 0 {
		fmt.Printf("Key with prefix %s is not found\n", prefix)
	}

	for k, v := range re {
		fmt.Printf("Key=%s\n%s\n", k, v)
	}

}
