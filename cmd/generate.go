package cmd

import (
	"github.com/codegangsta/cli"
)

var CmdGenerate = cli.Command{
	Name:   "generate",
	Usage:  "Generate all the tiles possibles from the ospbf file",
	Action: runGenerate,
	Flags: []cli.Flag{
		cli.StringFlag{"output-folder, O", "tiles/", "Folder where the tiles a saved", ""},
		cli.StringFlag{"bbox", "[]", "Permit to limit the boundaries of the tiles generated", ""},
	},
}

func runGenerate(ctx *cli.Context) {

}
