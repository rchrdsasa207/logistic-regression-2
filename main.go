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
	inputFileName = "data/circle.csv"

	epochs        = 500000
	learningRateW = 1e-10
	learningRateB = 1e-1
	power         = 3
)

var (
	rnd = rand.New(rand.NewSource(11))
)

func main() {
	// polinomial(100, 200)
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
		maxX = max(maxX, inputs[i][0])
		maxY = max(maxY, inputs[i][1])
	}

	pointPlt := dataPlot{
		posShape:   draw.PlusGlyph{},
		negShape:   draw.RingGlyph{},
		defaultClr: color.RGBA{A: 255},
		trueClr:    color.RGBA{G: 255, A: 255},
		falseClr:   color.RGBA{R: 255, A: 255},
	}
	newX := make([][]float64, len(inputs))
	for i, v := range inputs {
		newX[i] = polynomial(v[0], v[1])
	}

	xTrain, xTest, yTrain, yTest, trainIndices, testIndices := split(newX, y, 5)
	for i := range yTrain {
		pointPlt.addTrain(inputs[trainIndices[i]][0], inputs[trainIndices[i]][1], yTrain[i])
	}

	w := make([]float64, len(xTrain[0]))
	fmt.Println(len(xTrain[0]))
	var b, squaredGradB float64
	squaredGradW := make([]float64, len(w))
	epsilon := 1e-8
	learningRate := 1e-3

	for i := 0; i < epochs; i++ {

		p := inference(xTrain, w, b)
		dw, db := dCost(xTrain, yTrain, p)
		for i := range dw {
			squaredGradW[i] += dw[i] * dw[i]
			w[i] -= (learningRate / math.Sqrt(squaredGradW[i]+epsilon)) * dw[i]
		}
		squaredGradB += db * db
		b -= (learningRate / math.Sqrt(squaredGradB+epsilon)) * db
		if i%1000 == 0 {
			fmt.Println(dw, db)
		}
		// for i := range w {
		// 	w[i] -= dw[i] * learningRateW
		// }
		// b -= db * learningRateB
	}
	fmt.Println("Weight:", w, "Bias:", b)
	score := accuracy(xTest, yTest, w, b)
	fmt.Println("Accuracy:", score)

	prob := inference(xTest, w, b)
	for i := range prob {
		pointPlt.addTest(inputs[testIndices[i]][0], inputs[testIndices[i]][1], yTest[i], prob[i])
	}

	boundPlot := decBoundPlot{
		rows: int(maxY + 1.5),
		cols: int(maxX + 1.5),
		f:    func(c, r int) float64 { return p(polynomial(float64(c), float64(r)), w, b) },
	}
	plotters := []plot.Plotter{
		plotter.NewContour(boundPlot, []float64{0.5}, palette.Heat(1, 255)),
	}
	pps := pointPlt.series()
	for _, p := range pps {
		plotters = append(plotters, p)
	}
	legend := fmt.Sprintf("Accuracy: %.2f", score)
	if err := ebiten.RunGame(&App{img: ebiten.NewImageFromImage(Plot(screenWidth, screenHeight, legend, plotters...))}); err != nil {
		log.Fatal(err)
	}
}

func polynomial(x1, x2 float64) (res []float64) {
	for i := 0; i <= power; i++ {
		for j := 0; j <= power-i; j++ {
			res = append(res, math.Pow(x1, float64(i))*math.Pow(x2, float64(j)))
		}
	}
	return res
}

// func split(inputs [][]float64, y []float64) (xTrain, xTest [][]float64, yTrain, yTest []float64) {
// 	trainIndices := make(map[int]bool)
// 	for i := 0; i < len(inputs)/5*4; i++ {
// 		idx := rnd.Intn(len(inputs))
// 		for trainIndices[idx] {
// 			idx = rnd.Intn(len(inputs))
// 		}
// 		trainIndices[idx] = true
// 	}
// 	for i := range inputs {
// 		if trainIndices[i] {
// 			xTrain = append(xTrain, inputs[i])
// 			yTrain = append(yTrain, y[i])
// 		} else {
// 			xTest = append(xTest, inputs[i])
// 			yTest = append(yTest, y[i])
// 		}
// 	}
// 	return xTrain, xTest, yTrain, yTest
// }

func split(inputs [][]float64, y []float64, rate int) (xTrain, xTest [][]float64, yTrain, yTest []float64, trainIndices, testIndices []int) {
	indices := make([]int, len(y))
	for i := range indices {
		indices[i] = i
	}
	rand.Shuffle(len(indices), func(i, j int) { indices[i], indices[j] = indices[j], indices[i] })
	trainSplit := len(indices) / 5 * 4
	for _, i := range indices[:trainSplit] {
		xTrain = append(xTrain, inputs[i])
		yTrain = append(yTrain, y[i])
	}
	for _, i := range indices[trainSplit:] {
		xTest = append(xTest, inputs[i])
		yTest = append(yTest, y[i])
	}
	return xTrain, xTest, yTrain, yTest, indices[:trainSplit], indices[trainSplit:]
}

func inference(inputs [][]float64, w []float64, b float64) []float64 {
	var res []float64
	for _, x := range inputs {
		a := make([]float64, len(x))
		for i := range a {
			if i == 1 {
				a[i] = x[i] * x[i]
			} else {
				a[i] = x[i]
			}
		}
		res = append(res, p(a, w, b))
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
