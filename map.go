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
	Image     string
	ImageMask string
	ebitenImg *ebiten.Image
	imgMask   image.Image
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
	defer mapImageMaskFile.Close()

	m.imgMask, _, err = image.Decode(mapImageMaskFile)
	if err != nil {
		panic(err)
	}
}

func (m *Map) IsSafe(x, y int) bool {
	bounds := m.imgMask.Bounds()

	if x < 0 || x >= bounds.Dx() {
		panic("x is out of bounds")
	}
	if y < 0 || y >= bounds.Dy() {
		panic("y is out of bounds")
	}

	if r, g, b, _ := m.imgMask.At(x, y).RGBA(); r == 0 && g == 0 && b == 0 {
		return false
	}

	return true
}
