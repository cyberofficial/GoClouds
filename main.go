package main

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	screenWidth  = 800
	screenHeight = 600
	maxClouds    = 100
)

type Cloud struct {
	x, y    float64
	speed   float64
	size    float64
	opacity float64
}

type Game struct {
	clouds  []Cloud
	density float64
}

func NewGame() *Game {
	g := &Game{
		clouds:  make([]Cloud, maxClouds),
		density: 0.2, // Start with 20% density
	}

	// Initialize clouds with random properties
	for i := range g.clouds {
		g.clouds[i] = Cloud{
			x:       rand.Float64() * screenWidth,
			y:       rand.Float64() * screenHeight * 0.6, // Keep clouds in upper 60% of screen
			speed:   1 + rand.Float64()*2,                // Random speed between 1-3
			size:    30 + rand.Float64()*50,              // Random size between 30-80
			opacity: 0.3 + rand.Float64()*0.5,            // Random opacity between 0.3-0.8
		}
	}

	return g
}

func (g *Game) Update() error {
	// Update cloud positions
	for i := range g.clouds {
		g.clouds[i].x += g.clouds[i].speed
		if g.clouds[i].x > screenWidth+100 {
			g.clouds[i].x = -100
		}
	}

	// Adjust density with up/down arrows
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		g.density = math.Min(1.0, g.density+0.1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		g.density = math.Max(0.0, g.density-0.1)
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Clear the screen with sky blue
	screen.Fill(color.RGBA{135, 206, 235, 255})

	// Draw only the number of clouds based on current density
	activeClouds := int(math.Floor(g.density * float64(len(g.clouds))))

	for i := 0; i < activeClouds; i++ {
		cloud := g.clouds[i]
		// Draw a fluffy cloud using multiple overlapping circles
		drawCloud(screen, cloud)
	}

	// Draw density information
	ebitenutil.DebugPrint(screen, "Use Up/Down arrows to adjust cloud density")
}

func drawCloud(screen *ebiten.Image, cloud Cloud) {
	// Draw multiple overlapping circles to create a cloud shape
	circles := []struct{ dx, dy float64 }{
		{0, 0},
		{cloud.size * 0.5, cloud.size * 0.1},
		{cloud.size * 0.3, -cloud.size * 0.1},
		{cloud.size * 0.7, cloud.size * 0.05},
	}

	for _, c := range circles {
		ebitenutil.DrawCircle(
			screen,
			cloud.x+c.dx,
			cloud.y+c.dy,
			cloud.size*0.3,
			color.RGBA{
				255, 255, 255,
				uint8(cloud.opacity * 255),
			},
		)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Cloud Generation")

	game := NewGame()
	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
