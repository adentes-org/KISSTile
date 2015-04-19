package api

import (
	"../db"
	"../geo"
	"bytes"
	"fmt"
	"github.com/emirpasic/gods/sets/treeset"
	"github.com/gorilla/mux"
	"github.com/sapk/osmpbf"
	"image"
	"image/jpeg"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"
)

const tileFolder = "tiles/"

type Api struct {
	db        *db.Db
	isRunning bool
	queueId   int
	queueNb   int
}

var version = "0.6-beta"

func Init(db *db.Db) *Api {

	return &Api{
		db:        db,
		isRunning: false,
	}
}
func (this *Api) Version(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, version)
}

func (this *Api) Status(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "running")
}

func (this *Api) Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome!")
}

func (this *Api) Tile(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	vars := mux.Vars(r)
	z, _ := strconv.Atoi(vars["zoom"])
	x, _ := strconv.Atoi(vars["x"])
	y, _ := strconv.Atoi(vars["y"])

	p := path.Join(tileFolder, vars["zoom"], vars["x"], vars["y"]+".jpg")
	if _, err := os.Stat(p); os.IsNotExist(err) {
		/*
			if this.isRunning {
				http.Error(w, http.StatusText(500), 500)
				log.Printf("ERROR @GENERATION : %s", time.Since(start).String())
				return
			}
		*/
		qID := this.queueNb
		this.queueNb += 1
		for this.isRunning || this.queueId != qID {
			time.Sleep(150 * time.Millisecond)
		}
		//on revérifi qu'il a pas été généré entre temps
		if _, err := os.Stat(p); os.IsExist(err) {
			b, _ := ioutil.ReadFile(p)
			w.Header().Set("Content-Type", "image/jpeg")
			w.Header().Set("Content-Length", strconv.Itoa(len(b)))
			if _, err := w.Write(b); err != nil {
				log.Println("unable to write image.")
			}
			log.Printf("Serving tile cache after waiting in : %s", time.Since(start).String())
			return
		}

		this.isRunning = true
		//fmt.Fprintln(w, "Generating file zoom :", vars["zoom"], "x,y :", vars["x"], ",", vars["y"])
		tile := geo.NewTileFromZXY(z, x, y)
		//fmt.Fprintf(w, "\nResulting tile : %v", tile)
		//fmt.Fprintf(w, "\nNb tile at level %d : %.0f", tile.Z, geo.NbAtZLevel(tile.Z))
		plat, plon := geo.PrecisionAtZLevel(tile.Z)
		log.Printf("Resulting precision : %v %v", plat, plon)
		bbox := tile.GetBBOX()
		//fmt.Fprintf(w, "\nResulting bbox : %v", bbox)

		//TODO
		log.Printf("Starting way findind in bbox %v", bbox)
		ways, _ := this.db.WayIndex.GetWayInBBox(bbox, "natural")
		//fmt.Fprintf(w, "\nResulting ways : %d", len(ways))
		//log.Printf("List %v", ways)
		wanted := treeset.NewWith(db.CompareInt64)
		for _, way := range ways {
			wanted.Add(way)
			// element is the element from someSlice for where we are
		}

		//TODO add function to filter by zoom rendering (features)
		log.Printf("Searching for %d ways", wanted.Size())
		var found map[int64]*osmpbf.Way
		found, _ = this.db.GetWays(wanted)
		log.Printf("%d ways found", len(found))
		log.Printf("TIME ELAPSED @WaysFound : %s", time.Since(start).String())

		wanted_node := treeset.NewWith(db.CompareInt64)
		var nodeId int64
		for _, way := range found {
			for _, nodeId = range way.NodeIDs {
				wanted_node.Add(nodeId)
			}
		}
		log.Printf("Searching for %d nodes", wanted_node.Size())
		//founded_nodes, _ := this.db.GetNodes(&wanted_node)
		//TODO
		founded_nodes, _ := this.db.GetNodes(wanted_node, nil)
		log.Printf("%d nodes founded", len(founded_nodes))
		log.Printf("TIME ELAPSED @NodesFound : %s", time.Since(start).String())
		img, err := tile.DrawTile(found, founded_nodes, true)
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			log.Printf("ERROR @GENERATION : %s", time.Since(start).String())
			return
		} else {
			this.saveImageJpeg(tile, &img)
			this.writeImageJpeg(w, &img)
			log.Printf("TIME ELAPSED @END : %s", time.Since(start).String())
		}
		this.queueId++
		this.isRunning = false
	} else {
		b, _ := ioutil.ReadFile(p)
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Content-Length", strconv.Itoa(len(b)))
		if _, err := w.Write(b); err != nil {
			log.Println("unable to write image.")
		}

		log.Printf("Serving tile cache in : %s", time.Since(start).String())
	}
}

// writeImage encodes an image 'img' in jpeg format and writes it into ResponseWriter.
func (this *Api) writeImageJpeg(w http.ResponseWriter, img *image.Image) {

	buffer := new(bytes.Buffer)
	if err := jpeg.Encode(buffer, *img, nil); err != nil {
		log.Println("unable to encode image.")
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
	if _, err := w.Write(buffer.Bytes()); err != nil {
		log.Println("unable to write image.")
	}
}

// writeImage encodes an image 'img' in jpeg format and writes it into file
func (this *Api) saveImageJpeg(t *geo.Tile, img *image.Image) {

	buffer := new(bytes.Buffer)
	if err := jpeg.Encode(buffer, *img, nil); err != nil {
		log.Println("unable to encode image.")
	}

	p := path.Join(tileFolder, strconv.Itoa(t.Z), strconv.Itoa(t.X), strconv.Itoa(t.Y)+".jpg")
	if os.MkdirAll(path.Dir(p), 0777) != nil {
		panic("Unable to create directory for tagfile!")
	}
	os.Remove(p)
	ioutil.WriteFile(p, buffer.Bytes(), 0777)
}

/*
func Capabilities(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "//TOIMPLEMENT")
}

func Map(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "//TOIMPLEMENT")
}
*/
