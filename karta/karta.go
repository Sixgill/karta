package karta

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"math"
	"math/rand"
	"os"

	"github.com/peterhellberg/karta/palette"

	"code.google.com/p/draw2d/draw2d"
	"github.com/pzsz/voronoi"
	"github.com/pzsz/voronoi/utils"
)

const (
	// Define Tau since the Golang math package is lacking
	τ = 2 * math.Pi
)

// NewDiagram generates a new Voronoi diagram, relaxed by Lloyd's algorithm
func NewDiagram(w, h float64, c, r int) *voronoi.Diagram {
	bbox := voronoi.NewBBox(0, w, 0, h)
	sites := utils.RandomSites(bbox, c)

	// Compute voronoi diagram.
	diagram := voronoi.ComputeDiagram(sites, bbox, true)

	// Max number of iterations is 16
	if r > 16 {
		r = 16
	}

	// Relax using Lloyd's algorithm
	for i := 0; i < r; i++ {
		sites = utils.LloydRelaxation(diagram.Cells)
		diagram = voronoi.ComputeDiagram(sites, bbox, true)
	}

	return diagram
}

type CellPrefs struct {
	Index            int
	DistanceToCenter float64
	FillColor        color.RGBA
	StrokeColor      color.RGBA
}

// DrawDiagramImage draws a Voroni diagram to an image
func DrawDiagramImage(diagram *voronoi.Diagram, w, h int) image.Image {
	f, _ := os.OpenFile("karta.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0640)
	defer f.Close()
	log.SetOutput(f)

	unit := float64((w + h) / 20)

	allPrefs := make(map[int]CellPrefs)
	centerOfMap := voronoi.Vertex{float64(w / 2), float64(h / 2)}

	for i, cell := range diagram.Cells {
		allPrefs[i] = CellPrefs{
			Index:            i,
			DistanceToCenter: utils.Distance(cell.Site, centerOfMap),
		}
	}

	img := image.NewRGBA(image.Rect(0, 0, w, h))

	draw.Draw(img, img.Bounds(), &image.Uniform{palette.Darkblue}, image.ZP, draw.Src)

	p := draw2d.NewGraphicContext(img)
	p.SetFillColor(palette.Green)

	l := draw2d.NewGraphicContext(img)

	var tiny float64

	switch {
	case h <= 256:
		tiny = 1
	case h <= 512:
		tiny = 2
	default:
		tiny = 3
	}

	l.SetLineWidth(tiny)

	// Iterate over cells
	for i, cell := range diagram.Cells {
		prefs := allPrefs[i]

		d := prefs.DistanceToCenter

		switch {
		case d < unit*1.4 && rand.Intn(4) < 3:
			prefs.FillColor = palette.Darkergreen
			prefs.StrokeColor = palette.Darkgreen
		case d < unit*3.4:
			prefs.FillColor = palette.Green
			prefs.StrokeColor = palette.Darkgreen
		case d < unit*4.2 && rand.Intn(4) < 3:
			prefs.FillColor = palette.Blue
			prefs.StrokeColor = palette.Darkerblue
		default:
			prefs.FillColor = palette.Darkblue
			prefs.StrokeColor = palette.Darkerblue
		}

		// Make sure left and right edges of the map are deep water
		if cell.Site.X < unit || cell.Site.X > float64(w)-unit {
			if prefs.FillColor == palette.Green {
				prefs.FillColor = palette.Blue
				prefs.StrokeColor = palette.Darkerblue
			}

			if prefs.FillColor != palette.Blue {
				prefs.FillColor = palette.Darkblue
				prefs.StrokeColor = palette.Darkerblue
			}
		}

		// Make sure top and bottom edges of the map are deep water
		if cell.Site.Y < unit/1.2 || cell.Site.Y > float64(h)-unit/1.2 {
			if prefs.FillColor != palette.Blue {
				prefs.FillColor = palette.Darkblue
				prefs.StrokeColor = palette.Darkerblue
			}

			if prefs.FillColor == palette.Green {
				prefs.FillColor = palette.Blue
				prefs.StrokeColor = palette.Darkerblue
			}
		}

		if cell.Site.Y < unit/3 || cell.Site.Y > float64(h)-unit/3 {
			prefs.FillColor = palette.Darkblue
			prefs.StrokeColor = palette.Darkerblue
		}

		l.SetFillColor(prefs.FillColor)
		l.SetStrokeColor(prefs.StrokeColor)

		for _, hedge := range cell.Halfedges {
			a := hedge.GetStartpoint()
			b := hedge.GetEndpoint()

			l.MoveTo(a.X, a.Y)
			l.LineTo(b.X, b.Y)
		}

		l.FillStroke()

		p.ArcTo(cell.Site.X, cell.Site.Y, tiny*1.4, tiny*1.4, 0, τ)
		p.FillStroke()
	}

	l.Close()

	return img
}

// SaveImage saves an image to a file
func SaveImage(img image.Image, fn string) error {
	file, err := os.Create(fn)
	if err != nil {
		return err
	}
	defer file.Close()

	png.Encode(file, img)
	if err != nil {
		return err
	}

	return nil
}