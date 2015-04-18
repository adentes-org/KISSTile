package api

import (
	"../db"
	"../geo"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

type Api struct {
	db *db.Db
}

var version = "0.6-beta"

func Init(db *db.Db) *Api {

	return &Api{
		db: db,
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
	vars := mux.Vars(r)
	z, _ := strconv.Atoi(vars["zoom"])
	x, _ := strconv.Atoi(vars["x"])
	y, _ := strconv.Atoi(vars["y"])
	fmt.Fprintln(w, "Generating file zoom :", vars["zoom"], "x,y :", vars["x"], ",", vars["y"])
	tile := geo.NewTileFromZXY(z, x, y)
	fmt.Fprintf(w, "\nResulting tile : %v", tile)
	fmt.Fprintf(w, "\nNb tile at level %d : %d", tile.Z, geo.NbAtZLevel(tile.Z))
	plat, plon := geo.PrecisionAtZLevel(tile.Z)
	fmt.Fprintf(w, "\nResulting precision : %v %v", plat, plon)
	bbox := tile.GetBBOX()
	fmt.Fprintf(w, "\nResulting bbox : %v", bbox)
	bbox.OrderBbox()
	fmt.Fprintf(w, "\nResulting bbox : %v", bbox)

	//TODO
	log.Printf("Starting way findind in bbox %v", bbox)
	this.db.WayIndex.GetWayInBBox(bbox)
	this.db.WayIndex.Close()
}

/*
func Capabilities(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "//TOIMPLEMENT")
}

func Map(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "//TOIMPLEMENT")
}
*/
