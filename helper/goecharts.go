package helper

import (
	"log"
	"math"
	"math/rand"
	"os"

	"github.com/chenjiandongx/go-echarts/charts"
	"github.com/chenjiandongx/go-echarts/common"
)

var rangeColor = []string{
	"#313695", "#4575b4", "#74add1", "#abd9e9", "#e0f3f8",
	"#fee090", "#fdae61", "#f46d43", "#d73027", "#a50026",
}

func genSurface3dData0() [][3]interface{} {
	data := make([][3]interface{}, 0)
	for i := -60; i < 60; i++ {
		y := float64(i) / 60
		for j := -60; j < 60; j++ {
			x := float64(j) / 60
			z := math.Sin(x*math.Pi) * math.Sin(y*math.Pi)
			data = append(data, [3]interface{}{x, y, z})
		}
	}
	return data
}

func genSurface3dData1() [][3]interface{} {
	data := make([][3]interface{}, 0)
	for i := -30; i < 30; i++ {
		y := float64(i) / 10
		for j := -30; j < 30; j++ {
			x := float64(j) / 10
			z := math.Sin(x*x+y*y) * x / math.Pi
			data = append(data, [3]interface{}{x, y, z})
		}
	}
	return data
}

// Chart1 comment
func Chart1() {
	surface3d := charts.NewSurface3D()
	surface3d.SetGlobalOptions(
		charts.TitleOpts{Title: "surface3D-示例图"},
		charts.VisualMapOpts{
			Calculable: true,
			InRange:    charts.VMInRange{Color: rangeColor},
			Max:        3,
			Min:        -3,
		},
	)
	surface3d.AddZAxis("surface3d", genSurface3dData0())
	f, err := os.Create("chart1.html")
	if err != nil {
		log.Println(err)
	}
	surface3d.Render(f)
}

// Chart2 comment
func Chart2() {
	surface3d := charts.NewSurface3D()
	surface3d.SetGlobalOptions(
		charts.TitleOpts{Title: "surface3D-一朵玫瑰"},
		charts.VisualMapOpts{
			Calculable: true,
			InRange:    charts.VMInRange{Color: rangeColor},
			Max:        3,
			Min:        -3,
		},
	)
	surface3d.AddZAxis("surface3d", genSurface3dData1())
	f, err := os.Create("chart2.html")
	if err != nil {
		log.Println(err)
	}
	surface3d.Render(f)
}

// Chart3 comment
func Chart3() {
	pie := charts.NewPie()
	pie.SetGlobalOptions(
		charts.TitleOpts{Title: "商城对比"},
		charts.ToolboxOpts{Show: true},
		charts.InitOpts{Width: "600px", Height: "400px", PageTitle: "Awesome", Theme: common.ThemeType.Macarons},
	)
	pie.Add("Data", genKvData(),
		charts.LabelTextOpts{Show: true, Formatter: "{b}: {c}"},
		charts.PieOpts{Radius: []string{"30%", "75%"}, RoseType: "area"},
	)
	f, err := os.Create("chart3.html")
	if err != nil {
		log.Println(err)
	}
	pie.Render(f)
}

func genKvData() map[string]interface{} {
	kvData := make(map[string]interface{})
	nameItems := []string{"京东", "国美", "苏宁", "天猫"}
	cnt := len(nameItems)
	for i := 0; i < cnt; i++ {
		kvData[nameItems[i]] = rand.Intn(10000)
	}
	return kvData
}

// Chart4 comment
func Chart4() {
	var wcData = map[string]interface{}{
		"全自动":    10000,
		"甩干功能":   6181,
		"人工智能系统": 4386,
		"儿童锁":    4055,
		"滚筒":     2467,
		"烘干":     2244,
		"温水":     1898,
		"洗涤":     1484,
		"洗衣液":    1689,
		"半自动":    1112,
		"高度":     985,
		"内筒":     847,
		"悬浮":     582,
		"独立悬挂":   555,
		"经济":     550,
		"性价比":    462,
		"销量冠军":   366,
		"低噪音":    282,
		"经典灰色":   273,
		"排水孔":    265,
	}
	wc := charts.NewWordCloud()
	wc.SetGlobalOptions(charts.TitleOpts{Title: "WordCloud-示例图"})
	wc.Add("wordcloud", wcData, charts.WordCLoudOpts{SizeRange: []float32{14, 80}})
	f, err := os.Create("chart4.html")
	if err != nil {
		log.Println(err)
	}
	wc.Render(f)
}
