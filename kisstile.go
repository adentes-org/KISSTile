package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
)

type Config struct {
	bbox      [][]float64
	zoom      []uint8
	cacheSize int
}

func main() {
	fmt.Println("Starting!")
	/*
		var jsonConfig = []byte(`
			{"bbox": "Platypus", "zoom": [0,19], "cacheSize": 5000000}
			`)
	*/
	var jsonConfig = []byte(`
		{"bbox":[[48.964338,2.037728],[48.994092,2.073938]], "zoom": [0,0], "cacheSize": 5000}
		`)

	var config Config
	err := json.Unmarshal(jsonConfig, &config)
	if err != nil {
		fmt.Println("error:", err)
	}
	db, err := OpenDB("../../Téléchargements/ile-de-france-latest.osm.pbf", config.cacheSize)
	//db, err := DB.Open("../../Téléchargements/isle-of-man-latest.osm.pbf", config.cacheSize)

	log.Println("File opened!")

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Db starting ...")
	err = db.Start(runtime.GOMAXPROCS(-1)) // use several goroutines for faster decoding
	//2015/03/30 21:49:35 - 21:50:36 {%!s(int64=138) %!s(int64=21195685) %!s(int64=124459463) %!s(int64=3456192) %!s(int64=217390678) %!s(int64=36199)}
	//err = db.Start(runtime.GOMAXPROCS(1)) // use several goroutines for faster decoding
	//2015/03/30 21:47:19 - 21:48:19 {%!s(int64=138) %!s(int64=21195685) %!s(int64=124459463) %!s(int64=3456192) %!s(int64=217390678) %!s(int64=36199)}

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Scan starting ...")
	_, err = db.Scan()

	if err != nil {
		log.Fatal(err)
	}
	log.Println("Scan ended ...")

	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		port = 8080
	}

	tileHandler := NewTileHandler("/tiles", db)

	http.Handle("/tiles/", tileHandler)
	//	http.Handle("/", http.FileServer(http.Dir("public")))
	fmt.Printf("Listening on http://0.0.0.0:%d/\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
	/*
		bbox := [2][2]float64{{2.3087596893, 48.8680882999}, {2.3115706444, 48.8698949548}}

		DrawTile(bbox, 0, 0, db, true)
		DrawTile(bbox, 1, 0, db, true)
		DrawTile(bbox, 2, 0, db, true)
		DrawTile(bbox, 3, 0, db, true)
		DrawTile(bbox, 4, 0, db, true)
		DrawTile(bbox, 5, 0, db, true)
		DrawTile(bbox, 6, 0, db, true)
	*/
	/*
		var id int64
		id = 122626
		log.Printf("Searching node %d ... \n", id)
		node, err := db.GetNode(id)

		if err != nil {
			log.Fatal(err)
		}

		log.Printf("%s", node)

		// Know ways
		//ile  [0] : 27022294 | [1] : 103966553 112574188 112573980 [2] 314717968 314756815
		//ide de france [426] 315706792
		id = 315706792
		log.Printf("Searching way %d ... \n", id)
		way, err := db.GetWay(id)

		if err != nil {
			log.Fatal(err)
		}

		log.Printf("%s", way)

		log.Printf("Searching nodes of way %d ... \n", id)
		nodes, err := db.GetNodesOfWay(way)

		if err != nil {
			log.Fatal(err)
		}

		log.Printf("%s", nodes)

		//bbox := [2][2]float64{{48.859426, 2.288815}, {48.853609, 2.303685}}
		bbox := [2][2]float64{{48.8680882999, 2.3087596893}, {48.8698949548, 2.3115706444}}

		log.Printf("Searching nodes in bbox %s ... \n", bbox)
		nodesinbbox, err := db.GetNodesIn(bbox)

		if err != nil {
			log.Fatal(err)
		}

		log.Printf("%s", nodesinbbox)

		log.Printf("Searching ways containing nodes in bbox %s ... \n", bbox)
		wayscontaining, err := db.GetWaysContaining(nodesinbbox)

		if err != nil {
			log.Fatal(err)
		}

		log.Printf("%s", wayscontaining)
	*/
}
