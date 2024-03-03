package main

import (
	"image"
	"image/color"
	"log"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

func Plot(w, h int, legend string, ps ...plot.Plotter) *image.RGBA {
	p := plot.New()
	p.Add(append([]plot.Plotter{
		plotter.NewGrid(),
	}, ps...)...)
	p.Legend.Add(legend)
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	c := vgimg.NewWith(vgimg.UseImage(img))
	p.Draw(draw.New(c))
	return c.Image().(*image.RGBA)
}

type decBoundPlot struct {
	rows, cols int
	f          func(c, r int) float64
}

func (p decBoundPlot) Dims() (c, r int)   { return p.cols, p.rows }
func (p decBoundPlot) Z(c, r int) float64 { return p.f(c, r) }
func (p decBoundPlot) X(c int) float64    { return float64(c) }
func (p decBoundPlot) Y(r int) float64    { return float64(r) }

type dataPlot struct {
	trainPos, trainNeg, testTruePos, testTrueNeg, testFalsePos, testFalseNeg plotter.XYs
	posShape, negShape                                                       draw.GlyphDrawer
	defaultClr, trueClr, falseClr                                            color.RGBA
}

func add(xys *plotter.XYs, x, y float64) {
	*xys = append(*xys, plotter.XY{X: x, Y: y})
}

func (p *dataPlot) addTrain(x1, x2, y float64) {
	if y > 0.5 {
		add(&p.trainPos, x1, x2)
	} else {
		add(&p.trainNeg, x1, x2)
	}
}

func (p *dataPlot) addTest(x1, x2, y, label float64) {
	type key struct{ isPosLabel, isPosY bool }
	xys := map[key]*plotter.XYs{
		{false, false}: &p.testTrueNeg,
		{false, true}:  &p.testFalseNeg,
		{true, false}:  &p.testFalsePos,
		{true, true}:   &p.testTruePos,
	}[key{label > 0.5, y > 0.5}]
	add(xys, x1, x2)
}

func (p dataPlot) series() []*plotter.Scatter {
	sc := func(xys plotter.XYs, shape draw.GlyphDrawer, clr color.RGBA) *plotter.Scatter {
		scatter, err := plotter.NewScatter(xys)
		if err != nil {
			log.Fatal(err)
		}
		scatter.GlyphStyle.Color = clr
		scatter.GlyphStyle.Shape = shape
		return scatter
	}
	return []*plotter.Scatter{
		sc(p.trainPos, p.posShape, p.defaultClr),
		sc(p.trainNeg, p.negShape, p.defaultClr),
		sc(p.testTruePos, p.posShape, p.trueClr),
		sc(p.testFalsePos, p.negShape, p.falseClr),
		sc(p.testTrueNeg, p.negShape, p.trueClr),
		sc(p.testFalseNeg, p.posShape, p.falseClr),
	}
}
