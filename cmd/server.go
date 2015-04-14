package cmd

import (
	"./../modules/api"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/gorilla/mux"
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
	runAnalyze(ctx)
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", Index)
	router.HandleFunc("/api", api.Index)
	router.HandleFunc("/api/version", api.Version)
	router.HandleFunc("/api/status", api.Version)
	/*
		router.HandleFunc("/api/{apiVersion}/map", api.Map)
		router.HandleFunc("/api/{apiVersion}/permissions", api.Permissions)
		//NOTPLANNED router.HandleFunc("/api/{apiVersion}/changeset", api.Changeset)
		router.HandleFunc("/api/{apiVersion}/trackpoints", api.Trackpoints)
	*/
	router.HandleFunc("/api/tile/{zoom:[0-1]?[0-9]}/{x:[0-9]+}/{y:[0-9]+}", api.Tile)
	router.HandleFunc("/api/tile/{zoom:[0-1]?[0-9]}/{x:[0-9]+}/{y:[0-9]+}.{format}", api.Tile)
	serv := []string{"localhost", ctx.String("port")}
	log.Fatal(http.ListenAndServe(strings.Join(serv, ":"), router))
	log.Printf("Listening @ %s", strings.Join(serv, ":"))
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome!")
}
