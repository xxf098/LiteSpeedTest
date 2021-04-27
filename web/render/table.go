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
	horizontalpadding float64 // left + right
	verticalpadding   float64 // up + down
	tableTopPadding   float64 // padding for table
	lineWidth         float64
	fontHeight        float64
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

func NewTable(width int, height int, options TableOptions) Table {
	dc := NewContext(width, height)
	return Table{
		width:   width,
		height:  height,
		Context: dc,
		options: options,
	}
}

func (t *Table) drawHorizonLines() {
	y := t.options.fontHeight + t.options.tableTopPadding
	for i := 0; i < len(t.nodes)+4; i++ {
		t.drawHorizonLine(y - t.options.fontHeight)
		y = y + t.options.fontHeight + t.options.verticalpadding
	}
}

func (t *Table) drawHorizonLine(y float64) {
	t.DrawLine(0, y, float64(t.width), y)
	t.SetLineWidth(t.options.lineWidth)
	t.Stroke()
}

func (t *Table) drawVerticalLines() {
	padding := t.options.verticalpadding
	x := t.cellWidths.group + padding
	t.drawVerticalLine(x)
	x += t.cellWidths.remarks + padding
	t.drawVerticalLine(x)
	x += t.cellWidths.ping + padding
	t.drawVerticalLine(x)
	x += t.cellWidths.avgspeed + padding
	t.drawVerticalLine(x)
}

func (t *Table) drawVerticalLine(x float64) {
	t.DrawLine(x, t.options.tableTopPadding, x, float64(t.height)-15)
	t.SetLineWidth(0.5)
	t.Stroke()
}

func (t *Table) drawNodes() {
	horizontalpadding := t.options.horizontalpadding
	var x float64 = horizontalpadding / 2
	var y float64 = t.options.fontHeight + t.options.verticalpadding/2 + t.options.tableTopPadding
	for _, v := range t.nodes {
		t.DrawString(v.Group, x, y)
		x += t.cellWidths.group + horizontalpadding
		t.DrawString(v.Remarks, x, y)
		x += t.cellWidths.remarks + horizontalpadding
		t.DrawString(v.Ping, x, y)
		x += t.cellWidths.ping + horizontalpadding
		t.DrawString(v.AvgSpeed, x, y)
		x += t.cellWidths.avgspeed + horizontalpadding
		t.DrawString(v.MaxSpeed, x, y)
		y = y + t.options.fontHeight + t.options.verticalpadding
		x = horizontalpadding / 2
	}
}

func (t *Table) draw() error {
	t.drawHorizonLines()
	return nil
}

func calcWidth(fontface font.Face, nodes Nodes) *CellWidths {
	cellWidths := &CellWidths{}
	if len(nodes) < 1 {
		return cellWidths
	}
	cellWidths.group = getWidth(fontface, nodes[0].Group)
	for _, v := range nodes {
		width := getWidth(fontface, v.Ping)
		if cellWidths.ping < width {
			cellWidths.ping = width
		}
		width = getWidth(fontface, v.AvgSpeed)
		if cellWidths.avgspeed < width {
			cellWidths.avgspeed = width
		}
		width = getWidth(fontface, v.MaxSpeed)
		if cellWidths.maxspeed < width {
			cellWidths.maxspeed = width
		}
		width = getWidth(fontface, v.Remarks)
		if cellWidths.remarks < width {
			cellWidths.remarks = width
		}
	}
	return cellWidths
}

func calcHeight(fontface font.Face) float64 {
	return float64(fontface.Metrics().Height) / 64
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
