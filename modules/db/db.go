// db
package db

import (
	"../geo"
	"fmt"
	"github.com/emirpasic/gods/sets/treeset"
	//"github.com/emirpasic/gods/utils"
	"github.com/sapk/osmpbf"
	"io"
	"log"
	"os"
	"time"
)

//TODO use WayCount and NodeCount to optimize parsing
//TODO search better
const MaxUint64 = ^uint64(0)
const MaxInt64 = int64(MaxUint64 >> 1)

//TODO used Sizeof and used a Mo,Mb type
const CacheSize = 750000

//0,65Ko/Obj
//const CacheSize = -1

//TODO pass to Point to save space
//TODO optimize cache to take account of last access time to grabage not used
type Db struct {
	Decoder    *osmpbf.Decoder
	File       *os.File
	Descriptor FileDescriptor
	Cache      *NodeCache
	WayIndex   Index
	nbProc     int
}

//TODO in descritor use array for count that increment in that the last is the total
// Open returns a new Db that reads from file.
func OpenDB(file string) (*Db, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	return &Db{
		Decoder: osmpbf.NewDecoder(f),
		File:    f,
		Cache:   CreateNodeCache(CacheSize),
	}, nil
}

func (this *Db) Close() error {
	return this.WayIndex.Close()
}

// Get all nodes inside a bbox .
func (this *Db) Analyze(n int) error {
	//this.Descriptor.Nodes = make([]int64, 0, 10)
	this.nbProc = n
	err := this.Decoder.Start(this.nbProc)
	if err != nil {
		return err
	}

	//TODO load from backup if it's allready analyze
	_, err = this.Descriptor.LoadOrCreateOf(this.File.Name())
	if err != nil {
		_, err = this.Describe()
		if err != nil {
			return err
		}
	}
	this.Descriptor.Save(this.File.Name())

	//TODO check if Index is file from allr eady analyze and catch up analyze
	_, err = this.WayIndex.LoadOrCreateOf(this.File.Name())
	if err != nil {
		log.Printf("error : %v", err)
	}
	//Si la base est inférieur on parse
	//TODO better check and retrieve where we stop or what is missing
	log.Printf("There is %d ways indexed", this.WayIndex.Size())
	log.Printf("Last  ways indexed : %d", this.WayIndex.Last())
	log.Printf("There is %d ways in file", this.Descriptor.TotalWay())
	if this.WayIndex.Size() < this.Descriptor.TotalWay() {
		log.Printf("Start parsing for %d ways", this.Descriptor.TotalWay()-this.WayIndex.Size())
		//TODO not reset and catch up
		//this.WayIndex.ResetOf(this.File.Name())
		err = this.ParseWays()
		if err != nil {
			return err
		}
	}
	// Test
	/*
		log.Printf("%v", this.Descriptor.WaysId[0])
		a, e := this.WayIndex.Get(this.Descriptor.WaysId[0])
		log.Printf("%v %v", a, e)
	*/
	return nil
}
func CompareInt64(a, b interface{}) int {
	aInt := a.(int64)
	bInt := b.(int64)
	switch {
	case aInt > bInt:
		return 1
	case aInt < bInt:
		return -1
	default:
		return 0
	}
}

