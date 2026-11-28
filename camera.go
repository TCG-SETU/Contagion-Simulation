package main

import "github.com/hajimehoshi/ebiten/v2"

const (
	minZoom = 0.25
	maxZoom = 10.0
	keyPan = 8.0
)

type Camera struct {
	X, Y float32 
	Zoom float32 
}


// NewCamera chooses the largest zoom that fits the world on screen initially, giving the player a useful overview without manual adjustment.
func NewCamera(worldW, worldH int) *Camera {
	z := min(float32(screenWidth)/float32(worldW), float32(screenHeight)/float32(worldH))
	c := &Camera{Zoom: z}
	c.clamp(worldW, worldH)
	return c
}

// WorldToScreen centralises the coordinate conversion so drawing code stays consistent with camera position and zoom.
func (c *Camera) WorldToScreen(wx, wy float32) (float32, float32) {
	return (wx - c.X) * c.Zoom, (wy - c.Y) * c.Zoom
}

// Visible checks screen bounds before drawing so off-screen objects can be skipped.
func (c *Camera) Visible(wx, wy, size float32) bool {
	sx, sy := (wx-c.X)*c.Zoom, (wy-c.Y)*c.Zoom
	s := size * c.Zoom
	return sx+s >= 0 && sy+s >= 0 && sx-s <= screenWidth && sy-s <= screenHeight
}

// Update applies keyboard movement and zoom, then clamps the camera so the view cannot drift beyond the world.
func (c *Camera) Update(worldW, worldH int) {
	step := keyPan / c.Zoom

	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		c.X -= step
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		c.X += step
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		c.Y += step
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		c.Y -= step
	}
	if ebiten.IsKeyPressed(ebiten.KeyMinus) {
		c.Zoom = max(minZoom, min(maxZoom, c.Zoom-0.01))
	}
	if ebiten.IsKeyPressed(ebiten.KeyEqual) {
		c.Zoom = max(minZoom, min(maxZoom, c.Zoom+0.01))
	}

	c.clamp(worldW, worldH)
}

// clamp keeps camera coordinates within valid limits so the visible area remains aligned with the world.
func (c *Camera) clamp(worldW, worldH int) {
	viewW := float32(screenWidth) / c.Zoom
	viewH := float32(screenHeight) / c.Zoom

	if viewW >= float32(worldW) {
		c.X = (float32(worldW) - viewW) / 2
	} else {
		c.X = max(0, min(c.X, float32(worldW)-viewW))
	}
	if viewH >= float32(worldH) {
		c.Y = (float32(worldH) - viewH) / 2
	} else {
		c.Y = max(0, min(c.Y, float32(worldH)-viewH))
	}
}
