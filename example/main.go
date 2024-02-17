package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/palette"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg/draw"
)

type InputReader interface {
	Read() (x [][]float64, y []float64)
}

const (
	inputFileName = "data/exams.csv"

	epochs        = 1e+5
	learningRateW = 1e-3
	learningRateB = 1e-1
)

var (
	rnd = rand.New(rand.NewSource(10))
)

func main() {
	const screenWidth, screenHeight = 640, 480
	ebiten.SetWindowSize(screenWidth, screenWidth)

	var (
		inputs [][]float64
		y      []float64
	)
	inputs, y, err := read(inputFileName)
	if err != nil {
		log.Fatal(err)
	}
	var maxX, maxY float64
	for i := range inputs {
		if inputs[i][0] > maxX {
			maxX = inputs[i][0]
		}
		if inputs[i][1] > maxY {
			maxY = inputs[i][1]
		}
	}

	pointPlt := dataPlot{
		posShape:   draw.PlusGlyph{},
		negShape:   draw.RingGlyph{},
		defaultClr: color.RGBA{A: 255},
		trueClr:    color.RGBA{G: 255, A: 255},
		falseClr:   color.RGBA{R: 255, A: 255},
	}
	xTrain, xTest, yTrain, yTest := split(inputs, y)
	for i := range yTrain {
		pointPlt.addTrain(xTrain[i][0], xTrain[i][1], yTrain[i])
	}

	w := make([]float64, 2)
	var b float64
	for i := 0; i < epochs; i++ {
		p := inference(xTrain, w, b)
		dw, db := dCost(xTrain, yTrain, p)
		for i := range w {
			w[i] -= dw[i] * learningRateW
		}
		b -= db * learningRateB
	}
	fmt.Println("Weight:", w, "Bias:", b)
	fmt.Println("Accuracy:", accuracy(xTest, yTest, w, b))

	prob := inference(xTest, w, b)
	for i := range prob {
		pointPlt.addTest(xTest[i][0], xTest[i][1], yTest[i], prob[i])
	}

	boundPlot := decBoundPlot{
		rows: int(maxY + 1.5),
		cols: int(maxX + 1.5),
		f:    func(c, r int) float64 { return p([]float64{float64(c), float64(r)}, w, b) },
	}
	plotters := []plot.Plotter{
		plotter.NewContour(boundPlot, []float64{0.5}, palette.Heat(1, 255)),
	}
	pps := pointPlt.series()
	for _, p := range pps {
		plotters = append(plotters, p)
	}
	if err := ebiten.RunGame(&App{img: ebiten.NewImageFromImage(Plot(screenWidth, screenHeight, plotters...))}); err != nil {
		log.Fatal(err)
	}
}

func split(inputs [][]float64, y []float64) (xTrain, xTest [][]float64, yTrain, yTest []float64) {
	trainIndices := make(map[int]bool)
	for i := 0; i < len(inputs)/5*4; i++ {
		idx := rnd.Intn(len(inputs))
		for trainIndices[idx] {
			idx = rnd.Intn(len(inputs))
		}
		trainIndices[idx] = true
	}
	for i := 0; i < len(inputs); i++ {
		if trainIndices[i] {
			xTrain = append(xTrain, inputs[i])
			yTrain = append(yTrain, y[i])
		} else {
			xTest = append(xTest, inputs[i])
			yTest = append(yTest, y[i])
		}
	}
	return xTrain, xTest, yTrain, yTest
}

func inference(inputs [][]float64, w []float64, b float64) []float64 {
	var res []float64
	for _, x := range inputs {
		res = append(res, p(x, w, b))
	}
	return res
}

func p(x []float64, w []float64, b float64) float64 {
	return sigmoid(dot(w, x) + b)
}

func sigmoid(z float64) float64 {
	return 1.0 / (1.0 + math.Exp(-z))
}

func dot(a []float64, b []float64) (res float64) {
	for i := 0; i < len(a); i++ {
		res += a[i] * b[i]
	}
	return res
}

func dCost(inputs [][]float64, y, p []float64) (dw []float64, db float64) {
	m := len(inputs)
	n := len(inputs[0])
	dw = make([]float64, n)
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			dw[j] += (p[i] - y[i]) * inputs[i][j]
		}
		db += p[i] - y[i]
	}
	for j := 0; j < n; j++ {
		dw[j] /= float64(m)
	}
	db /= float64(m)
	return dw, db
}

func accuracy(inputs [][]float64, y []float64, w []float64, b float64) float64 {
	p := inference(inputs, w, b)
	var trueOut float64
	for i := range p {
		if int(p[i]+0.5) == int(y[i]+0.5) {
			trueOut++
		}
	}
	return trueOut / float64(len(p))
}
