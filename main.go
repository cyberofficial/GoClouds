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
	sunRadius    = 40
	groundHeight = 150
	numTrees     = 5
)

type Cloud struct {
	x, y    float64
	speed   float64
	size    float64
	opacity float64
}

type Tree struct {
	x     float64
	size  float64
	shade float64
}

type Game struct {
	clouds  []Cloud
	trees   []Tree
	density float64
}

func NewGame() *Game {
	g := &Game{
		clouds:  make([]Cloud, maxClouds),
		trees:   make([]Tree, numTrees),
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

	// Initialize trees with random properties
	spacing := float64(screenWidth) / float64(numTrees+1)
	for i := range g.trees {
		g.trees[i] = Tree{
			x:     spacing * float64(i+1),
			size:  50 + rand.Float64()*30,   // Random size between 50-80
			shade: 0.7 + rand.Float64()*0.3, // Random shade variation
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

func drawGround(screen *ebiten.Image) {
	// Draw main ground
	ebitenutil.DrawRect(
		screen,
		0,
		float64(screenHeight-groundHeight),
		float64(screenWidth),
		float64(groundHeight),
		color.RGBA{34, 139, 34, 255}, // Forest green
	)
}

func drawTree(screen *ebiten.Image, tree Tree) {
	trunkWidth := tree.size * 0.2
	trunkHeight := tree.size * 0.4

	// Draw trunk
	ebitenutil.DrawRect(
		screen,
		tree.x-trunkWidth/2,
		float64(screenHeight-groundHeight)-trunkHeight,
		trunkWidth,
		trunkHeight,
		color.RGBA{139, 69, 19, 255}, // Brown
	)

	// Draw triangular tree top (3 segments for fuller look)
	for i := 0; i < 3; i++ {
		segment := float64(i)
		segmentHeight := tree.size * 0.4
		segmentWidth := tree.size * (1.0 - segment*0.2)

		vertices := []struct{ x, y float64 }{
			{tree.x, float64(screenHeight-groundHeight) - trunkHeight - segmentHeight*(segment+1)},              // Top
			{tree.x - segmentWidth/2, float64(screenHeight-groundHeight) - trunkHeight - segmentHeight*segment}, // Bottom left
			{tree.x + segmentWidth/2, float64(screenHeight-groundHeight) - trunkHeight - segmentHeight*segment}, // Bottom right
		}

		// Draw filled triangle
		shade := uint8(tree.shade * 255)
		for y := vertices[1].y; y > vertices[0].y; y-- {
			progress := (vertices[1].y - y) / (vertices[1].y - vertices[0].y)
			width := segmentWidth * (1 - progress)
			ebitenutil.DrawLine(
				screen,
				tree.x-width/2,
				y,
				tree.x+width/2,
				y,
				color.RGBA{0, shade, 0, 255},
			)
		}
	}
}

func drawSun(screen *ebiten.Image) {
	// Draw the main sun circle
	ebitenutil.DrawCircle(
		screen,
		float64(screenWidth-100), // Position from right
		80,                       // Position from top
		sunRadius,
		color.RGBA{255, 220, 0, 255}, // Bright yellow
	)

	// Draw sun rays
	numRays := 12
	rayLength := float64(sunRadius) * 0.5
	centerX := float64(screenWidth - 100)
	centerY := float64(80)

	for i := 0; i < numRays; i++ {
		angle := float64(i) * (2 * math.Pi / float64(numRays))
		endX := centerX + math.Cos(angle)*rayLength*1.5
		endY := centerY + math.Sin(angle)*rayLength*1.5
		startX := centerX + math.Cos(angle)*rayLength
		startY := centerY + math.Sin(angle)*rayLength

		ebitenutil.DrawLine(
			screen,
			startX,
			startY,
			endX,
			endY,
			color.RGBA{255, 220, 0, 255},
		)
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Clear the screen with sky blue
	screen.Fill(color.RGBA{135, 206, 235, 255})

	// Draw the sun
	drawSun(screen)

	// Draw the ground
	drawGround(screen)

	// Draw trees
	for _, tree := range g.trees {
		drawTree(screen, tree)
	}

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
