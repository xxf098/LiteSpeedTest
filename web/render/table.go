package render

import (
	"encoding/csv"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/xxf098/lite-proxy/download"
	"golang.org/x/image/font"
)

var (
	colorgroup = [][]int{
		{255, 255, 255},
		{102, 255, 102},
		{255, 255, 102},
		{255, 178, 102},
		{255, 102, 102},
		{226, 140, 255},
		{102, 204, 255},
		{102, 102, 255},
	}
	bounds = []int{0, 64 * 1024, 512 * 1024, 4 * 1024 * 1024, 16 * 1024 * 1024, 24 * 1024 * 1024, 32 * 1024 * 1024, 40 * 1024 * 1024}

	// colorgroup = [][]int{
	// 	{255, 255, 255},
	// 	{128, 255, 0},
	// 	{255, 255, 0},
	// 	{255, 128, 192},
	// 	{255, 0, 0},
	// }
	// bounds = []int{0, 64 * 1024, 512 * 1024, 4 * 1024 * 1024, 16 * 1024 * 1024}

	i18n = map[string]map[string]string{
		"cn": {
			"title":    "LiteSpeedTest结果表",
			"createAt": "测试时间",
			"traffic":  "总流量: %s. 总时间: %s, 可用节点: [%s]",
		},
		"en": {
			"title":    "LiteSpeedTest Result Table",
			"createAt": "Create At",
			"traffic":  "Traffic used: %s. Time used: %s, Working Nodes: [%s]",
		},
	}
)

type Node struct {
	Id       int
	Group    string `en:"Group" cn:"组名"`
	Remarks  string `en:"Remarks" cn:"备注"`
	Protocol string `en:"Protocol" cn:"协议"`
	Ping     string `en:"Ping" cn:"Ping"`
	AvgSpeed int64  `en:"AvgSpeed" cn:"平均速度"`
	MaxSpeed int64  `en:"MaxSpeed" cn:"最大速度"`
	IsOk     bool
	Traffic  int64
}

func getNodeHeaders(language string) ([]string, map[string]string) {
	kvs := map[string]string{}
	keys := []string{}
	t := reflect.TypeOf(Node{})
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if v, ok := f.Tag.Lookup(language); ok {
			kvs[f.Name] = v
			keys = append(keys, f.Name)
		}
	}
	return keys, kvs
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
	fontSize          int
	smallFontRatio    float64
	fontPath          string
	language          string
}

func NewTableOptions(horizontalpadding float64, verticalpadding float64, tableTopPadding float64,
	lineWidth float64, fontSize int, smallFontRatio float64, fontPath string, language string) TableOptions {
	return TableOptions{
		horizontalpadding: horizontalpadding,
		verticalpadding:   verticalpadding,
		tableTopPadding:   tableTopPadding,
		lineWidth:         lineWidth,
		fontSize:          fontSize,
		smallFontRatio:    smallFontRatio,
		fontPath:          fontPath,
		language:          language,
	}
}

type CellWidths struct {
	Group    float64
	Remarks  float64
	Protocol float64
	Ping     float64
	AvgSpeed float64
	MaxSpeed float64
}

func (c CellWidths) toMap() map[string]float64 {
	m := map[string]float64{}
	m["Group"] = c.Group
	m["Remarks"] = c.Remarks
	m["Protocol"] = c.Protocol
	m["Ping"] = c.Ping
	m["AvgSpeed"] = c.AvgSpeed
	m["MaxSpeed"] = c.MaxSpeed
	return m
}

type I18N struct {
	createAt string
	title    string
	traffic  string
}

func NewI18N(kvs map[string]string) I18N {
	return I18N{
		createAt: kvs["createAt"],
		title:    kvs["title"],
		traffic:  kvs["traffic"],
	}
}

type Table struct {
	width  int
	height int
	*Context
	nodes      Nodes
	options    TableOptions
	cellWidths *CellWidths
	i18n       I18N
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
	options := NewTableOptions(40, 30, 0.5, 0.5, 22, 0.5, fontPath, "cn")
	return NewTableWithOption(nodes, &options)
}

