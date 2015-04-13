// db
package db

import (
	//"github.com/mattn/go-sqlite3"
	"./../geo"
	//"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Index struct {
	db     *bolt.DB
	bucket *bolt.Bucket
}

const bucket_name string = "Ways"

func (this *Index) LoadOrCreateOf(file string) (*Index, error) {

	file = this.getFilenameFromFile(file)

	log.Printf("for debug removing : %s", file)
	os.Remove(file)
	//	return this, errors.New("Not implemented")
	var err error
	this.db, err = bolt.Open(file, 0777, nil)
	if err != nil {
		log.Fatal(err)
	}
	err = this.db.Update(func(tx *bolt.Tx) error {
		this.bucket, err = tx.CreateBucket([]byte(bucket_name))
		if err != nil {
			log.Printf("create bucket: %s", err)
			return fmt.Errorf("create bucket: %s", err)
		}
		//Check if need ParseWay ?
		return errors.New("Database empty and need to be filled")
	})
	return this, err
}

func (this *Index) Add(wayid int64, bb geo.Bbox) error {
	return this.db.Update(func(tx *bolt.Tx) error {
		//b := tx.Bucket([]byte(bucket_name))
		b, err := tx.CreateBucketIfNotExists([]byte(bucket_name))
		if err != nil {
			log.Printf("create bucket: %s", err)
			return err
		}
		//TODO BSON
		//id := make([]byte, 8)
		//binary.PutVarint(id, wayid)
		data, err := json.Marshal(bb)
		err = b.Put([]byte(fmt.Sprint(wayid)), data)
		return err
	})
}

//For testing purpose

func (this *Index) Get(wayid int64) error {
	return this.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket_name))
		//id := make([]byte, 8)
		//binary.PutVarint(id, wayid)
		v := b.Get([]byte(fmt.Sprint(wayid)))
		log.Printf("The answer is: %s\n", v)
		return nil
	})
}
func (this *Index) getFilenameFromFile(file string) string {
	var extension = filepath.Ext(file)
	return strings.Join([]string{file[0 : len(file)-len(extension)], "idx"}, ".")
}
