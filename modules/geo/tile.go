package geo

import (
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
const lon_panel = 180 * 2

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
	return lat, lon
}

func (t *Tile) GetBBOX() Bbox {
	//TODO cache the result in struct of tile
	p_lat, p_lon := PrecisionAtZLevel(t.Z)
	return Bbox{Point{t.Lat, t.Lon}, Point{t.Lat - p_lat, t.Lon + p_lon}}

}
