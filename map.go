package main

import (
	"image"
	"image/color"
	_ "image/jpeg"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var directions = [8][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}, {1, 1}, {1, -1}, {-1, 1}, {-1, -1}}

type Map struct {
	Image, ImageMask string
	ebitenImg        *ebiten.Image
	imgMask          image.Image
	safe             []bool
	mw, mh           int
}

func (m *Map) Draw(screen *ebiten.Image, cam *Camera, worldW, worldH int) {
	if m.ebitenImg == nil {
		var err error
		// N.B. Image decoders must be imported, i.e. _ "image/jpeg" in the import section
		m.ebitenImg, _, err = ebitenutil.NewImageFromFile(m.Image)
		if err != nil {
			panic(err)
		}
	}

	imgW, imgH := m.ebitenImg.Bounds().Dx(), m.ebitenImg.Bounds().Dy()

	imgOptions := ebiten.DrawImageOptions{}
	// stretch image
	imgOptions.GeoM.Scale(float64(worldW)/float64(imgW), float64(worldH)/float64(imgH))
	// translate to screen
	imgOptions.GeoM.Scale(float64(cam.Zoom), float64(cam.Zoom))
	sx, sy := cam.WorldToScreen(0, 0)
	imgOptions.GeoM.Translate(float64(sx), float64(sy))
	// smooth when zoomed out
	imgOptions.Filter = ebiten.FilterLinear
	// dim the map to make out agents better
	imgOptions.ColorScale.SetWithColor(color.RGBA{R: 100, G: 100, B: 100, A: 200})
	screen.DrawImage(m.ebitenImg, &imgOptions)
}

func (m *Map) LoadMask() {
	mapImageMaskFile, err := os.Open(m.ImageMask)
	if err != nil {
		panic(err)
	}
	defer func(mapImageMaskFile *os.File) {
		err := mapImageMaskFile.Close()
		if err != nil {
			panic(err)
		}
	}(mapImageMaskFile)

	m.imgMask, _, err = image.Decode(mapImageMaskFile)
	// Save safe squares at load-time to prevent RGBA.at() overhead
	bounds := m.imgMask.Bounds()
	m.mw, m.mh = bounds.Dx(), bounds.Dy()
	m.safe = make([]bool, m.mw*m.mh)
	for y := 0; y < m.mh; y++ {
		for x := 0; x < m.mw; x++ {
			r, g, b, _ := m.imgMask.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			m.safe[y*m.mw+x] = !(r == 0 && g == 0 && b == 0)
		}
	}

	if err != nil {
		panic(err)
	}
}

func (m *Map) IsSafe(x, y int) bool {

	if x < 0 || x >= m.mw {
		panic("x is out of bounds")
	}
	if y < 0 || y >= m.mh {
		panic("y is out of bounds")
	}

	return m.safe[y*m.mw+x]
}

// Used to prevent invalid paths when building the Flow Fields
func (m *Map) canStep(x, y, dx, dy int) bool {
	if !m.IsSafe(x+dx, y+dy) {
		return false
	}
	if dx != 0 && dy != 0 { // no corner cutting
		return m.IsSafe(x+dx, y) && m.IsSafe(x, y+dy)
	}
	return true
}

func (m *Map) BuildFlowField(cx, cy int) []int32 {
	distances := make([]int32, m.mw*m.mh)
	for i := range distances {
		distances[i] = -1
	}

	// BFS (Breadth First Search) each tile to the town center (cx, cy)
	queue := make([]int32, 0, m.mw*m.mh)
	start := cy*m.mw + cx
	distances[start] = 0
	queue = append(queue, int32(start))

	for head := 0; head < len(queue); head++ {
		current := int(queue[head])
		x, y := current%m.mw, current/m.mw
		for _, d := range directions {
			if !m.canStep(x, y, d[0], d[1]) {
				continue
			}
			ni := (y+d[1])*m.mw + (x + d[0])
			if distances[ni] == -1 {
				distances[ni] = distances[current] + 1
				queue = append(queue, int32(ni))
			}
		}
	}
	return distances
}
