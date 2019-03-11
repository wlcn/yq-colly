package helper

import (
	"log"
	"math"
	"os"

	"github.com/chenjiandongx/go-echarts/charts"
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
