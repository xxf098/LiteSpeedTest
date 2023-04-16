package render

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"time"

	"github.com/xxf098/lite-proxy/constant"
	"github.com/xxf098/lite-proxy/download"
	"golang.org/x/image/font"
)

type Theme struct {
	colorgroup [][]int
	bounds     []int
}

var (
	themes = map[string]Theme{
		"original": Theme{
			colorgroup: [][]int{
				{255, 255, 255},
				{128, 255, 0},
				{255, 255, 0},
				{255, 128, 192},
				{255, 0, 0},
			},
			bounds: []int{0, 64 * 1024, 512 * 1024, 4 * 1024 * 1024, 16 * 1024 * 1024},
		},
		"rainbow": Theme{
			colorgroup: [][]int{
				{255, 255, 255},
				{102, 255, 102},
				{255, 255, 102},
				{255, 178, 102},
				{255, 102, 102},
				{226, 140, 255},
				{102, 204, 255},
				{102, 102, 255},
			},
			bounds: []int{0, 64 * 1024, 512 * 1024, 4 * 1024 * 1024, 16 * 1024 * 1024, 24 * 1024 * 1024, 32 * 1024 * 1024, 40 * 1024 * 1024},
		},
	}

	i18n = map[string]string{
		"cn": `{
			"Title":    "Lite SpeedTest 结果表",
			"CreateAt": "测试时间",
			"Traffic":  "总流量: %s. 总时间: %s, 可用节点: [%s]"
		}`,
		"en": `{
			"Title":    "Lite SpeedTest Result Table",
			"CreateAt": "Create At",
			"Traffic":  "Traffic used: %s. Time used: %s, Working Nodes: [%s]"
		}`,
	}
)

type Node struct {
	Id       int    `json:"id"`
	Group    string `en:"Group" cn:"群组名" json:"group"`
	Remarks  string `en:"Remarks" cn:"备注" json:"remarks"`
	Protocol string `en:"Protocol" cn:"协议" json:"protocol"`
	Ping     string `en:"Ping" cn:"Ping" json:"ping"`
	AvgSpeed int64  `en:"AvgSpeed" cn:"平均速度" json:"avg_speed"`
	MaxSpeed int64  `en:"MaxSpeed" cn:"最大速度" json:"max_speed"`
	IsOk     bool   `json:"isok"`
	Traffic  int64  `json:"traffic"`
	Link     string `json:"link,omitempty"` // api only
}

func getNodeHeaders(language string) ([]string, map[string]string) {
	kvs := map[string]string{}
	var keys []string
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

func (nodes Nodes) Sort(sortMethod string) {
	sort.Slice(nodes[:], func(i, j int) bool {
		switch sortMethod {
		case "speed":
			return nodes[i].MaxSpeed < nodes[j].MaxSpeed
		case "rspeed":
			return nodes[i].MaxSpeed > nodes[j].MaxSpeed
		case "ping":
			return nodes[i].Ping < nodes[j].Ping
		case "rping":
			return nodes[i].Ping > nodes[j].Ping
		default:
			return true
		}
	})
}

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
	theme             Theme
	timezone          string
	fontBytes         []byte
}