// Scan before processing need to be call frist after start.
func (this *Db) ParseWays() error {
	//Le dernier est un faux

	start := time.Now()

	//On pull les dernier et on sync in order to "close" database at the end.
	defer this.WayIndex.PullBatch(true)
	//	defer this.WayIndex.db.Close()

	var bb geo.Bbox
	var ways []*osmpbf.Way
	wanted := treeset.NewWith(CompareInt64)
	//wanted := make(map[int64]*osmpbf.Node)

	var cw, cow, cn, last int64
	cow = this.WayIndex.Size()
	last = this.WayIndex.Last()
	found := make(map[int64]*osmpbf.Node, 0)
	//TODO find nodes by block of wa in order to not redecode the start of a block
	for i := 0; i < len(this.Descriptor.Ways)-1; i++ {
		if this.Descriptor.WaysId[i+1] < last {
			//There always be i+1 since the for condition
			continue
		}
		//log.Printf("Parsing block : %d", i)
		objects, err := this.Decoder.DecodeBlocAt(this.Descriptor.Ways[i])
		if err != nil {
			return err
		}
		for _, v := range objects {
			switch v := v.(type) {
			case *osmpbf.Way:
				//TODO better take to long and my be more slow that the little advatage it gave by resuming
				//if has, _ := this.WayIndex.db.Has(Int64bytes(v.ID), nil); has
				if v.ID <= last {
					//log.Printf("Passing %d", v.ID)
					//cow++
					continue
				}

				cw++
				//log.Printf("Adding to search %d", v.ID)
				ways = ExtendWays(ways, v)
				//TODO check ways with no nodes maybe ?
				//TODO used an ordered list (invert)
				for _, nodeId := range v.NodeIDs {
					/*
						for i := 0; i < len(wanted); i++ {
							if nodeId <= wanted[i] {
								break
							}
						}
						//We don't insert if nodeId == wanted[i] (duplicate)
						if i == len(wanted) || nodeId < wanted[i] {
							wanted = append(wanted[:i], append([]int64{nodeId}, wanted[i:]...)...)
						}

						//wanted = append(wanted, nodeId)
						//wanted[nodeId] = nil
					*/
					wanted.Add(nodeId)
				}
				break
			}
		}
		if wanted.Size() > CacheSize || i == len(this.Descriptor.Ways)-2 {
			log.Printf("On parse les points pour %d ways soit %d nodes recherchés", len(ways), wanted.Size())
			//TODO reused allready found on previous round
			found, _ = this.GetNodes(wanted, found)
			//On reset wanted to save space (not obligated but in case it not clear by getNodes)
			wanted.Clear()

			for _, way := range ways {
				for id, nodeId := range way.NodeIDs {
					cn++
					node := found[nodeId]
					p := geo.Point{node.Lon, node.Lat}

					if id == 0 {
						bb = geo.Bbox{p, p}
					} else {
						//TODO
						//Will enlarge bb if needed
						bb.AddInnerPoint(p)
					}
				}
				// environ 15% de temps en plus
				tag := "other"
				if _, ok := way.Tags["natural"]; ok {
					tag = "natural"
				}
				this.WayIndex.Add(way.ID, tag, bb)
			}
			//TODO check For update of file
			//this.WayIndex.db.Sync()

			//For testing purpose
			//log.Printf("%v %v", ways[0].ID, ways[0])
			//a, e := this.WayIndex.Get(ways[0].ID)
			//log.Printf("%v %v", a, e)
			/* //TODO ajout par batch
			log.Println("Starting db insertion")
			this.WayIndex.AddBatch(ways, bb)
			log.Println("db insertion ended")
			*/
			ways = make([]*osmpbf.Way, 0)

			estimation := time.Since(start).Minutes() * (float64(this.Descriptor.TotalWay()-cow) / float64(cw))

			time_esti, _ := time.ParseDuration(fmt.Sprintf("%.4fm", estimation))
			log.Printf("%dk/%dk %.2f/100 TIME ELAPSED : %s ESTIMATION : %s\r", (cw+cow)/1000, this.Descriptor.TotalWay()/1000, float64((cw+cow)*100)/float64(this.Descriptor.TotalWay()), time.Since(start).String(), time_esti.String())
		}
	}

	found = make(map[int64]*osmpbf.Node, 0)
	wanted.Clear()
	return nil
}

