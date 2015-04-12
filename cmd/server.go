package cmd

import (
	"github.com/codegangsta/cli"
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

}
