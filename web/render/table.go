package render

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/xxf098/lite-proxy/download"
	"golang.org/x/image/font"
)

var colorgroup = [][]int{
	{255, 255, 255},
	{102, 255, 102},
	{255, 255, 102},
	{255, 178, 102},
	{255, 102, 102},
	{226, 140, 255},
	{102, 204, 255},
	{102, 102, 255},
}

var bounds = []int{0, 64 * 1024, 512 * 1024, 4 * 1024 * 1024, 16 * 1024 * 1024, 24 * 1024 * 1024, 32 * 1024 * 1024, 40 * 1024 * 1024}

type Node struct {
	Group    string
	Remarks  string
	Protocol string
	Ping     string
	AvgSpeed int64
	MaxSpeed int64
}

type Nodes []Node

func CSV2Nodes(path string) (Nodes, error) {
	recordFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer recordFile.Close()
	reader := csv.NewReader(recordFile)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	nodes := make(Nodes, len(records))
	for i, v := range records {
		if len(v) < 6 {
			continue
		}
		avg, err := strconv.Atoi(v[4])
		if err != nil {
			continue
		}
		max, err := strconv.Atoi(v[5])
		if err != nil {
			continue
		}
		nodes[i] = Node{
			Group:    v[0],
			Remarks:  v[1],
			Protocol: v[2],
			Ping:     v[3],
			AvgSpeed: int64(avg),
			MaxSpeed: int64(max),
		}
	}
	return nodes, nil
}

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
	protocol float64
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

func DefaultTable(nodes Nodes, fontPath string) (*Table, error) {
	fontSize := 22
	fontface, err := LoadFontFace(fontPath, float64(fontSize))
	if err != nil {
		return nil, err
	}
	widths := calcWidth(fontface, nodes)
	fontHeight := calcHeight(fontface)
	var horizontalpadding float64 = 40
	tableWidth := widths.group + horizontalpadding + widths.remarks + horizontalpadding + widths.protocol + horizontalpadding + widths.ping + horizontalpadding + widths.avgspeed + horizontalpadding + widths.maxspeed + horizontalpadding
	options := TableOptions{
		horizontalpadding: horizontalpadding,
		verticalpadding:   30,
		tableTopPadding:   20,
		lineWidth:         0.5,
		fontHeight:        fontHeight,
	}
	tableHeight := (fontHeight+options.verticalpadding)*float64((len(nodes)+4)) + options.tableTopPadding*2
	table := NewTable(int(tableWidth), int(tableHeight), options)
	table.nodes = nodes
	table.cellWidths = widths
	table.SetFontFace(fontface)
	return &table, nil
}