// TODO: load font by name
func NewTableWithOption(nodes Nodes, options *TableOptions) (*Table, error) {
	fontSize := options.fontSize
	fontPath := options.fontPath
	fontface, err := LoadFontFace(fontPath, float64(fontSize))
	if err != nil {
		return nil, err
	}
	widths := calcWidth(fontface, nodes)
	fontHeight := calcHeight(fontface)
	options.fontHeight = fontHeight
	horizontalpadding := options.horizontalpadding
	tableWidth := widths.Group + horizontalpadding + widths.Remarks + horizontalpadding + widths.Protocol + horizontalpadding + widths.Ping + horizontalpadding + widths.AvgSpeed + horizontalpadding + widths.MaxSpeed + horizontalpadding + options.lineWidth*2
	tableHeight := (fontHeight+options.verticalpadding)*float64((len(nodes)+4)) + options.tableTopPadding*2 + options.fontHeight*options.smallFontRatio
	table := NewTable(int(tableWidth), int(tableHeight), *options)
	table.nodes = nodes
	table.cellWidths = widths
	table.i18n = NewI18N(i18n[options.language])
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
	var x float64
	t.drawFullVerticalLine(t.options.lineWidth)
	ks, _ := getNodeHeaders(t.options.language)
	cellWidths := t.cellWidths.toMap()
	for i := 1; i < len(cellWidths); i++ {
		k := ks[i-1]
		x += cellWidths[k] + padding
		t.drawVerticalLine(x)
	}
	x += cellWidths[ks[len(cellWidths)-1]] + padding
	t.drawFullVerticalLine(x)
}

func (t *Table) drawVerticalLine(x float64) {
	height := (t.options.fontHeight+t.options.verticalpadding)*float64((len(t.nodes)+2)) + t.options.tableTopPadding
	y := t.options.tableTopPadding + t.options.fontHeight + t.options.verticalpadding
	t.DrawLine(x, y, x, height)
	t.SetLineWidth(t.options.lineWidth)
	t.Stroke()
}

func (t *Table) drawFullVerticalLine(x float64) {
	height := (t.options.fontHeight+t.options.verticalpadding)*float64((len(t.nodes)+4)) + t.options.tableTopPadding
	y := t.options.tableTopPadding
	t.DrawLine(x, y, x, height)
	t.SetLineWidth(t.options.lineWidth)
	t.Stroke()
}

func (t *Table) drawTitle() {
	// horizontalpadding := t.options.horizontalpadding
	title := t.i18n.title
	var x float64 = float64(t.width)/2 - getWidth(t.fontFace, title)/2
	var y float64 = t.options.fontHeight + t.options.verticalpadding/2 + t.options.tableTopPadding
	t.DrawString(title, x, y)
}

func (t *Table) drawHeader() {
	horizontalpadding := t.options.horizontalpadding
	var x float64 = horizontalpadding / 2
	var y float64 = t.options.fontHeight + t.options.verticalpadding/2 + t.options.tableTopPadding + t.options.fontHeight + t.options.verticalpadding
	cellWidths := t.cellWidths.toMap()
	ks, kvs := getNodeHeaders(t.options.language)
	for _, k := range ks {
		adjust := cellWidths[k]/2 - getWidth(t.fontFace, kvs[k])/2
		t.DrawString(kvs[k], x+adjust, y)
		x += cellWidths[k] + horizontalpadding
	}
}

func (t *Table) drawTraffic(traffic string) {
	// horizontalpadding := t.options.horizontalpadding
	var x float64 = t.options.horizontalpadding / 2
	var y float64 = (t.options.fontHeight+t.options.verticalpadding)*float64((len(t.nodes)+2)) + t.options.tableTopPadding + t.fontHeight + t.options.verticalpadding/2
	t.DrawString(traffic, x, y)
}

func (t *Table) FormatTraffic(traffic string, time string, workingNode string) string {
	return fmt.Sprintf(t.i18n.traffic, traffic, time, workingNode)
}

func (t *Table) drawGeneratedAt() {
	// horizontalpadding := t.options.horizontalpadding
	msg := fmt.Sprintf("%s %s", t.i18n.createAt, time.Now().Format(time.RFC3339))
	var x float64 = t.options.horizontalpadding / 2
	var y float64 = (t.options.fontHeight+t.options.verticalpadding)*float64((len(t.nodes)+3)) + t.options.tableTopPadding + t.fontHeight + t.options.verticalpadding/2
	t.DrawString(msg, x, y)
}

func (t *Table) drawPoweredBy() {
	fontSize := int(float64(t.options.fontSize) * t.options.smallFontRatio)
	fontface, err := LoadFontFace(t.options.fontPath, float64(fontSize))
	if err != nil {
		return
	}
	t.SetFontFace(fontface)
	msg := "powered by https://github.com/xxf098"
	var x float64 = float64(t.width) - getWidth(fontface, msg) - t.options.lineWidth
	var y float64 = (t.options.fontHeight+t.options.verticalpadding)*float64((len(t.nodes)+4)) + t.options.fontHeight*t.options.smallFontRatio
	t.DrawString(msg, x, y)
}

