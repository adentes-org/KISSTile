package api

import (
	"../geo"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

var version = "0.6-beta"

func Version(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, version)
}

func Status(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "running")
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome!")
}

func Tile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	z, _ := strconv.Atoi(vars["zoom"])
	x, _ := strconv.Atoi(vars["x"])
	y, _ := strconv.Atoi(vars["y"])
	fmt.Fprintln(w, "Generating file zoom :", vars["zoom"], "x,y :", vars["x"], ",", vars["y"])
	tile := geo.NewTileFromZXY(z, x, y)
	fmt.Fprintf(w, "Resulting tile : %v", tile)
}

/*
func Capabilities(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "//TOIMPLEMENT")
}

func Map(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "//TOIMPLEMENT")
}
*/
