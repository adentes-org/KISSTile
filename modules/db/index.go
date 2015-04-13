// db
package db

/*
//TODO try leveldb or ameliore with block and compression like pbf with miinimal bbox


import (
	//"github.com/mattn/go-sqlite3"
	"./../geo"
	"encoding/binary"
	"errors"
	//"gopkg.in/mgo.v2/bson"
	//	"fmt"
	//	"github.com/boltdb/bolt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
)

type Index struct {
	//	db     *bolt.DB
	//	bucket *bolt.Bucket
	db *os.File
}

const bucket_name string = "Ways"

func (this *Index) LoadOrCreateOf(file string) (*Index, error) {

	file = this.getFilenameFromFile(file)

	log.Printf("for debug removing : %s", file)
	os.Remove(file)
	//	return this, errors.New("Not implemented")
	var err error
	//TODO re-get last states
	this.db, err = os.OpenFile(file, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0777)
	//this.db, err = os.Create(file)
	if err != nil {
		log.Printf("%v", err)
	}

	return this, errors.New("Database empty and need to be filled")
}

func (this *Index) Add(wayid int64, bb geo.Bbox) error {
	//TODO optimize for batch write
	//TODO use  this.tmp and write by bloclk useless ?
	data := Int64bytes(wayid)
	//tmp := Float64bytes(bb[0].Lat)
	data = append(data, Float64bytes(bb[0].Lat)[:]...)
	data = append(data, Float64bytes(bb[0].Lon)[:]...)
	data = append(data, Float64bytes(bb[1].Lat)[:]...)
	data = append(data, Float64bytes(bb[1].Lon)[:]...)
	//, Float64bytes(bb[0].Lon), Float64bytes(bb[1].Lat), Float64bytes(bb[1].Lon))
	this.db.Write(data)
	return nil
}

//For testing purpose

func (this *Index) Get(wayid int64) (geo.Bbox, error) {

	return geo.Bbox{}, nil
}
func (this *Index) getFilenameFromFile(file string) string {
	var extension = filepath.Ext(file)
	return strings.Join([]string{file[0 : len(file)-len(extension)], "idx"}, ".")
}
func Float64frombytes(bytes []byte) float64 {
	bits := binary.LittleEndian.Uint64(bytes)
	float := math.Float64frombits(bits)
	return float
}

func Float64bytes(float float64) []byte {
	bits := math.Float64bits(float)
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, bits)
	return bytes
}

func Int64bytes(i int64) []byte {
	bytes := make([]byte, 8)
	binary.PutVarint(bytes, i)
	return bytes
}
*/
