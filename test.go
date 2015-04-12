// test
package main

import (
	"fmt"
	"github.com/ajstarks/svgo"
	"github.com/sapk/osmpbf"
	"io"
	"log"
	"os"
	"runtime"
	"time"
)

func main2() {
	fmt.Println("Hello World!")
	fmt.Println(time.Now())
	fmt.Println(len(os.Args), os.Args)

	min := [2]float64{48.964338, 2.037728} //Top left
	max := [2]float64{48.994092, 2.073938} //Bottom Right
	//max = [2]float64{v.Lat, v.Lon}

	f, err := os.Open("Téléchargements/ile-de-france-latest.osm.pbf")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	d := osmpbf.NewDecoder(f)
	err = d.Start(runtime.GOMAXPROCS(-1)) // use several goroutines for faster decoding
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Start %v\n", time.Now())

	//var nodes map[int64][2]float64
	//*
	var nodes = make(map[int64][2]float64)
	var nc, wc, rc uint64
	var cnc, cwc, crc uint64

	for {
		if v, _, err := d.Decode(); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		} else {
			switch v := v.(type) {
			case *osmpbf.Node:
				// Process Node v.
				if v.Lat > min[0] && v.Lat < max[0] && v.Lon > min[1] && v.Lon < max[1] {
					nodes[v.ID] = [2]float64{v.Lat, v.Lon}
					cnc++
				}

				nc++
			case *osmpbf.Way:
				// Process Way v.
				//if v.Tags["highway"] == "residential" {
				//	fmt.Printf("%v", v.Tags)
				//}
				//fmt.Printf("\n%v\n", v)
				//fmt.Println(v.Tags)
				//if val, ok := nodes["foo"]; ok {				}

				for _, node := range v.NodeIDs {
					if _, ok := nodes[node]; ok {
						cwc++
					}
				}
				wc++
				break
			case *osmpbf.Relation:
				// Process Relation v.
				rc++
			default:
				log.Fatalf("unknown type %T\n", v)
			}
			if (nc+wc+rc)%1000000 == 0 {
				fmt.Printf("Nodes: %dk, Ways: %dk, Relations: %dk @", nc/1000, wc/1000, rc/1000)
				fmt.Println(time.Now())
			}
		}
	}
	fmt.Printf("Nodes: %d, Ways: %d, Relations: %d\n", nc, wc, rc)
	fmt.Printf("Nodes: %d, Ways: %d, Relations: %d\n", cnc, cwc, crc)

	//*/
	fmt.Println(time.Now())
	//MakeSVG()
}

func MakeSVG() {
	width := 500
	height := 500
	canvas := svg.New(os.Stdout)
	canvas.Start(width, height)
	canvas.Circle(width/2, height/2, 100)
	canvas.Text(width/2, height/2, "Hello, SVG", "text-anchor:middle;font-size:30px;fill:white")
	canvas.End()
}
