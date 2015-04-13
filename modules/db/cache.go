package db

import (
	"github.com/sapk/osmpbf"
)

type NodeCache struct {
	Nodes map[int64]*osmpbf.Node
	Size  int
}

func CreateNodeCache(size int) *NodeCache {
	return &NodeCache{
		Nodes: make(map[int64]*osmpbf.Node),
		Size:  size,
	}
}
func (this *NodeCache) get(id int64) (*osmpbf.Node, bool) {
	val, err := this.Nodes[id]
	return val, err
}

func (this *NodeCache) add(node *osmpbf.Node) {
	if len(this.Nodes) > this.Size {
		i := this.Size / 1000 //On enleve un pour mille du cache
		for id := range this.Nodes {
			delete(this.Nodes, id)
			i--
			if i < 0 {
				break
			}
		}
	}
	this.Nodes[node.ID] = node
}
