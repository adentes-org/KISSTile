package cmd

import (
	"github.com/codegangsta/cli"
	"log"
	"runtime"

	"./../modules/db"
)

var CmdAnalyze = cli.Command{
	Name:        "analyze",
	Usage:       "Run the pre-analyze of the ospbf",
	Description: "This is run by the serve before running if it wasn't done before",
	Action:      runAnalyze,
}

func runAnalyze(ctx *cli.Context) {
	//TODO check for args and no default
	file := "../isle-of-man-latest.osm.pbf"
	if len(ctx.Args()) > 0 {
		file = ctx.Args()[0]
	}

	log.Printf("Analyzing file : %s", file)

	db, _ := db.OpenDB(file)

	log.Printf("NbProc : %d", runtime.GOMAXPROCS(-1))

	db.Analyze(runtime.GOMAXPROCS(-1))

	//log.Printf("Result : %s %s", db, err)
}
