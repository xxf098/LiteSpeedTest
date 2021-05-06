package render

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"testing"
)

func TestDraw(t *testing.T) {
	fontPath, _ := filepath.Abs("../misc/WenQuanYiMicroHei-01.ttf")
	fontSize := 22
	fontface, err := LoadFontFace(fontPath, float64(fontSize))
	if err != nil {
		panic(err)
	}
	nodes := make([]Node, 50)
	for i := 0; i < 50; i++ {
		nodes[i] = Node{
			Group:    "节点列表",
			Remarks:  fmt.Sprintf("美国加利福尼亚免费测试%d", i),
			Protocol: "vmess",
			Ping:     fmt.Sprintf("%d", rand.Intn(800-50)+50),
			AvgSpeed: int64((rand.Intn(20-1) + 1) * 1024 * 1024),
			MaxSpeed: int64((rand.Intn(60-20) + 20) * 1024 * 1024),
		}
	}
	widths := calcWidth(fontface, nodes)
	fontHeight := calcHeight(fontface)
	var horizontalpadding float64 = 40
	tableWidth := widths.Group + horizontalpadding + widths.Remarks + horizontalpadding + widths.Protocol + horizontalpadding + widths.Ping + horizontalpadding + widths.AvgSpeed + horizontalpadding + widths.MaxSpeed + horizontalpadding
	options := TableOptions{
		horizontalpadding: horizontalpadding,
		verticalpadding:   30,
		tableTopPadding:   20,
		lineWidth:         0.5,
		fontHeight:        fontHeight,
	}
	tableHeight := (fontHeight+options.verticalpadding)*float64((len(nodes)+4)) + options.tableTopPadding*2
	fmt.Printf("width: %f, height: %f\n", tableWidth, tableHeight)
	table := NewTable(int(tableWidth), int(tableHeight), options)
	table.nodes = nodes
	table.cellWidths = widths
	// set background
	table.SetFontFace(fontface)
	msg := fmt.Sprintf("Traffic used : %s. Time used : %s, Working Nodes: [%s]", "10.6G", "12:50", "50/50")
	table.Draw("out.png", msg)
}

func TestDefaultTable(t *testing.T) {
	nodes := make([]Node, 50)
	for i := 0; i < 50; i++ {
		nodes[i] = Node{
			Group:    "节点列表",
			Remarks:  fmt.Sprintf("美国加利福尼亚免费测试%d", i),
			Protocol: "vmess",
			Ping:     fmt.Sprintf("%d", rand.Intn(800-50)+50),
			AvgSpeed: int64((rand.Intn(20-1) + 1) * 1024 * 1024),
			MaxSpeed: int64((rand.Intn(60-5) + 5) * 1024 * 1024),
		}
	}
	fontPath, _ := filepath.Abs("../misc/WenQuanYiMicroHei-01.ttf")
	table, _ := DefaultTable(nodes, fontPath)
	msg := table.FormatTraffic("10.2G", "3m13s", "50/50")
	table.Draw("out.png", msg)
}

func TestCSV2Nodes(t *testing.T) {
	nodes, err := CSV2Nodes("/home/arch/Downloads/test.csv")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(nodes)
}

func TestGetNodeHeaders(t *testing.T) {
	_, tags := getNodeHeaders("en")
	for k, v := range tags {
		fmt.Printf("%s:%s\n", k, v)
	}
}
