package cmd

import (
	"./../modules/api"
	"github.com/codegangsta/cli"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

var CmdServer = cli.Command{
	Name:        "server",
	Usage:       "Start KISSTle web server",
	Description: "The only thing you need to be ready, it will take care of all the other things for you",
	Action:      runServer,
	Flags: []cli.Flag{
		cli.StringFlag{"port, p", "80", "Port number to listen for the server", ""},
		cli.StringFlag{"cache, c", "tiles/", "Folder used to cached all the tiles already generated", ""},
	},
}

func runServer(ctx *cli.Context) {
	//TODO check for args and no default
	file := "../isle-of-man-latest.osm.pbf"
	if len(ctx.Args()) > 0 {
		file = ctx.Args()[0]
	}

	db := Analyze(file)
	log.Println("Database ready !")
	ap := api.Init(db)

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", Index)
	router.HandleFunc("/api", ap.Index)
	router.HandleFunc("/api/version", ap.Version)
	router.HandleFunc("/api/status", ap.Version)
	/*
		router.HandleFunc("/api/{apiVersion}/map", api.Map)
		router.HandleFunc("/api/{apiVersion}/permissions", api.Permissions)
		//NOTPLANNED router.HandleFunc("/api/{apiVersion}/changeset", api.Changeset)
		router.HandleFunc("/api/{apiVersion}/trackpoints", api.Trackpoints)
	*/
	router.HandleFunc("/api/tile/{zoom:[0-1]?[0-9]}/{x:[0-9]+}/{y:[0-9]+}", ap.Tile)
	router.HandleFunc("/api/tile/{zoom:[0-1]?[0-9]}/{x:[0-9]+}/{y:[0-9]+}.{format}", ap.Tile)
	serv := []string{"localhost", ctx.String("port")}
	log.Fatal(http.ListenAndServe(strings.Join(serv, ":"), router))
	log.Printf("Listening @ %s", strings.Join(serv, ":"))
}

func Index(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprintln(w, "Welcome!")
	log.Println("Someone get index page !")
	b, _ := ioutil.ReadFile("web/index.html")
	w.Write(b)

}
