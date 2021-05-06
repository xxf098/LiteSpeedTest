package render

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"testing"
)

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
	table, err := DefaultTable(nodes, fontPath)
	if err != nil {
		t.Error(err)
	}
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