// Scan before processing need to be call frist after start.
func (this *Db) Describe() (FileDescriptor, error) {
	var nc, wc, rc, last_pos int64
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
				nc++
				if len(this.Descriptor.Nodes) == 0 || this.Descriptor.Nodes[len(this.Descriptor.Nodes)-1] != pos {
					this.Descriptor.Nodes = ExtendInt64(this.Descriptor.Nodes, pos)
					this.Descriptor.NodesId = ExtendInt64(this.Descriptor.NodesId, v.ID)
				}
			case *osmpbf.Way:
				// Process Way v.
				wc++
				if len(this.Descriptor.Ways) == 0 || this.Descriptor.Ways[len(this.Descriptor.Ways)-1] != pos {
					this.Descriptor.Ways = ExtendInt64(this.Descriptor.Ways, pos)
					this.Descriptor.WaysId = ExtendInt64(this.Descriptor.WaysId, v.ID)
				}
			case *osmpbf.Relation:
				// Process Relation v.
				rc++
				if len(this.Descriptor.Relations) == 0 || this.Descriptor.Relations[len(this.Descriptor.Relations)-1] != pos {
					this.Descriptor.Relations = ExtendInt64(this.Descriptor.Relations, pos)
					this.Descriptor.RelationsId = ExtendInt64(this.Descriptor.RelationsId, v.ID)
				}
			default:
				return this.Descriptor, fmt.Errorf("Unknown type %T", v)
			}
			//Initialisation de la position
			if last_pos == 0 {
				last_pos = pos
			}
			//Si on change de block on insert les counts
			if last_pos != pos {
				last_pos = pos
				this.Descriptor.NodeCount = ExtendInt64(this.Descriptor.NodeCount, nc)
				this.Descriptor.WayCount = ExtendInt64(this.Descriptor.WayCount, wc)
				this.Descriptor.RelationCount = ExtendInt64(this.Descriptor.RelationCount, rc)
			}
		}
	}

	//On ajout une nouvelle entrée de fin pour faciliter le traitement
	//TODO find better
	end, _ := this.File.Seek(0, 1)
	this.Descriptor.Nodes = ExtendInt64(this.Descriptor.Nodes, end)
	this.Descriptor.Ways = ExtendInt64(this.Descriptor.Ways, end)
	this.Descriptor.Relations = ExtendInt64(this.Descriptor.Relations, end)
	this.Descriptor.NodesId = ExtendInt64(this.Descriptor.NodesId, MaxInt64)
	this.Descriptor.WaysId = ExtendInt64(this.Descriptor.WaysId, MaxInt64)
	this.Descriptor.RelationsId = ExtendInt64(this.Descriptor.RelationsId, MaxInt64)
	this.Descriptor.NodeCount = ExtendInt64(this.Descriptor.NodeCount, nc)
	this.Descriptor.WayCount = ExtendInt64(this.Descriptor.WayCount, wc)
	this.Descriptor.RelationCount = ExtendInt64(this.Descriptor.RelationCount, rc)
	//TODO use array
	//this.Descriptor.NodeCount = nc
	//this.Descriptor.WayCount = wc
	//this.Descriptor.RelationCount = rc

	log.Printf("DB contains Nodes: %dk, Ways: %dk, Relations: %dk", nc/1000, wc/1000, rc/1000)
	log.Printf("DB descritor Nodes: %d, Ways: %d, Relations: %d",
		len(this.Descriptor.Nodes), len(this.Descriptor.Ways), len(this.Descriptor.Relations))

	//log.Printf("%s", this.Descriptor)

	this.File.Seek(this.Descriptor.Nodes[0], 0)

	return this.Descriptor, nil
}

func (this *Db) getBlockContainingWays(wanted *treeset.Set) ([]int, error) {
	var ret []int
	lwi := 0
	l := (*wanted).Size()
	list := (*wanted).Values()
	for i := 0; i < len(this.Descriptor.Ways)-1; i++ {
		//this.Descriptor.NodesId[i] Contain the first id of the block i
		//The last is a fake and contain MaxInt64
		//TODO use a array of int64 less memery consuming than map
		for wi := lwi; wi < l; wi++ {
			if list[wi].(int64) >= this.Descriptor.WaysId[i] && list[wi].(int64) <= this.Descriptor.WaysId[i+1] {
				ret = append(ret, i)
				lwi = wi
				break
			}
		}
	}
	return ret, nil
}

// Get multiples node by ids at a time.
//TODO not used map array but ordered list and clear one at a time
//TODO determine block we only need to decrypt
func (this *Db) GetWays(wanted *treeset.Set) (map[int64]*osmpbf.Way, error) {
	found := make(map[int64]*osmpbf.Way)
	l := (*wanted).Size()
	list := (*wanted).Values()
	log.Printf("Nodes to find : %v", l)
	blocks, _ := this.getBlockContainingWays(wanted)
	for _, i := range blocks {
		ns, err := this.Decoder.DecodeBlocAt(this.Descriptor.Ways[i])
		if err != nil {
			return nil, err
		}
		for _, v := range ns {
			switch v := v.(type) {
			case *osmpbf.Way:
				if v.ID == list[0] {
					found[v.ID] = v
					if len(list) == 1 {
						//Il ne restait plus qu'un élément et on vient de le trouver
						log.Printf("All %d nodes found", len(found))
						list = nil
						return found, nil
					}
					//On enleve le première élément
					list = list[1:]
				}
				break
			}
		}
	}
	log.Printf("All %d ways found", len(found))

	return found, nil
}
func (this *Db) getBlockContainingNodes(wanted *treeset.Set) ([]int, error) {
	var ret []int
	lwi := 0
	l := (*wanted).Size()
	list := (*wanted).Values()
	for i := 0; i < len(this.Descriptor.Nodes)-1; i++ {
		//this.Descriptor.NodesId[i] Contain the first id of the block i
		//The last is a fake and contain MaxInt64
		//TODO use a array of int64 less memery consuming than map
		for wi := lwi; wi < l; wi++ {
			if list[wi].(int64) >= this.Descriptor.NodesId[i] && list[wi].(int64) <= this.Descriptor.NodesId[i+1] {
				ret = append(ret, i)
				lwi = wi
				break
			}
		}
	}
	return ret, nil
}

