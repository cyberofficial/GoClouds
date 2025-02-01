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
	groundOffset = 20 // Offset for isometric perspective
	treeDepth    = 15 // How far below the horizon trees are planted
	shadowDepth  = 35 // How far down cloud shadows appear
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
	clouds                 []Cloud
	trees                  []Tree
	density                float64
	sunX, sunY             float64
	isDraggingSun          bool
	dragStartX, dragStartY float64
}

func NewGame() *Game {
	g := &Game{
		clouds:  make([]Cloud, maxClouds),
		trees:   make([]Tree, numTrees),
		density: 0.2, // Start with 20% density
		sunX:    float64(screenWidth - 100),
		sunY:    float64(80),
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

	// Handle sun dragging
	cursorX, cursorY := ebiten.CursorPosition()
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		// Check if click is within sun bounds
		dx := float64(cursorX) - g.sunX
		dy := float64(cursorY) - g.sunY
		if dx*dx+dy*dy <= sunRadius*sunRadius {
			g.isDraggingSun = true
			g.dragStartX = float64(cursorX) - g.sunX
			g.dragStartY = float64(cursorY) - g.sunY
		}
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && g.isDraggingSun {
		// Update sun position while dragging
		g.sunX = float64(cursorX) - g.dragStartX
		g.sunY = float64(cursorY) - g.dragStartY

		// Keep sun within screen bounds
		g.sunX = math.Max(sunRadius, math.Min(float64(screenWidth)-sunRadius, g.sunX))
		g.sunY = math.Max(sunRadius, math.Min(float64(screenHeight)*0.5, g.sunY))
	} else {
		g.isDraggingSun = false
	}

	return nil
}

func drawGround(screen *ebiten.Image) {
	// Draw main ground with isometric grid effect
	baseY := float64(screenHeight - groundHeight + groundOffset)

	// Base ground color
	ebitenutil.DrawRect(
		screen,
		0,
		baseY,
		float64(screenWidth),
		float64(groundHeight),
		color.RGBA{34, 139, 34, 255}, // Forest green
	)

	// Draw isometric grid
	gridSize := 40.0
	rows := int(groundHeight/gridSize) + 1
	cols := int(screenWidth/gridSize) + 2

	for row := 0; row < rows; row++ {
		for col := -1; col < cols; col++ {
			// Calculate isometric tile corners
			x1 := float64(col)*gridSize - (float64(row) * gridSize * 0.5)
			y1 := baseY + float64(row)*gridSize*0.5

			// Draw diagonal lines for isometric effect
			ebitenutil.DrawLine(
				screen,
				x1, y1,
				x1+gridSize, y1+gridSize*0.5,
				color.RGBA{24, 120, 24, 100},
			)
			ebitenutil.DrawLine(
				screen,
				x1+gridSize, y1+gridSize*0.5,
				x1+gridSize*2, y1,
				color.RGBA{44, 160, 44, 100},
			)
		}
	}
}

func drawTree(screen *ebiten.Image, tree Tree, sunX, sunY float64) {
	trunkWidth := tree.size * 0.2
	trunkHeight := tree.size * 0.4
	baseY := float64(screenHeight-groundHeight+groundOffset) + treeDepth // Move trees down into the ground

	// Draw tree shadow
	shadowLength := tree.size * 0.5
	shadowAngle := math.Atan2(baseY-sunY, tree.x-sunX) // Calculate shadow angle based on sun position
	for i := 0.0; i < shadowLength; i++ {
		alpha := uint8(40 * (1 - i/shadowLength))
		ebitenutil.DrawCircle(
			screen,
			tree.x+math.Cos(shadowAngle)*i*0.5,
			baseY+math.Sin(shadowAngle)*i*0.5-2,
			trunkWidth*0.6*(1-i/shadowLength),
			color.RGBA{0, 0, 0, alpha},
		)
	}

	// Draw trunk with 3D effect
	// Main trunk
	ebitenutil.DrawRect(
		screen,
		tree.x-trunkWidth/2,
		baseY-trunkHeight,
		trunkWidth,
		trunkHeight,
		color.RGBA{139, 69, 19, 255}, // Brown
	)

	// Trunk right shading
	ebitenutil.DrawRect(
		screen,
		tree.x+trunkWidth/2-2,
		baseY-trunkHeight,
		4,
		trunkHeight,
		color.RGBA{110, 50, 15, 255}, // Darker brown for depth
	)

	// Draw triangular tree top with 3D effect (3 segments for fuller look)
	for i := 0; i < 3; i++ {
		segment := float64(i)
		segmentHeight := tree.size * 0.4
		segmentWidth := tree.size * (1.0 - segment*0.2)

		top := baseY - trunkHeight - segmentHeight*(segment+1)
		bottom := baseY - trunkHeight - segmentHeight*segment

		// Draw filled triangle with gradient and side shading
		shade := uint8(tree.shade * 255)
		for y := bottom; y > top; y-- {
			progress := (bottom - y) / (bottom - top)
			width := segmentWidth * (1 - progress)

			// Main triangle body
			ebitenutil.DrawLine(
				screen,
				tree.x-width/2,
				y,
				tree.x+width/2,
				y,
				color.RGBA{0, shade, 0, 255},
			)

			// Right side shading for 3D effect
			rightShade := color.RGBA{0, uint8(float64(shade) * 0.7), 0, 255}
			ebitenutil.DrawLine(
				screen,
				tree.x+width/2,
				y,
				tree.x+width/2+5,
				y+2,
				rightShade,
			)
		}
	}
}

func (g *Game) drawSun(screen *ebiten.Image) {
	// Draw the main sun circle
	ebitenutil.DrawCircle(
		screen,
		g.sunX,
		g.sunY,
		sunRadius,
		color.RGBA{255, 220, 0, 255}, // Bright yellow
	)

	// Draw sun rays
	numRays := 12
	rayLength := float64(sunRadius) * 0.5

	for i := 0; i < numRays; i++ {
		angle := float64(i) * (2 * math.Pi / float64(numRays))
		endX := g.sunX + math.Cos(angle)*rayLength*1.5
		endY := g.sunY + math.Sin(angle)*rayLength*1.5
		startX := g.sunX + math.Cos(angle)*rayLength
		startY := g.sunY + math.Sin(angle)*rayLength

		ebitenutil.DrawLine(
			screen,
			startX,
			startY,
			endX,
			endY,
			color.RGBA{255, 220, 0, 255},
		)
	}

	// Draw drag indicator if sun is being hovered
	if g.isDraggingSun {
		ebitenutil.DrawCircle(
			screen,
			g.sunX,
			g.sunY,
			sunRadius+2,
			color.RGBA{255, 255, 255, 100},
		)
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Clear the screen with sky blue
	screen.Fill(color.RGBA{135, 206, 235, 255})

	// Draw the sun
	g.drawSun(screen)

	// Draw the ground
	drawGround(screen)

	// Draw cloud shadows first
	activeClouds := int(math.Floor(g.density * float64(len(g.clouds))))
	for i := 0; i < activeClouds; i++ {
		cloud := g.clouds[i]
		g.drawCloudShadow(screen, cloud)
	}

	// Draw trees
	for _, tree := range g.trees {
		drawTree(screen, tree, g.sunX, g.sunY)
	}

	// Draw clouds after trees
	for i := 0; i < activeClouds; i++ {
		cloud := g.clouds[i]
		drawCloud(screen, cloud)
	}

	// Draw density information
	ebitenutil.DebugPrint(screen, "Use Up/Down arrows to adjust cloud density")
}

func (g *Game) drawCloudShadow(screen *ebiten.Image, cloud Cloud) {
	groundHorizon := float64(screenHeight - groundHeight + groundOffset)
	// Calculate shadow position based on sun's position
	shadowOffsetX := (cloud.x - g.sunX) * 0.2
	shadowOffsetY := (cloud.y - g.sunY) * 0.3 // Increased Y offset effect
	baseY := groundHorizon + shadowDepth      // Base shadow position

	// Calculate shadow stretch based on cloud height
	heightFactor := cloud.y / screenHeight // 0 at top, 1 at bottom
	stretchX := 1.5 + heightFactor         // More stretch for higher clouds
	stretchY := 0.3 + heightFactor*0.2     // Flatter shadows for higher clouds

	// Adjust shadow angle based on sun position
	angleToSun := math.Atan2(cloud.y-g.sunY, cloud.x-g.sunX)
	shadowAngleAdjust := math.Sin(angleToSun) * 15 // Add some vertical displacement based on sun angle

	// Draw multiple overlapping shadow ellipses
	circles := []struct{ dx, dy float64 }{
		{0, 0},
		{cloud.size * 0.5, cloud.size * 0.1},
		{cloud.size * 0.3, -cloud.size * 0.1},
		{cloud.size * 0.7, cloud.size * 0.05},
	}

	for _, c := range circles {
		shadowX := cloud.x + shadowOffsetX + c.dx
		shadowY := baseY + shadowOffsetY*0.3 + c.dy + shadowAngleAdjust
		shadowSizeX := cloud.size * 0.4 * stretchX
		shadowSizeY := cloud.size * 0.4 * stretchY

		// Draw multiple thin ellipses to create elongated shadow
		steps := 10
		for i := 0; i < steps; i++ {
			progress := float64(i) / float64(steps)
			currentSize := shadowSizeX * (1 - progress*0.5)
			currentY := shadowY + progress*shadowSizeY

			// Skip drawing if the shadow line would be above the ground horizon
			if currentY < groundHorizon {
				continue
			}

			// Fade out shadows more quickly near the horizon
			fadeOffset := 1.0
			if currentY-groundHorizon < 20 {
				fadeOffset = (currentY - groundHorizon) / 20
			}

			ebitenutil.DrawLine(
				screen,
				shadowX-currentSize,
				currentY,
				shadowX+currentSize,
				currentY,
				color.RGBA{
					0, 0, 0,
					uint8(cloud.opacity * 40 * (1 - progress) * fadeOffset), // Fade out towards edges and near horizon
				},
			)
		}
	}
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