func NewTableOptions(horizontalpadding float64, verticalpadding float64, tableTopPadding float64,
	lineWidth float64, fontSize int, smallFontRatio float64, fontPath string,
	language string, t string, timezone string, fontBytes []byte) TableOptions {
	theme, ok := themes[t]
	if !ok {
		theme = themes["rainbow"]
	}
	return TableOptions{
		horizontalpadding: horizontalpadding,
		verticalpadding:   verticalpadding,
		tableTopPadding:   tableTopPadding,
		lineWidth:         lineWidth,
		fontSize:          fontSize,
		smallFontRatio:    smallFontRatio,
		fontPath:          fontPath,
		language:          language,
		theme:             theme,
		timezone:          timezone,
		fontBytes:         fontBytes,
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
	data, _ := json.Marshal(&c)
	m := map[string]float64{}
	// ignore error
	json.Unmarshal(data, &m)
	return m
}

type I18N struct {
	CreateAt string
	Title    string
	Traffic  string
}

func NewI18N(data string) (*I18N, error) {
	i18n := &I18N{}
	err := json.Unmarshal([]byte(data), i18n)
	if err != nil {
		return nil, err
	}
	return i18n, nil
}

type Table struct {
	width  int
	height int
	*Context
	nodes      Nodes
	options    TableOptions
	cellWidths *CellWidths
	i18n       *I18N
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
	options := NewTableOptions(40, 30, 0.5, 0.5, 24, 0.5, fontPath, "en", "rainbow", "Asia/Shanghai", nil)
	return NewTableWithOption(nodes, &options)
}

// TODO: load font by name
func NewTableWithOption(nodes Nodes, options *TableOptions) (*Table, error) {
	fontSize := options.fontSize
	fontPath := options.fontPath
	fontface, err := LoadFontFaceByBytes(options.fontBytes, fontPath, float64(fontSize))
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
	result, err := NewI18N(i18n[options.language])
	if err != nil {
		return nil, err
	}
	table.i18n = result
	table.SetFontFace(fontface)
	return &table, nil
}

func (t *Table) drawHorizonLines() {
	y := t.options.tableTopPadding
	for i := 0; i <= len(t.nodes)+4; i++ {
		t.drawHorizonLine(y)
		y += t.options.fontHeight + t.options.verticalpadding
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
	title := t.i18n.Title
	var x float64 = float64(t.width)/2 - getWidth(t.fontFace, title)/2
	var y float64 = t.options.fontHeight/2 + t.options.verticalpadding/2 + t.options.tableTopPadding
	t.centerString(title, x, y)
}

func (t *Table) drawHeader() {
	horizontalpadding := t.options.horizontalpadding
	var x float64 = horizontalpadding / 2
	var y float64 = t.options.fontHeight/2 + t.options.verticalpadding/2 + t.options.tableTopPadding + t.options.fontHeight + t.options.verticalpadding
	cellWidths := t.cellWidths.toMap()
	ks, kvs := getNodeHeaders(t.options.language)
	for _, k := range ks {
		adjust := cellWidths[k]/2 - getWidth(t.fontFace, kvs[k])/2
		t.centerString(kvs[k], x+adjust, y)
		x += cellWidths[k] + horizontalpadding
	}
}

func (t *Table) drawTraffic(traffic string) {
	// horizontalpadding := t.options.horizontalpadding
	var x float64 = t.options.horizontalpadding / 2
	var y float64 = (t.options.fontHeight+t.options.verticalpadding)*float64((len(t.nodes)+2)) + t.options.tableTopPadding + t.fontHeight/2 + t.options.verticalpadding/2
	t.centerString(traffic, x, y)
}

func (t *Table) FormatTraffic(traffic string, time string, workingNode string) string {
	return fmt.Sprintf(t.i18n.Traffic, traffic, time, workingNode)
}

func (t *Table) drawGeneratedAt() {
	// horizontalpadding := t.options.horizontalpadding
	msg := fmt.Sprintf("%s %s", t.i18n.CreateAt, time.Now().Format(time.RFC3339))
	// https://github.com/golang/go/issues/20455
	if runtime.GOOS == "android" {
		loc, _ := time.LoadLocation(t.options.timezone)
		now := time.Now()
		msg = fmt.Sprintf("%s %s", t.i18n.CreateAt, now.In(loc).Format(time.RFC3339))
	}
	var x float64 = t.options.horizontalpadding / 2
	var y float64 = (t.options.fontHeight+t.options.verticalpadding)*float64((len(t.nodes)+3)) + t.options.tableTopPadding + t.fontHeight/2 + t.options.verticalpadding/2
	t.centerString(msg, x, y)
}

func (t *Table) drawPoweredBy() {
	fontSize := int(float64(t.options.fontSize) * t.options.smallFontRatio)
	fontface, err := LoadFontFaceByBytes(t.options.fontBytes, t.options.fontPath, float64(fontSize))
	if err != nil {
		return
	}
	t.SetFontFace(fontface)
	msg := constant.Version + " powered by https://github.com/xxf098"
	var x float64 = float64(t.width) - getWidth(fontface, msg) - t.options.lineWidth
	var y float64 = (t.options.fontHeight+t.options.verticalpadding)*float64((len(t.nodes)+4)) + t.options.fontHeight*t.options.smallFontRatio
	t.DrawString(msg, x, y)
}

func (t *Table) centerString(s string, x, y float64) {
	t.DrawStringAnchored(s, x, y, 0, 0.4)
}

func (t *Table) drawNodes() {
	horizontalpadding := t.options.horizontalpadding
	var x float64 = horizontalpadding / 2
	var y float64 = t.options.fontHeight/2 + t.options.verticalpadding/2 + t.options.tableTopPadding + (t.options.fontHeight+t.options.verticalpadding)*2
	for _, v := range t.nodes {
		t.centerString(v.Group, x, y)
		x += t.cellWidths.Group + horizontalpadding
		t.centerString(v.Remarks, x, y)
		x += t.cellWidths.Remarks + horizontalpadding
		adjust := t.cellWidths.Protocol/2 - getWidth(t.fontFace, v.Protocol)/2
		t.centerString(v.Protocol, x+adjust, y)
		x += t.cellWidths.Protocol + horizontalpadding
		adjust = t.cellWidths.Ping/2 - getWidth(t.fontFace, v.Ping)/2
		t.centerString(v.Ping, x+adjust, y)
		x += t.cellWidths.Ping + horizontalpadding
		avgSpeed := download.ByteCountIECTrim(v.AvgSpeed)
		adjust = t.cellWidths.AvgSpeed/2 - getWidth(t.fontFace, avgSpeed)/2
		t.centerString(avgSpeed, x+adjust, y)
		x += t.cellWidths.AvgSpeed + horizontalpadding
		maxSpeed := download.ByteCountIECTrim(v.MaxSpeed)
		adjust = t.cellWidths.MaxSpeed/2 - getWidth(t.fontFace, maxSpeed)/2
		t.centerString(maxSpeed, x+adjust, y)
		y += t.options.fontHeight + t.options.verticalpadding
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
		r, g, b := getSpeedColor(t.nodes[i].AvgSpeed, t.options.theme)
		t.SetRGB255(r, g, b)
		t.Fill()
		t.DrawRectangle(x2, y, wMax, h)
		r, g, b = getSpeedColor(t.nodes[i].MaxSpeed, t.options.theme)
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

func (t *Table) Encode(traffic string) ([]byte, error) {
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
	var buf bytes.Buffer
	err := t.EncodePNG(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (t *Table) EncodeB64(traffic string) (string, error) {
	bytes, err := t.Encode(traffic)
	if err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(bytes), nil
}

func getSpeedColor(speed int64, theme Theme) (int, int, int) {
	bounds := theme.bounds
	colorgroup := theme.colorgroup
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

func getWidth(fontface font.Face, s string) float64 {
	a := font.MeasureString(fontface, s)
	return float64(a >> 6)
}
