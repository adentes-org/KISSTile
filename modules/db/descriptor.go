package db

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

type FileDescriptor struct {
	Nodes         []int64
	NodeCount     int64
	Ways          []int64
	WayCount      int64
	Relations     []int64
	RelationCount int64
	NodesId       []int64
	WaysId        []int64
	RelationsId   []int64
}

func (this *FileDescriptor) LoadOrCreateOf(file string) (*FileDescriptor, error) {

	file = this.getFilenameFromFile(file)
	log.Printf("loading descriptor : %s", file)

	f, e := ioutil.ReadFile(file)
	if e != nil {
		log.Printf("File error: %v\n", e)
		return nil, e
	} else {
		json.Unmarshal(f, &this)
	}
	//log.Printf("Results: %v\n", this)

	return this, nil
}

func (this *FileDescriptor) Save(file string) error {
	file = this.getFilenameFromFile(file)
	desc, _ := json.Marshal(this)
	//log.Println(string(desc))
	ioutil.WriteFile(file, desc, 0777)
	return nil
}

func (this *FileDescriptor) getFilenameFromFile(file string) string {
	var extension = filepath.Ext(file)
	return strings.Join([]string{file[0 : len(file)-len(extension)], "desc"}, ".")
}
