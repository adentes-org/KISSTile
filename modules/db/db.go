// db
package db

import (
	"../geo"
	"fmt"
	"github.com/sapk/osmpbf"
	"io"
	"log"
	"os"
	"time"
)

//TODO search better
const MaxUint64 = ^uint64(0)
const MaxInt64 = int64(MaxUint64 >> 1)

//TODO used Sizeof and used a Mo,Mb type
const CacheSize = 500000

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

//TODO pass to Point to save space
//TODO optimize cache to take account of last access time to grabage not used
type Db struct {
	Decoder    *osmpbf.Decoder
	File       *os.File
	Descriptor FileDescriptor
	Cache      *NodeCache
	nbProc     int
}

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

// Get all nodes inside a bbox .
func (this *Db) Analyze(n int) error {
	//this.Descriptor.Nodes = make([]int64, 0, 10)
	this.nbProc = n
	err := this.Decoder.Start(this.nbProc)
	if err != nil {
		return err
	}

	_, err = this.Describe()
	if err != nil {
		return err
	}

	err = this.ParseWays()
	if err != nil {
		return err
	}

	return nil
}

// Scan before processing need to be call frist after start.
func (this *Db) ParseWays() error {
	//	ways := make(map[int64]*osmpbf.Way)
	//Le dernier est un faux

	start := time.Now()

	var bb geo.Bbox
	var ways []*osmpbf.Way
	wanted := make(map[int64]*osmpbf.Node)

	var cw, cn int64
	//TODO find nodes by block of wa in order to not redecode the start of a block
	for i := 0; i < len(this.Descriptor.Ways)-1; i++ {
		objects, err := this.Decoder.DecodeBlocAt(this.Descriptor.Ways[i])
		if err != nil {
			return err
		}
		for _, v := range objects {
			switch v := v.(type) {
			case *osmpbf.Way:
				cw++
				ways = ExtendWays(ways, v)
				//				ways[v.ID] = v
				//TODO check ways with no nodes maybe ?
				/*
					for id, nodeId := range v.NodeIDs {
						cn++
						node, err := this.GetNode(nodeId)
						if err != nil {
							return err
						}
						p := geo.Point{node.Lon, node.Lat}
						if id == 0 {
							bb = geo.Bbox{p, p}
						} else {
							//TODO
							if !p.InBbox(bb) {
								//We need to enlarge the Bbox
							}
						}
					}
				*/

				//TODO used an orderedt list
				for _, nodeId := range v.NodeIDs {
					wanted[nodeId] = nil
				}
				/*
					//TODO declancher if end
					if len(wanted) > CacheSize {
						log.Printf("On parse les points pour %d soit %d nodes recherchés", len(ways), len(wanted))
						// On cherche les points
						for i := 0; i < len(this.Descriptor.Nodes)-1; i++ {
							ns, err := this.Decoder.DecodeBlocAt(this.Descriptor.Nodes[i])
							if err != nil {
								return err
							}
							for _, v := range ns {
								switch v := v.(type) {
								case *osmpbf.Node:
									if _, ok := wanted[v.ID]; ok {
										delete(wanted, v.ID)
										found[v.ID] = v
									}
									break
								}
							}
						}

						for _, way := range ways {
							for id, nodeId := range way.NodeIDs {
								cn++
								node := found[nodeId]
								p := geo.Point{node.Lon, node.Lat}

								if id == 0 {
									bb = geo.Bbox{p, p}
								} else {
									//TODO
									if !p.InBbox(bb) {
										//We need to enlarge the Bbox
									}
								}
							}
						}

						log.Printf("All %d nodes found", len(found))

						wanted = make(map[int64]*osmpbf.Node)
						found = make(map[int64]*osmpbf.Node)
						ways = make([]*osmpbf.Way, 0)
					}
				*/
				break
			}
			/*
				if cw%30000 == 1 {
					log.Printf("Ways parsed : %d soit %d nodes", cw, cn)
				}
			*/
		}
		if len(wanted) > CacheSize || i == len(this.Descriptor.Ways)-2 {
			log.Printf("On parse les points pour %d ways soit %d nodes recherchés", len(ways), len(wanted))
			found, _ := this.GetNodes(&wanted)
			//On reset wanted to save space (not obligated but in case it not clear by getNodes)
			wanted = make(map[int64]*osmpbf.Node, 0)

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
			}

			ways = make([]*osmpbf.Way, 0)

			estimation := time.Since(start).Minutes() * (float64(this.Descriptor.WayCount) / float64(cw))

			time_esti, _ := time.ParseDuration(fmt.Sprintf("%.4fm", estimation))
			log.Printf("%dk/%dk %.2f/100 TIME ELAPSED : %s ESTIMATION : %s\r", cw/1000, this.Descriptor.WayCount/1000, float64(cw*100)/float64(this.Descriptor.WayCount), time.Since(start).String(), time_esti.String())
		}
	}

	return nil
}

// Scan before processing need to be call frist after start.
func (this *Db) Describe() (FileDescriptor, error) {
	var nc, wc, rc int64
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
		}
		/*
			if (nc+wc+rc)%2000000 == 0 {
				log.Printf("Nodes: %dk, Ways: %dk, Relations: %dk", nc/1000, wc/1000, rc/1000)
			}
		*/
	}

	//On ajout une nouvelle entrée de fin pour faciliter le traitement
	//TODO find better
	end, _ := this.File.Seek(0, 1)
	this.Descriptor.Nodes = ExtendInt64(this.Descriptor.Nodes, end)
	this.Descriptor.Ways = ExtendInt64(this.Descriptor.Ways, end)
	this.Descriptor.Relations = ExtendInt64(this.Descriptor.Relations, end)
	this.Descriptor.NodesId = ExtendInt64(this.Descriptor.NodesId, MaxInt64)
	this.Descriptor.RelationsId = ExtendInt64(this.Descriptor.RelationsId, MaxInt64)
	this.Descriptor.RelationsId = ExtendInt64(this.Descriptor.RelationsId, MaxInt64)

	this.Descriptor.NodeCount = nc
	this.Descriptor.WayCount = wc
	this.Descriptor.RelationCount = rc

	log.Printf("DB contains Nodes: %dk, Ways: %dk, Relations: %dk", nc/1000, wc/1000, rc/1000)
	log.Printf("DB descritor Nodes: %d, Ways: %d, Relations: %d",
		len(this.Descriptor.Nodes), len(this.Descriptor.Ways), len(this.Descriptor.Relations))

	//log.Printf("%s", this.Descriptor)

	this.File.Seek(this.Descriptor.Nodes[0], 0)

	return this.Descriptor, nil
}

// Get multiples node by ids at a time.
func (this *Db) GetNodes(wanted *map[int64]*osmpbf.Node) (map[int64]*osmpbf.Node, error) {
	found := make(map[int64]*osmpbf.Node)
	// On cherche les points
	for i := 0; i < len(this.Descriptor.Nodes)-1; i++ {
		ns, err := this.Decoder.DecodeBlocAt(this.Descriptor.Nodes[i])
		if err != nil {
			return nil, err
		}
		for _, v := range ns {
			switch v := v.(type) {
			case *osmpbf.Node:
				if _, ok := (*wanted)[v.ID]; ok {
					delete(*wanted, v.ID)
					found[v.ID] = v
				}
				break
			}
		}
		//TODO only read needed
		/*
			if len(wanted) == 0 {
				// on arrete y'a plus rien à chercher
				break
			}
		*/
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
