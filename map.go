package main

import (
	"image"
	"image/color"
	_ "image/jpeg"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Map struct {
	Image, ImageMask string
	ebitenImg        *ebiten.Image
	imgMask          image.Image
	safe             []bool
	mw, mh           int
}

// Draw applies the camera transform and dims the background so agents remain visible against the map.
func (m *Map) Draw(screen *ebiten.Image, cam *Camera, worldW, worldH int) {
	if m.ebitenImg == nil {
		var err error
		m.ebitenImg, _, err = ebitenutil.NewImageFromFile(m.Image)
		if err != nil {
			panic(err)
		}
	}

	imgW, imgH := m.ebitenImg.Bounds().Dx(), m.ebitenImg.Bounds().Dy()

	imgOptions := ebiten.DrawImageOptions{}
	imgOptions.GeoM.Scale(float64(worldW)/float64(imgW), float64(worldH)/float64(imgH))
	imgOptions.GeoM.Scale(float64(cam.Zoom), float64(cam.Zoom))
	sx, sy := cam.WorldToScreen(0, 0)
	imgOptions.GeoM.Translate(float64(sx), float64(sy))
	imgOptions.Filter = ebiten.FilterLinear
	imgOptions.ColorScale.SetWithColor(color.RGBA{R: 100, G: 100, B: 100, A: 200})
	screen.DrawImage(m.ebitenImg, &imgOptions)
}

// LoadMask converts the collision image into a boolean lookup table so movement checks can test safe cells without repeatedly decoding pixel colours.
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

// IsSafe rejects out-of-range coordinates before indexing the mask, preventing invalid grid access.
func (m *Map) IsSafe(x, y int) bool {

	if x < 0 || x >= m.mw {
		panic("x is out of bounds")
	}
	if y < 0 || y >= m.mh {
		panic("y is out of bounds")
	}

	return m.safe[y*m.mw+x]
}
