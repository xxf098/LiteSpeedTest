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
			AvgSpeed: fmt.Sprintf("%d.%dMB", rand.Intn(20-5)+5, rand.Intn(20-5)+5),
			MaxSpeed: fmt.Sprintf("%d.%dMB", rand.Intn(70-20)+20, rand.Intn(70-20)+20),
		}
	}
	widths := calcWidth(fontface, nodes)
	fontHeight := calcHeight(fontface)
	var horizontalpadding float64 = 40
	tableWidth := widths.group + horizontalpadding + widths.remarks + horizontalpadding + widths.protocol + horizontalpadding + widths.ping + horizontalpadding + widths.avgspeed + horizontalpadding + widths.maxspeed + horizontalpadding
	options := TableOptions{
		horizontalpadding: horizontalpadding,
		verticalpadding:   30,
		tableTopPadding:   20,
		lineWidth:         0.6,
		fontHeight:        fontHeight,
	}
	tableHeight := (fontHeight+options.verticalpadding)*float64((len(nodes)+4)) + options.tableTopPadding*2
	fmt.Printf("width: %f, height: %f\n", tableWidth, tableHeight)
	table := NewTable(int(tableWidth), int(tableHeight), options)
	table.nodes = nodes
	table.cellWidths = widths
	// set background
	table.SetRGB255(255, 255, 255)
	table.Clear()
	table.SetRGB255(0, 0, 0)
	table.SetFontFace(fontface)
	table.drawHorizonLines()
	table.drawVerticalLines()
	table.drawSpeed()
	table.drawTitle()
	table.drawHeader()
	table.drawNodes()
	table.drawTraffic("9.45GB", "06:24", "50/50")
	table.drawGeneratedAt()
	table.SavePNG("out.png")
}
