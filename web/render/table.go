package render

import (
	"golang.org/x/image/font"
)

type Node struct {
	Group    string
	Remarks  string
	Ping     string
	AvgSpeed string
	MaxSpeed string
}

type Nodes []Node

type TableOptions struct {
}

type CellWidths struct {
	group    float64
	remarks  float64
	ping     float64
	avgspeed float64
	maxspeed float64
}

type Table struct {
	width  int
	height int
	*Context
	nodes      Nodes
	options    TableOptions
	cellWidths *CellWidths
}

func NewTable(width int, height int) Table {
	dc := NewContext(width, height)
	return Table{
		width:   width,
		height:  height,
		Context: dc,
	}
}

func (t *Table) calcWidths(fontface font.Face) {
	if len(t.nodes) < 1 {
		return
	}
	cellWidths := CellWidths{}
	cellWidths.group = getWidth(fontface, t.nodes[0].Group)
	cellWidths.ping = getWidth(fontface, t.nodes[0].Ping)
	t.cellWidths = &cellWidths
}

func getWidth(fontface font.Face, text string) float64 {
	var totalWidth float64 = 0
	for _, r := range text {
		awidth, _ := fontface.GlyphAdvance(r)
		iwidthf := float64(awidth) / 64
		// fmt.Printf("%.2f\n", iwidthf)
		totalWidth += float64(iwidthf)
	}
	return totalWidth
}
