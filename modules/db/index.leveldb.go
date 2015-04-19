// db
//TODO try leveldb or ameliore with block and compression like pbf with miinimal bbox
package db

import (
	//"github.com/mattn/go-sqlite3"
	"./../geo"
	//	"bytes"
	"encoding/binary"
	//	"errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Index struct {
	//	db     *bolt.DB
	//	bucket *bolt.Bucket
	db         *leveldb.DB
	batch      *leveldb.Batch
	batch_size int
	last       int64
}

const bucket_name string = "Ways"

//TODO add level of way in order to filter way in

func (this *Index) Close() error {
	return this.db.Close()
}
func (this *Index) GetWayInBBox(bb geo.Bbox, tag string) ([]int64, error) {
	var ret []int64
	size := bb.Size()
	iter := this.db.NewIterator(util.BytesPrefix([]byte(tag+":")), nil)
	for iter.Next() {
		key := iter.Key()
		//log.Println(iter.Key())
		if key[0] != byte('_') {
			tmp := Bboxfrombytes(iter.Value())
			//TODO determine best ratio
			if bb.IntersectWith(tmp) && size < tmp.Size()*1000 && size*1000 > tmp.Size() {
				//log.Println(iter.Key())
				//				ret = append(ret, Int64frombytes(key))
				k, _ := strconv.ParseInt(strings.Split(string(key[:]), ":")[1], 10, 64)
				ret = append(ret, k)
			}
		}
	}
	iter.Release()
	err := iter.Error()
	log.Printf("%d Ways found in %v", len(ret), bb)
	return ret, err
}
func (this *Index) LoadOrCreateOf(file string) (*Index, error) {

	file = this.getFilenameFromFile(file)

	//log.Printf("for debug removing : %s", file)
	//TODO renewable or recache analyze
	//os.RemoveAll(file)
	//	return this, errors.New("Not implemented")
	var err error
	//TODO re-get last states
	//Uses recover in order to catch also unclosed
	//Test upgrade size file to limit operation ???

	this.db, err = leveldb.OpenFile(file, &opt.Options{
		BlockSize:                     32 * opt.KiB,
		CompactionTableSize:           2 * opt.MiB,
		CompactionTableSizeMultiplier: 8,
		CompactionTotalSize:           20 * opt.MiB,
		CompactionTotalSizeMultiplier: 10,
		CompactionL0Trigger:           8,
	})
	if err != nil {
		log.Printf("There seem to have a problem with the index. Will try to correct that %v", err)

		this.db, err = leveldb.RecoverFile(file, &opt.Options{
			BlockSize:                     32 * opt.KiB,
			CompactionTableSize:           2 * opt.MiB,
			CompactionTableSizeMultiplier: 8,
			CompactionTotalSize:           20 * opt.MiB,
			CompactionTotalSizeMultiplier: 10,
			CompactionL0Trigger:           8,
		})
	}
	this.batch = new(leveldb.Batch)
	this.batch_size = 32 * opt.KiB / (8 * 5)
	if err != nil {
		log.Printf("%v", err)
	}

	//	stats, _ := this.db.GetProperty("leveldb.stats")
	//	log.Printf("stats : %v", stats)
	return this, nil
}

func (this *Index) Add(wayid int64, tag string, bb geo.Bbox) error {
	//TODO optimize for batch write
	//TODO use  this.tmp and write by bloclk useless ?
	//TODO check if el exist in order to not count in count
	//data := Int64bytes(wayid)
	//tmp := Float64bytes(bb[0].Lat)
	data := Float64bytes(bb[0].Lat)
	data = append(data, Float64bytes(bb[0].Lon)[:]...)
	data = append(data, Float64bytes(bb[1].Lat)[:]...)
	data = append(data, Float64bytes(bb[1].Lon)[:]...)
	//, Float64bytes(bb[0].Lon), Float64bytes(bb[1].Lat), Float64bytes(bb[1].Lon))
	/*
		this.db.Put(Float64bytes(bb[0].Lat), data, &opt.WriteOptions{
			true,
		})
	*/
	this.last = wayid
	//	this.batch.Put(Int64bytes(wayid), data)
	this.batch.Put([]byte(tag+":"+strconv.FormatInt(wayid, 10)), data)

	//TODO set const for batch max_size
	if this.batch.Len() > this.batch_size {
		err := this.PullBatch(false)
		if err != nil {
			log.Printf("error %v", err)
		}
	}

	return nil
}

func (this *Index) ResetOf(file_base string) error {
	file := this.getFilenameFromFile(file_base)
	os.RemoveAll(file)
	this, _ = this.LoadOrCreateOf(file_base)
	return nil
}

func (this *Index) PullBatch(sync bool) error {
	//TODO check if el exist in order to not count in count
	this.batch.Put([]byte("_counter"), Int64bytes(this.Size()+int64(this.batch.Len())))
	this.batch.Put([]byte("_last"), Int64bytes(this.last))
	err := this.db.Write(this.batch, &opt.WriteOptions{
		Sync: sync,
	})
	this.batch = new(leveldb.Batch)
	return err
}

func (this *Index) Size() int64 {
	var counter int64
	bytes, err := this.db.Get([]byte("_counter"), nil)
	if err != nil {
		counter = 0
	} else {
		counter = Int64frombytes(bytes)
	}
	return counter
	//	this.db.SizeOf()
}
func (this *Index) Last() int64 {
	var last int64
	bytes, err := this.db.Get([]byte("_last"), nil)
	if err != nil {
		last = -1
	} else {
		last = Int64frombytes(bytes)
	}
	return last
}

//For testing purpose
//TODO
func (this *Index) Get(wayid int64) (geo.Bbox, error) {
	bytes, err := this.db.Get(Int64bytes(wayid), nil)
	//log.Println(len(bytes))
	return geo.Bbox{geo.Point{Float64frombytes(bytes[0:8]), Float64frombytes(bytes[8:16])}, geo.Point{Float64frombytes(bytes[16:24]), Float64frombytes(bytes[24:32])}}, err
}

func Bboxfrombytes(bytes []byte) geo.Bbox {
	return geo.Bbox{geo.Point{Float64frombytes(bytes[0:8]), Float64frombytes(bytes[8:16])}, geo.Point{Float64frombytes(bytes[16:24]), Float64frombytes(bytes[24:32])}}
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

/*
func Intbytes(i int) []byte {
	bytes := make([]byte, 4)
	binary.PutVarint(bytes, i)
	return bytes
}

func Intfrombytes(bytes []byte) int {
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.LittleEndian, &ret)
	return ret
}
*/
func Int64frombytes(bytes []byte) int64 {
	i, _ := binary.Varint(bytes)
	return i
}
