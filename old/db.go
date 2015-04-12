package main

import (
	"fmt"
	"github.com/sapk/osmpbf"
	"io"
	"log"
	"os"
)

type FileDescriptor struct {
	Nodes       []int64
	Ways        []int64
	Relations   []int64
	NodesId     []int64
	WaysId      []int64
	RelationsId []int64
}

type Db struct {
	Bucket        map[int64][2]float64
	Wanted        []int64
	Decoder       *osmpbf.Decoder
	File          *os.File
	Descriptor    FileDescriptor
	maxBucketSize int
	nbProc        int
}

// Open returns a new Db that reads from file.
func OpenDB(file string, maxBucketSize int) (*Db, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	//maxBucketSize = 5 * 1024 * 1024
	return &Db{
		Bucket:        make(map[int64][2]float64),
		Decoder:       osmpbf.NewDecoder(f),
		File:          f,
		maxBucketSize: maxBucketSize,
	}, nil
}

// Scan before processing.
func (this *Db) Scan() (FileDescriptor, error) {
	//var nc, wc, rc int64
	for {
		//pos, _ := this.File.Seek(0, 1)
		if v, pos, err := this.Decoder.Decode(); err == io.EOF {
			break
		} else if err != nil {
			return this.Descriptor, err
		} else {
			switch v := v.(type) {
			case *osmpbf.Node:
				// Process Node v.
				//nc++
				if len(this.Descriptor.Nodes) == 0 || this.Descriptor.Nodes[len(this.Descriptor.Nodes)-1] != pos {
					this.Descriptor.Nodes = Extend(this.Descriptor.Nodes, pos)
					this.Descriptor.NodesId = Extend(this.Descriptor.NodesId, v.ID)
				}
			case *osmpbf.Way:
				// Process Way v.
				//wc++
				if len(this.Descriptor.Ways) == 0 || this.Descriptor.Ways[len(this.Descriptor.Ways)-1] != pos {
					this.Descriptor.Ways = Extend(this.Descriptor.Ways, pos)
					this.Descriptor.WaysId = Extend(this.Descriptor.WaysId, v.ID)
				}
			case *osmpbf.Relation:
				// Process Relation v.
				//rc++
				if len(this.Descriptor.Relations) == 0 || this.Descriptor.Relations[len(this.Descriptor.Relations)-1] != pos {
					this.Descriptor.Relations = Extend(this.Descriptor.Relations, pos)
					this.Descriptor.RelationsId = Extend(this.Descriptor.RelationsId, v.ID)
				}
			default:
				return this.Descriptor, fmt.Errorf("unknown type %T", v)
			}
		}
		/*
			if (nc+wc+rc)%2000000 == 0 {
				log.Printf("Nodes: %dk, Ways: %dk, Relations: %dk", nc/1000, wc/1000, rc/1000)
			}
		*/
	}
	//this.Descriptor = FileDescriptor{sc, nc, sw, wc, sr, rc}

	end, _ := this.File.Seek(0, 1)
	this.Descriptor.Nodes = Extend(this.Descriptor.Nodes, end)
	this.Descriptor.Ways = Extend(this.Descriptor.Ways, end)
	this.Descriptor.Relations = Extend(this.Descriptor.Relations, end)
	this.Descriptor.NodesId = Extend(this.Descriptor.NodesId, 0)
	this.Descriptor.RelationsId = Extend(this.Descriptor.RelationsId, 0)
	this.Descriptor.RelationsId = Extend(this.Descriptor.RelationsId, 0)

	//log.Printf("DB contains Nodes: %dk, Ways: %dk, Relations: %dk", nc/1000, wc/1000, rc/1000)
	log.Printf("DB descritor Nodes: %d, Ways: %d, Relations: %d",
		len(this.Descriptor.Nodes), len(this.Descriptor.Ways), len(this.Descriptor.Relations))

	//log.Printf("%s", this.Descriptor)

	this.File.Seek(this.Descriptor.Nodes[0], 0)

	return this.Descriptor, nil
}

// Get all nodes inside a bbox .
func (this *Db) Start(n int) error {
	//this.Descriptor.Nodes = make([]int64, 0, 10)
	this.nbProc = n
	return this.Decoder.Start(this.nbProc)
}

func formatBBOX(bbox [2][2]float64) [2][2]float64 {
	if bbox[0][0] > bbox[1][0] {
		tmp := bbox[0][0]
		bbox[0][0] = bbox[1][0]
		bbox[1][0] = tmp
	}
	if bbox[0][1] > bbox[1][1] {
		tmp := bbox[0][1]
		bbox[0][1] = bbox[1][1]
		bbox[1][1] = tmp
	}
	return bbox
}

// Get all nodes inside a bbox .
func (this *Db) GetWaysContaining(nodes map[int64]*osmpbf.Node) (map[int64]*osmpbf.Way, error) {

	ways := make(map[int64]*osmpbf.Way)
	//Le dernier est un fake
	for i := 0; i < len(this.Descriptor.Ways)-1; i++ {
		objects, err := this.Decoder.DecodeBlocAt(this.Descriptor.Ways[i])
		if err != nil {
			log.Fatal(err)
		}
		for _, v := range objects {
			switch v := v.(type) {
			case *osmpbf.Way:
				for _, id := range v.NodeIDs {
					if _, ok := nodes[id]; ok {
						ways[v.ID] = v
						break
					}
				}
			}
		}
	}

	return ways, nil
}