func (t *Table) drawNodes() {
	horizontalpadding := t.options.horizontalpadding
	var x float64 = horizontalpadding / 2
	var y float64 = t.options.fontHeight + t.options.verticalpadding/2 + t.options.tableTopPadding + (t.options.fontHeight+t.options.verticalpadding)*2
	for _, v := range t.nodes {
		t.DrawString(v.Group, x, y)
		x += t.cellWidths.Group + horizontalpadding
		t.DrawString(v.Remarks, x, y)
		x += t.cellWidths.Remarks + horizontalpadding
		adjust := t.cellWidths.Protocol/2 - getWidth(t.fontFace, v.Protocol)/2
		t.DrawString(v.Protocol, x+adjust, y)
		x += t.cellWidths.Protocol + horizontalpadding
		adjust = t.cellWidths.Ping/2 - getWidth(t.fontFace, v.Ping)/2
		t.DrawString(v.Ping, x+adjust, y)
		x += t.cellWidths.Ping + horizontalpadding
		avgSpeed := download.ByteCountIECTrim(v.AvgSpeed)
		adjust = t.cellWidths.AvgSpeed/2 - getWidth(t.fontFace, avgSpeed)/2
		t.DrawString(avgSpeed, x+adjust, y)
		x += t.cellWidths.AvgSpeed + horizontalpadding
		maxSpeed := download.ByteCountIECTrim(v.MaxSpeed)
		adjust = t.cellWidths.MaxSpeed/2 - getWidth(t.fontFace, maxSpeed)/2
		t.DrawString(maxSpeed, x+adjust, y)
		y = y + t.options.fontHeight + t.options.verticalpadding
		x = horizontalpadding / 2
	}
}

func (t *Table) drawSpeed() {
	padding := t.options.horizontalpadding
	var lineWidth float64 = t.options.lineWidth
	var x1 float64 = t.cellWidths.Group + padding + t.cellWidths.Remarks + padding + t.cellWidths.Protocol + padding + t.cellWidths.Ping + padding + lineWidth
	var x2 float64 = t.cellWidths.Group + padding + t.cellWidths.Remarks + padding + t.cellWidths.Protocol + padding + t.cellWidths.Ping + padding + t.cellWidths.AvgSpeed + padding + lineWidth
	var y float64 = t.options.tableTopPadding + lineWidth + (t.options.fontHeight+t.options.verticalpadding)*2
	var wAvg float64 = t.cellWidths.AvgSpeed + padding - lineWidth*2
	var wMax float64 = t.cellWidths.MaxSpeed + padding - lineWidth*2
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
	t.drawPoweredBy()
	t.SavePNG(path)
}

func getSpeedColor(speed int64) (int, int, int) {
	for i := 0; i < len(bounds)-1; i++ {
		if speed >= int64(bounds[i]) && speed <= int64(bounds[i+1]) {
			level := float64(speed-int64(bounds[i])) / float64(bounds[i+1]-bounds[i])
			return getColor(colorgroup[i], colorgroup[i+1], level)
		}
	}
	l := len(colorgroup)
	return colorgroup[l-1][0], colorgroup[l-1][1], colorgroup[l-1][2]
}

func getColor(lc []int, rc []int, level float64) (int, int, int) {
	r := float64(lc[0])*(1-level) + float64(rc[0])*level
	g := float64(lc[1])*(1-level) + float64(rc[1])*level
	b := float64(lc[2])*(1-level) + float64(rc[2])*level
	return int(r), int(g), int(b)
}

func calcWidth(fontface font.Face, nodes Nodes) *CellWidths {
	cellWidths := &CellWidths{}
	if len(nodes) < 1 {
		return cellWidths
	}
	cellWidths.Group = getWidth(fontface, nodes[0].Group)
	cellWidths.Protocol = getWidth(fontface, "Protocol")

	for _, v := range nodes {
		width := getWidth(fontface, v.Ping)
		if cellWidths.Ping < width {
			cellWidths.Ping = width
		}
		width = getWidth(fontface, download.ByteCountIECTrim(v.AvgSpeed))
		if cellWidths.AvgSpeed < width {
			cellWidths.AvgSpeed = width
		}
		width = getWidth(fontface, download.ByteCountIECTrim(v.MaxSpeed))
		if cellWidths.MaxSpeed < width {
			cellWidths.MaxSpeed = width
		}
		width = getWidth(fontface, v.Remarks)
		if cellWidths.Remarks < width {
			cellWidths.Remarks = width
		}
	}
	if cellWidths.Group < getWidth(fontface, "Group") {
		cellWidths.Group = getWidth(fontface, "Group")
	}
	if cellWidths.Remarks < getWidth(fontface, "Remarks") {
		cellWidths.Remarks = getWidth(fontface, "Remarks")
	}
	if cellWidths.Ping < getWidth(fontface, "Ping") {
		cellWidths.Ping = getWidth(fontface, "Ping")
	}
	if cellWidths.AvgSpeed < getWidth(fontface, "AvgSpeed") {
		cellWidths.AvgSpeed = getWidth(fontface, "AvgSpeed")
	}
	if cellWidths.MaxSpeed < getWidth(fontface, "MaxSpeed") {
		cellWidths.MaxSpeed = getWidth(fontface, "MaxSpeed")
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
