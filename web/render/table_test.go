package render

import (
	"fmt"
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
	nodes := []Node{
		{
			Group:    "节点列表",
			Remarks:  "美国加利福尼亚免费测试1",
			Protocol: "vmess",
			Ping:     "1206",
			AvgSpeed: "18.18MB",
			MaxSpeed: "32.18MB",
		},
		{
			Group:    "节点列表",
			Remarks:  "美国加利福尼亚免费测试2",
			Ping:     "80",
			Protocol: "trojan",
			AvgSpeed: "18.18MB",
			MaxSpeed: "32.18MB",
		},
		{
			Group:    "节点列表",
			Remarks:  "美国加利福尼亚免费测试3",
			Protocol: "vmess",
			Ping:     "80",
			AvgSpeed: "18.18MB",
			MaxSpeed: "32.18MB",
		},
		{
			Group:    "节点列表",
			Remarks:  "美国加利福尼亚免费测试4",
			Protocol: "ss",
			Ping:     "80",
			AvgSpeed: "18.18MB",
			MaxSpeed: "32.18MB",
		},
		{
			Group:    "节点列表",
			Remarks:  "美国加利福尼亚免费测试5",
			Protocol: "ssr",
			Ping:     "180",
			AvgSpeed: "18.18MB",
			MaxSpeed: "32.18MB",
		},
		{
			Group:    "节点列表",
			Remarks:  "美国加利福尼亚免费测试6",
			Protocol: "vmess",
			Ping:     "120",
			AvgSpeed: "18.18MB",
			MaxSpeed: "32.18MB",
		},
	}
	widths := calcWidth(fontface, nodes)
	fontHeight := calcHeight(fontface)
	var horizontalpadding float64 = 40
	tableWidth := widths.group + horizontalpadding + widths.remarks + horizontalpadding + widths.protocol + horizontalpadding + widths.ping + horizontalpadding + widths.avgspeed + horizontalpadding + widths.maxspeed + horizontalpadding
	options := TableOptions{
		horizontalpadding: horizontalpadding,
		verticalpadding:   36,
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
	table.SetRGB(1, 1, 1)
	table.Clear()
	table.SetRGB(0, 0, 0)
	table.SetFontFace(fontface)
	table.drawHorizonLines()
	table.drawVerticalLines()
	table.drawSpeed()
	table.drawTitle()
	table.drawHeader()
	table.drawNodes()
	table.SavePNG("out.png")
}
