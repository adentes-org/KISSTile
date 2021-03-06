package main

import (
	"os"
	"runtime"

	"github.com/codegangsta/cli"

	"github.com/adentes-org/KISSTile/cmd"
)

const APP_VER = "0.0.1.0001 Alpha"

func main() {
	//TODO
	//TODO use mbrickade bbox float32 for date and relative inside
	runtime.GOMAXPROCS(4)

	app := cli.NewApp()
	app.Name = "KISSTile"
	app.Usage = "Deliver OpenStreetMap the simplest way the file given in argument"
	app.Version = APP_VER
	app.Commands = []cli.Command{
		cmd.CmdAnalyze,
		cmd.CmdGenerate,
		cmd.CmdServer,
		cmd.CmdClean,
	}
	app.Flags = append(app.Flags, []cli.Flag{}...)
	app.Run(os.Args)

}
