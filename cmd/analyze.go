package cmd

import (
	"github.com/codegangsta/cli"
	"log"
	"runtime"

	"github.com/adentes-org/KISSTile/modules/modules/db"
)

type BoolFlag struct {
	Name   string
	Usage  string
	EnvVar string
}

var CmdAnalyze = cli.Command{
	Name:        "analyze",
	Usage:       "Run the pre-analyze of the ospbf",
	Description: "This is run by the serve before running if it wasn't done before",
	Action:      runAnalyze,
	Flags: []cli.Flag{
		cli.BoolFlag{"force-rescan, Rf", "Clean index and descriptor and force re-scan //TOIMPLEMENT", ""},
	},
}

func runAnalyze(ctx *cli.Context) {
	//TODO check for args and no default
	file := "../isle-of-man-latest.osm.pbf"
	if len(ctx.Args()) > 0 {
		file = ctx.Args()[0]
	}

	Analyze(file)
}

func Analyze(file string) *db.Db {

	log.Printf("Analyzing file : %s", file)

	db, _ := db.OpenDB(file)

	log.Printf("NbProc : %d", runtime.GOMAXPROCS(-1))

	//	db.Analyze(runtime.GOMAXPROCS(-1))
	db.Analyze(runtime.GOMAXPROCS(-1))

	return db
	//db.Close()
	//log.Printf("Result : %s %s", db, err)
}
