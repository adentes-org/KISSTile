package main

import (
	"code.google.com/p/draw2d/draw2d"
	"log"
	//	"code.google.com/p/freetype-go/freetype"
	//	"code.google.com/p/freetype-go/freetype/truetype"
	//"fmt"
	"image"
	"image/color"
	"image/draw"
	//"image/jpeg"
	//"os"
	//	"math"
	//"strconv"
	"time"
)

const tileSize = 256

func DrawTile(bbox [2][2]float64, zoom int64, font int, db *Db, debug bool) (image.Image, error) {
	var err error

	t := time.Now()

	//get good order for bbox
	bbox = formatBBOX(bbox)
	// Create white image
	img := image.NewRGBA(image.Rect(0, 0, tileSize, tileSize))
	draw.Draw(img, img.Bounds(), image.White, image.ZP, draw.Src)

	if debug {
		log.Printf("Searching nodes in bbox %s ... \n", bbox)
	}
	nodes, err := db.GetNodesIn(bbox)

	if err != nil {
		return nil, err
	}

	if debug {
		log.Printf("Found : %d nodes \n", len(nodes))
		log.Printf("Searching ways containing nodes in bbox %s ... \n", bbox)
	}

	ways, err := db.GetWaysContaining(nodes)

	if err != nil {
		return nil, err
	}

	if debug {
		log.Printf("Found : %d ways \n", len(ways))
	}
	//Get missing node outside bbox to get direction of last point

	var nodes_missing int
	for _, way := range ways {
		for _, nid := range way.NodeIDs {
			if _, ok := nodes[nid]; !ok {
				nodes[nid], err = db.GetNode(nid)
				if err != nil {
					return nil, err
				}
				nodes_missing++
			}
		}
	}

	if debug {
		log.Printf("Found : %d nodes missing \n", nodes_missing)
		log.Printf("Generating img ...\n")
	}
	coords := [][]float64{}
	for _, node := range nodes {
		x, y := getRelativeXY(Point{bbox[0][0], bbox[0][1]}, Point{node.Lat, node.Lon}, float64(zoom))
		coords = append(coords, []float64{x, y})
	}
	drawPolyLine(img, color.Black, coords)
	/*
		if debug {
			log.Printf("Saving img ...\n")
		}

		out, err := os.Create("./output" + strconv.Itoa(int(zoom)) + ".jpg")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		var opt jpeg.Options

		opt.Quality = 80
		// ok, write out the data into the new JPEG file

		err = jpeg.Encode(out, img, &opt) // put quality to 80%
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if debug {
			log.Printf("It takes : %s\n", time.Since(t).String())
		}
	*/
	log.Printf("It takes : %s\n", time.Since(t).String())
	return img, nil
}

func drawPolyLine(img *image.RGBA, color color.Color, coords [][]float64) {
	path := draw2d.NewPathStorage()
	for i, coord := range coords {
		if i == 0 {
			path.MoveTo(coord[0], coord[1])
		} else {
			path.LineTo(coord[0], coord[1])
		}
	}
	gc := draw2d.NewGraphicContext(img)
	gc.SetStrokeColor(color)
	gc.Stroke(path)
}