// Get all nodes inside a bbox .
func (this *Db) GetNodesIn(bbox [2][2]float64) (map[int64]*osmpbf.Node, error) {
	bbox = formatBBOX(bbox)
	//log.Printf("Searching nodes in bbox %s ... \n", bbox)
	//var nodes int64[]
	nodes := make(map[int64]*osmpbf.Node)
	//TODO thinks if it is better like that or with bucket
	//Le dernier est un fake
	for i := 0; i < len(this.Descriptor.Nodes)-1; i++ {
		objects, err := this.Decoder.DecodeBlocAt(this.Descriptor.Nodes[i])
		if err != nil {
			log.Fatal(err)
		}
		for _, v := range objects {
			switch v := v.(type) {
			case *osmpbf.Node:
				if v.Lon > bbox[0][0] && v.Lon < bbox[1][0] && v.Lat > bbox[0][1] && v.Lat < bbox[1][1] {
					//nodes = Extend(nodes, v.ID)
					nodes[v.ID] = v
				}
			}
		}
	}

	return nodes, nil
}
func (this *Db) GetNodesOfWay(w *osmpbf.Way) ([]*osmpbf.Node, error) {
	var err error
	nodes := make([]*osmpbf.Node, len(w.NodeIDs))
	log.Printf("Searching for %s", w.NodeIDs)

	for i, id := range w.NodeIDs {
		nodes[i], err = this.GetNode(id)
		if err != nil {
			return nil, err
		}

	}

	return nodes, nil
}

// Get on node by id.
func (this *Db) GetNode(id int64) (*osmpbf.Node, error) {

	var i int
	for i = 1; i < len(this.Descriptor.NodesId); i++ {
		uid := this.Descriptor.NodesId[i]
		//log.Printf("search id:%d uid:%d i: %d pos:%d \n", id, uid, i, this.Descriptor.Nodes[i])
		if id < uid {
			//log.Printf("Finded @ id:%d uid:%d i: %d pos:%d \n", id, this.Descriptor.NodesId[i-1], i-1, this.Descriptor.Nodes[i-1])
			break
		}
		/*
			else if i == len(this.Descriptor.WaysId)-1 {
				// On arrive à la fin mais on pas matché => On arrête ici ça sert à rien le dernier étant vide
				return nil, fmt.Errorf("Way not found !", nil)
			}
		*/
		//TODO optimize in odrer it doesn't look the last if the id doesn't exist
	}
	//log.Printf("%d \n", i-1)

	objects, err := this.Decoder.DecodeBlocAt(this.Descriptor.Nodes[i-1])
	if err != nil {
		log.Fatal(err)
	}
	for _, v := range objects {
		switch v := v.(type) {
		case *osmpbf.Node:
			//log.Printf("%d\n", v.ID)
			if v.ID == id {
				//log.Printf("%d\n", v.ID)
				return v, nil
			}
		}
	}
	return nil, fmt.Errorf("Node not found !", nil)
}

// Get on way by id.
func (this *Db) GetWay(id int64) (*osmpbf.Way, error) {
	var i int
	for i = 1; i < len(this.Descriptor.WaysId); i++ {
		uid := this.Descriptor.WaysId[i]
		//log.Printf("search id:%d uid:%d i: %d pos:%d \n", id, uid, i, this.Descriptor.Ways[i])
		if id < uid {
			//log.Printf("Finded @ id:%d uid:%d i: %d pos:%d \n", id, this.Descriptor.WaysId[i-1], i-1, this.Descriptor.Ways[i-1])
			break
		}
		/*
			else if i == len(this.Descriptor.WaysId)-1 {
				// On arrive à la fin mais on pas matché => On arrête ici ça sert à rien le dernier étant vide
				return nil, fmt.Errorf("Way not found !", nil)
			}
		*/
		//TODO optimize in odrer it doesn't look the last if the id doesn't exist
	}
	//log.Printf("%d \n", i-1)

	objects, err := this.Decoder.DecodeBlocAt(this.Descriptor.Ways[i-1])
	if err != nil {
		log.Fatal(err)
	}
	for _, v := range objects {
		switch v := v.(type) {
		case *osmpbf.Way:
			//log.Printf("%d\n", v.ID)
			if v.ID == id {
				//log.Printf("%d\n", v.ID)
				return v, nil
			}
		}
	}
	return nil, fmt.Errorf("Way not found !", nil)
}

func Extend(slice []int64, element int64) []int64 {
	//log.Printf("len: %d, cap: %d\n", len(slice), cap(slice))
	n := len(slice)
	if n == cap(slice) {

		//log.Println("slice is full!")
		// Slice is full; must grow.
		// We double its size and add 1, so if the size is zero we still grow.
		newSlice := make([]int64, len(slice), len(slice)+10)
		copy(newSlice, slice)
		slice = newSlice
	}
	slice = slice[0 : n+1]
	slice[n] = element
	//log.Printf("%s\n", slice)
	return slice
}

func ExtendWays(slice []osmpbf.Way, element osmpbf.Way) []osmpbf.Way {
	//log.Printf("len: %d, cap: %d\n", len(slice), cap(slice))
	n := len(slice)
	if n == cap(slice) {

		//log.Println("slice is full!")
		// Slice is full; must grow.
		// We double its size and add 1, so if the size is zero we still grow.
		newSlice := make([]osmpbf.Way, len(slice), len(slice)+10)
		copy(newSlice, slice)
		slice = newSlice
	}
	slice = slice[0 : n+1]
	slice[n] = element
	//log.Printf("%s\n", slice)
	return slice
}