func (t *Table) drawHorizonLines() {
	y := t.options.fontHeight + t.options.tableTopPadding
	for i := 0; i <= len(t.nodes)+4; i++ {
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
	padding := t.options.horizontalpadding
	x := t.cellWidths.group + padding
	t.drawVerticalLine(x)
	x += t.cellWidths.remarks + padding
	t.drawVerticalLine(x)
	x += t.cellWidths.protocol + padding
	t.drawVerticalLine(x)
	x += t.cellWidths.ping + padding
	t.drawVerticalLine(x)
	x += t.cellWidths.avgspeed + padding
	t.drawVerticalLine(x)
}

func (t *Table) drawVerticalLine(x float64) {
	height := (t.options.fontHeight+t.options.verticalpadding)*float64((len(t.nodes)+2)) + t.options.tableTopPadding
	y := t.options.tableTopPadding + t.options.fontHeight + t.options.verticalpadding
	t.DrawLine(x, y, x, height)
	t.SetLineWidth(t.options.lineWidth)
	t.Stroke()
}

func (t *Table) drawTitle() {
	// horizontalpadding := t.options.horizontalpadding
	title := "LiteSpeedTest Result Table"
	var x float64 = float64(t.width)/2 - getWidth(t.fontFace, title)/2
	var y float64 = t.options.fontHeight + t.options.verticalpadding/2 + t.options.tableTopPadding
	t.DrawString(title, x, y)
}

func (t *Table) drawHeader() {
	horizontalpadding := t.options.horizontalpadding
	var x float64 = horizontalpadding / 2
	var y float64 = t.options.fontHeight + t.options.verticalpadding/2 + t.options.tableTopPadding + t.options.fontHeight + t.options.verticalpadding
	adjust := t.cellWidths.group/2 - getWidth(t.fontFace, "Group")/2
	t.DrawString("Group", x+adjust, y)
	x += t.cellWidths.group + horizontalpadding
	adjust = t.cellWidths.remarks/2 - getWidth(t.fontFace, "Remarks")/2
	t.DrawString("Remarks", x+adjust, y)
	x += t.cellWidths.remarks + horizontalpadding
	t.DrawString("Protocol", x, y)
	x += t.cellWidths.protocol + horizontalpadding
	adjust = t.cellWidths.ping/2 - getWidth(t.fontFace, "Ping")/2
	t.DrawString("Ping", x+adjust, y)
	x += t.cellWidths.ping + horizontalpadding
	t.DrawString("AvgSpeed", x, y)
	x += t.cellWidths.avgspeed + horizontalpadding
	t.DrawString("MaxSpeed", x, y)
}

func (t *Table) drawTraffic(traffic string) {
	// horizontalpadding := t.options.horizontalpadding
	var x float64 = t.options.horizontalpadding / 2
	var y float64 = (t.options.fontHeight+t.options.verticalpadding)*float64((len(t.nodes)+2)) + t.options.tableTopPadding + t.fontHeight + t.options.verticalpadding/2
	t.DrawString(traffic, x, y)
}

func (t *Table) drawGeneratedAt() {
	// horizontalpadding := t.options.horizontalpadding
	msg := fmt.Sprintf("Generated at %s", time.Now().Format(time.RFC3339))
	var x float64 = t.options.horizontalpadding / 2
	var y float64 = (t.options.fontHeight+t.options.verticalpadding)*float64((len(t.nodes)+3)) + t.options.tableTopPadding + t.fontHeight + t.options.verticalpadding/2
	t.DrawString(msg, x, y)
}

func (t *Table) drawNodes() {
	horizontalpadding := t.options.horizontalpadding
	var x float64 = horizontalpadding / 2
	var y float64 = t.options.fontHeight + t.options.verticalpadding/2 + t.options.tableTopPadding + (t.options.fontHeight+t.options.verticalpadding)*2
	for _, v := range t.nodes {
		t.DrawString(v.Group, x, y)
		x += t.cellWidths.group + horizontalpadding
		t.DrawString(v.Remarks, x, y)
		x += t.cellWidths.remarks + horizontalpadding
		adjust := t.cellWidths.protocol/2 - getWidth(t.fontFace, v.Protocol)/2
		t.DrawString(v.Protocol, x+adjust, y)
		x += t.cellWidths.protocol + horizontalpadding
		adjust = t.cellWidths.ping/2 - getWidth(t.fontFace, v.Ping)/2
		t.DrawString(v.Ping, x+adjust, y)
		x += t.cellWidths.ping + horizontalpadding
		avgSpeed := download.ByteCountIECTrim(v.AvgSpeed)
		adjust = t.cellWidths.avgspeed/2 - getWidth(t.fontFace, avgSpeed)/2
		t.DrawString(avgSpeed, x+adjust, y)
		x += t.cellWidths.avgspeed + horizontalpadding
		maxSpeed := download.ByteCountIECTrim(v.MaxSpeed)
		adjust = t.cellWidths.maxspeed/2 - getWidth(t.fontFace, maxSpeed)/2
		t.DrawString(maxSpeed, x+adjust, y)
		y = y + t.options.fontHeight + t.options.verticalpadding
		x = horizontalpadding / 2
	}
}

func (t *Table) drawSpeed() {
	padding := t.options.horizontalpadding
	var lineWidth float64 = t.options.lineWidth
	var x1 float64 = t.cellWidths.group + padding + t.cellWidths.remarks + padding + t.cellWidths.protocol + padding + t.cellWidths.ping + padding + lineWidth
	var x2 float64 = t.cellWidths.group + padding + t.cellWidths.remarks + padding + t.cellWidths.protocol + padding + t.cellWidths.ping + padding + t.cellWidths.avgspeed + padding + lineWidth
	var y float64 = t.options.tableTopPadding + lineWidth + (t.options.fontHeight+t.options.verticalpadding)*2
	var wAvg float64 = t.cellWidths.avgspeed + padding - lineWidth*2
	var wMax float64 = t.cellWidths.maxspeed + padding - lineWidth*2
	var h float64 = t.options.fontHeight + t.options.verticalpadding - 2*lineWidth
	for i := 0; i < len(t.nodes); i++ {
		t.DrawRectangle(x1, y, wAvg, h)
		r, g, b := getSpeedColor(t.nodes[i].AvgSpeed)
		t.SetRGB255(r, g, b)
		t.Fill()
		t.DrawRectangle(x2, y, wMax, h)
		r, g, b = getSpeedColor(t.nodes[i].MaxSpeed)
		t.SetRGB255(r, g, b)
		t.Fill()
		y = y + t.options.fontHeight + t.options.verticalpadding
	}
	t.SetRGB255(0, 0, 0)
}

func (t *Table) Draw(path string, traffic string) {
	t.SetRGB255(255, 255, 255)
	t.Clear()
	t.SetRGB255(0, 0, 0)
	t.drawHorizonLines()
	t.drawVerticalLines()
	t.drawSpeed()
	t.drawTitle()
	t.drawHeader()
	t.drawNodes()
	t.drawTraffic(traffic)
	t.drawGeneratedAt()
	t.SavePNG(path)
}

func getSpeedColor(speed int64) (r int, g int, b int) {
	index := 0
	for i, v := range bounds {
		index = i
		if speed < int64(v) {
			break
		}
	}
	group := colorgroup[index]
	return group[0], group[1], group[2]
}

func calcWidth(fontface font.Face, nodes Nodes) *CellWidths {
	cellWidths := &CellWidths{}
	if len(nodes) < 1 {
		return cellWidths
	}
	cellWidths.group = getWidth(fontface, nodes[0].Group)
	cellWidths.protocol = getWidth(fontface, "Protocol")

	for _, v := range nodes {
		width := getWidth(fontface, v.Ping)
		if cellWidths.ping < width {
			cellWidths.ping = width
		}
		width = getWidth(fontface, download.ByteCountIECTrim(v.AvgSpeed))
		if cellWidths.avgspeed < width {
			cellWidths.avgspeed = width
		}
		width = getWidth(fontface, download.ByteCountIECTrim(v.MaxSpeed))
		if cellWidths.maxspeed < width {
			cellWidths.maxspeed = width
		}
		width = getWidth(fontface, v.Remarks)
		if cellWidths.remarks < width {
			cellWidths.remarks = width
		}
	}
	if cellWidths.group < getWidth(fontface, "Group") {
		cellWidths.group = getWidth(fontface, "Group")
	}
	if cellWidths.remarks < getWidth(fontface, "Remarks") {
		cellWidths.remarks = getWidth(fontface, "Remarks")
	}
	if cellWidths.ping < getWidth(fontface, "Ping") {
		cellWidths.ping = getWidth(fontface, "Ping")
	}
	if cellWidths.avgspeed < getWidth(fontface, "AvgSpeed") {
		cellWidths.avgspeed = getWidth(fontface, "AvgSpeed")
	}
	if cellWidths.maxspeed < getWidth(fontface, "MaxSpeed") {
		cellWidths.maxspeed = getWidth(fontface, "MaxSpeed")
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