// Get multiples node by ids at a time.
func (this *Db) GetNodes(wanted *treeset.Set, alfound map[int64]*osmpbf.Node) (map[int64]*osmpbf.Node, error) {
	found := make(map[int64]*osmpbf.Node)
	nb := wanted.Size()
	for id, val := range alfound {
		if wanted.Contains(id) {
			found[id] = val
			wanted.Remove(id)
		}
	}
	nb -= wanted.Size()
	alfound = nil
	log.Printf("Nodes reused : %d", nb)
	// On cherche les points
	l := (*wanted).Size()
	list := (*wanted).Values()
	log.Printf("Nodes to find : %v", l)
	blocks, _ := this.getBlockContainingNodes(wanted)
	log.Printf("Block to read : %d / %d", len(blocks), len(this.Descriptor.NodesId))
	for _, i := range blocks {
		ns, err := this.Decoder.DecodeBlocAt(this.Descriptor.Nodes[i])
		if err != nil {
			return nil, err
		}
		for _, v := range ns {
			switch v := v.(type) {
			case *osmpbf.Node:
				if v.ID == list[0] {
					found[v.ID] = v
					if len(list) == 1 {
						//Il ne restait plus qu'un élément et on vient de le trouver
						log.Printf("All %d nodes found", len(found))
						list = nil
						return found, nil
					}
					//On enleve le première élément
					list = list[1:]
				}
				break
			}
		}

	}
	log.Printf("All %d nodes found", len(found))
	return found, nil
}

// Get on node by id.
func (this *Db) GetNode(id int64) (*osmpbf.Node, error) {
	// TODO implement a cache system and generalize method Get

	// On verifie qu'il est pas dans le cache
	if node, ok := this.Cache.get(id); ok {
		return node, nil
	}

	//	var ret *osmpbf.Node
	var i int
	for i = 1; i < len(this.Descriptor.NodesId); i++ {
		uid := this.Descriptor.NodesId[i]
		//log.Printf("search id:%d uid:%d i: %d pos:%d \n", id, uid, i, this.Descriptor.Nodes[i])
		if id < uid {
			//log.Printf("Finded @ id:%d uid:%d i: %d pos:%d \n", id, this.Descriptor.NodesId[i-1], i-1, this.Descriptor.Nodes[i-1])
			break
		}
		//TODO optimize in odrer it doesn't look the last if the id doesn't exist
	}
	objects, err := this.Decoder.DecodeBlocAt(this.Descriptor.Nodes[i-1])
	if err != nil {
		log.Fatal(err)
	}
	for _, v := range objects {
		switch v := v.(type) {
		case *osmpbf.Node:
			// On verifie qu'il n'est pas dans le cache
			if _, ok := this.Cache.get(id); !ok {
				// Si c'est le cas on le cache car il y'a de grande chance qu'il soit réutilisé
				this.Cache.add(v)
			}
			if v.ID == id {
				//On le retourne si c'est celui que l'on cherche
				return v, nil
				//			ret = v
				//TODO determine if we cache the nexts ?
			}
		}
	}
	/*
		if ret != nil {
			return ret, nil
		}
	*/
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

func ExtendInt64(slice []int64, element int64) []int64 {
	//log.Printf("len: %d, cap: %d\n", len(slice), cap(slice))
	n := len(slice)
	if n == cap(slice) {

		//log.Println("slice is full!")
		// Slice is full; must grow. We add 10 cases
		newSlice := make([]int64, len(slice), len(slice)+10)
		copy(newSlice, slice)
		slice = newSlice
	}
	slice = slice[0 : n+1]
	slice[n] = element
	//log.Printf("%s\n", slice)
	return slice
}

func ExtendWays(slice []*osmpbf.Way, element *osmpbf.Way) []*osmpbf.Way {
	//log.Printf("len: %d, cap: %d\n", len(slice), cap(slice))
	n := len(slice)
	if n == cap(slice) {

		//log.Println("slice is full!")
		// Slice is full; must grow.
		// We double its size and add 1, so if the size is zero we still grow.
		newSlice := make([]*osmpbf.Way, len(slice), len(slice)+10)
		copy(newSlice, slice)
		slice = newSlice
	}
	slice = slice[0 : n+1]
	slice[n] = element
	//log.Printf("%s\n", slice)
	return slice
}
