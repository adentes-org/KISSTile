package geo

import (
	"code.google.com/p/draw2d/draw2d"
	"errors"
	"github.com/sapk/osmpbf"
	"image"
	"image/color"
	"image/draw"
	//"log"
	"math"
)

type Tile struct {
	Z   int
	X   int
	Y   int
	Lat float64
	Lon float64
}

const lat_panel = 85.0511 * 2

//TODO  do not draw to small BBox compare to too bigger and not bigger once
//const lat_panel = 90 * 2
const lon_panel = 180 * 2

const tileSize = 256

type Conversion interface {
	deg2num(t *Tile) (x int, y int)
	num2deg(t *Tile) (lat float64, lon float64)
}

func NewTileFromZoomLatLon(zoom int, lat float64, lon float64) *Tile {
	var t Tile
	// On verifie que l'on est bien sur terre (360 + 85.0.511)
	t.Lat = lat
	t.Lon = lon
	t.Z = zoom
	t.X, t.Y = Deg2num(&t)
	return &t
}
func NewTileFromZXY(z int, x int, y int) *Tile {
	var t Tile
	t.Z = z
	nbtile := NbAtZLevel(z)
	t.X = x % int(nbtile)
	t.Y = y % int(nbtile)
	t.Lat, t.Lon = Num2deg(&t)
	return &t
}

//func (*Tile) Deg2num(t *Tile) (x int, y int) {
func Deg2num(t *Tile) (x int, y int) {
	x = int(math.Floor((t.Lon + 180.0) / 360.0 * (math.Exp2(float64(t.Z)))))
	y = int(math.Floor((1.0 - math.Log(math.Tan(t.Lat*math.Pi/180.0)+1.0/math.Cos(t.Lat*math.Pi/180.0))/math.Pi) / 2.0 * (math.Exp2(float64(t.Z)))))
	return x, y
}

//func (*Tile) Num2deg(t *Tile) (lat float64, lon float64) {
func Num2deg(t *Tile) (lat float64, lon float64) {
	n := math.Exp2(float64(t.Z))
	lat = 180.0 / math.Pi * math.Atan(math.Sinh(math.Pi*(1-2*float64(t.Y)/n)))
	lon = float64(t.X)/n*360.0 - 180.0
	return lat, lon
}

func NbAtZLevel(zoom int) float64 {
	//number of tile
	//TODO cache the result in struct of tile
	return math.Exp2(float64(zoom))
}
func PrecisionAtZLevel(zoom int) (lat float64, lon float64) {
	//number of tile
	//TODO cache the result in struct of tile
	n := NbAtZLevel(zoom)
	//Precision
	lat = lat_panel / n
	lon = lon_panel / n
	//TODO why ?
	return lat * math.Pi / 2, lon
}

func (t *Tile) GetBBOX() Bbox {
	//TODO cache the result in struct of tile
	p_lat, p_lon := PrecisionAtZLevel(t.Z)
	bb := Bbox{Point{t.Lat - p_lat, t.Lon}, Point{t.Lat, t.Lon + p_lon}}
	//bb.OrderBbox()
	return bb

}

func (t *Tile) DrawTile(ways map[int64]*osmpbf.Way, nodes map[int64]*osmpbf.Node, debug bool) (image.Image, error) {
	//TODO plot unique point
	// Create white image
	img := image.NewRGBA(image.Rect(0, 0, tileSize, tileSize))
	draw.Draw(img, img.Bounds(), image.White, image.ZP, draw.Src)
	plat, plon := PrecisionAtZLevel(t.Z)

	// Plot ways
	for _, way := range ways {
		/*
			if mapFeatures[fName].MinZoom > zoom {
				continue
			}
		*/
		//*
		//TODO filter before
		if val, ok := way.Tags["natural"]; !ok {
			//FORTEST if it'nst a natural we pass
			if val != "coastline" {
				continue
			}
		}
		//*/
		coords := [][]float64{}
		for _, nodeID := range way.NodeIDs {
			//TODO relative path ° to px
			if _, ok := nodes[nodeID]; !ok {
				return nil, errors.New("node not found")
			}
			coords = append(coords, []float64{(nodes[nodeID].Lon - t.Lon) * tileSize / plon, (t.Lat - nodes[nodeID].Lat) * tileSize / plat})
		}
		drawPolyLine(img, color.Black, coords)
		//	log.Printf("way : %v", way)
		//	log.Printf("coords : %v", coords)
	}
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
