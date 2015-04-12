package cmd

import (
	"github.com/codegangsta/cli"
)

var CmdClean = cli.Command{
	Name:   "clean",
	Usage:  "Clean index,tiles and caches generated from previous usages",
	Action: runClean,
	Flags: []cli.Flag{
		cli.StringFlag{"cache, c", "tiles/", "Folder where the tiles a saved", ""},
	},
}

func runClean(ctx *cli.Context) {

}
