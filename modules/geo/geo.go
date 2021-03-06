// geo

package geo

import (
	"math"
)

type Point struct {
	Lon, Lat float64
}

//TODO pass to float32 point in order to save space for index
// 0 : SW 1 : NE
type Bbox [2]Point

func (bb *Bbox) IntersectWith(bb_inter Bbox) bool {
	return bb[0].InBbox(&bb_inter) || bb[1].InBbox(&bb_inter) || bb_inter[0].InBbox(bb) || bb_inter[1].InBbox(bb)
}

//Represente distance of the bbox in order to filter
func (bb *Bbox) Size() float64 {
	return math.Sqrt(math.Pow(bb[1].Lat-bb[0].Lat, 2) + math.Pow(bb[1].Lon-bb[0].Lon, 2))
}
func (bb *Bbox) OrderBbox() {
	// 0 < 1 in order to have [SW,NE]
	if bb[0].Lon > bb[1].Lon {
		tmp := bb[0].Lon
		bb[0].Lon = bb[1].Lon
		bb[1].Lon = tmp
	}
	if bb[0].Lat > bb[1].Lat {
		tmp := bb[0].Lat
		bb[0].Lat = bb[1].Lat
		bb[1].Lat = tmp
	}
}

func (bb *Bbox) AddInnerPoint(p Point) {
	if !p.InBbox(bb) {
		//We need to enlarge the Bbox
		bb[0].Lat = min(p.Lat, bb[0].Lat)
		bb[1].Lat = max(p.Lat, bb[1].Lat)
		bb[0].Lon = min(p.Lon, bb[0].Lon)
		bb[1].Lon = max(p.Lon, bb[1].Lon)
	}
}
func (p *Point) InBbox(bb *Bbox) bool {
	return p.Lon >= bb[0].Lon && p.Lon <= bb[1].Lon && p.Lat >= bb[0].Lat && p.Lat <= bb[1].Lat
}

/*
func BboxFromZXY(z int, x int, y int) Bbox {
	return Bbox{Point{}, Point{}}
}
*/
func min(a, b float64) float64 {
	if a <= b {
		return a
	}
	return b
}
func max(a, b float64) float64 {
	if a >= b {
		return a
	}
	return b
}
