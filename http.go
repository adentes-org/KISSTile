package main

import (
	"image/png"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type TileHandler struct {
	prefix string
	db     *Db
}

// prefix should be of the form "/tiles" (without the trailing slash)
func NewTileHandler(prefix string, db *Db) *TileHandler {
	return &TileHandler{
		prefix: prefix,
		db:     db,
	}
}

func (th *TileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len(th.prefix):]

	if !(strings.HasPrefix(path, "/") && strings.HasSuffix(path, ".png")) {
		w.Write([]byte("404"))
		return
	}

	xyz := strings.Split(path[1:len(path)-4], "/")
	if len(xyz) != 3 {
		w.Write([]byte("404"))
		return
	}

	xyz_ := []int64{}
	for _, value := range xyz {
		intVal, err := strconv.Atoi(value)
		if err != nil {
			w.Write([]byte("404"))
			return
		}
		xyz_ = append(xyz_, int64(intVal))
	}
	zoom := xyz_[0]
	x := xyz_[1]
	y := xyz_[2]
	nwPt := getLonLatFromTileName(x, y, zoom)
	sePt := getLonLatFromTileName(x+1, y+1, zoom)
	//img, err := DrawTile(nwPt, sePt, zoom, th.font, th.data, debug)
	bbox := [2][2]float64{{nwPt.Lon_, nwPt.Lat_}, {sePt.Lon_, sePt.Lat_}}
	log.Printf("Searching nodes in bbox %s ... \n", bbox)

	img, err := DrawTile(bbox, zoom, 0, th.db, true)
	if err != nil {
		panic(err)
	}
	// Ignore broken pipe errors
	png.Encode(w, img)
}
