package main

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"sort"

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
	x, y  float64
	size  float64
	shade float64
}

type Menu struct {
	visible      bool
	treeDensity  int
	cloudCount   int
	maxClouds    int
	selectedTree int     // -1 when no tree is selected
	treeShadow   float64 // new: shadow scale factor (e.g., 1.0 default)
}

type Game struct {
	clouds                 []Cloud
	trees                  []Tree
	density                float64
	sunX, sunY             float64
	isDraggingSun          bool
	dragStartX, dragStartY float64
	menu                   Menu
	draggedTree            int // -1 when no tree is being dragged
	dragTreeStartX         float64
}

func NewGame() *Game {
	g := &Game{
		clouds:      make([]Cloud, maxClouds),
		trees:       make([]Tree, numTrees),
		density:     0.2, // Start with 20% density
		sunX:        float64(screenWidth - 100),
		sunY:        float64(80),
		draggedTree: -1,
		menu: Menu{
			visible:      false,
			treeDensity:  numTrees,
			cloudCount:   maxClouds,
			maxClouds:    maxClouds,
			selectedTree: -1,
			treeShadow:   1.0, // new default shadow value
		},
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
	for i := range g.trees {
		baseY := float64(screenHeight-groundHeight+groundOffset) + treeDepth
		g.trees[i] = Tree{
			x:     50 + rand.Float64()*float64(screenWidth-100), // Random position with margin
			y:     baseY,
			size:  50 + rand.Float64()*30,   // Random size between 50-80
			shade: 0.7 + rand.Float64()*0.3, // Random shade variation
		}
	}

	return g
}

func (g *Game) Update() error {
	// Check for escape key to close window
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	// Toggle menu with M key
	if inpututil.IsKeyJustPressed(ebiten.KeyM) {
		g.menu.visible = !g.menu.visible
	}

	// cloud positions in a single loop
	for i := range g.clouds {
		g.clouds[i].x += g.clouds[i].speed
		if g.clouds[i].x > screenWidth+100 {
			g.clouds[i].x = -100
		}
	}

	// Handle menu controls when visible
	if g.menu.visible {
		// Adjust tree density with up/down arrows
		if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
			g.menu.treeDensity = min(20, g.menu.treeDensity+1)
			g.updateTreeCount()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
			g.menu.treeDensity = max(1, g.menu.treeDensity-1)
			g.updateTreeCount()
		}

		// Adjust cloud count with left/right arrows
		if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
			g.menu.cloudCount = max(0, g.menu.cloudCount-10)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
			g.menu.cloudCount = min(g.menu.maxClouds, g.menu.cloudCount+10)
		}

		// New: Adjust tree shadow value with S (decrease) and D (increase)
		if inpututil.IsKeyJustPressed(ebiten.KeyS) {
			g.menu.treeShadow = math.Max(0.2, g.menu.treeShadow-0.1)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyD) {
			g.menu.treeShadow = math.Min(2.0, g.menu.treeShadow+0.1)
		}
	} else {
		// Original density controls when menu is hidden
		if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
			g.density = math.Min(1.0, g.density+0.1)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
			g.density = math.Max(0.0, g.density-0.1)
		}
	}

	cursorX, cursorY := ebiten.CursorPosition()

	// Handle mouse input
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		// Check for sun dragging first
		dx := float64(cursorX) - g.sunX
		dy := float64(cursorY) - g.sunY
		if dx*dx+dy*dy <= sunRadius*sunRadius {
			g.isDraggingSun = true
			g.dragStartX = float64(cursorX) - g.sunX
			g.dragStartY = float64(cursorY) - g.sunY
		} else {
			// Check for tree dragging
			for i, tree := range g.trees {
				// Expand hitbox to include both trunk and tree crown
				dx := float64(cursorX) - tree.x
				crownTop := tree.y - tree.size*1.2 // Account for full tree height
				if math.Abs(dx) < tree.size*0.4 && float64(cursorY) >= crownTop && float64(cursorY) <= tree.y {
					g.draggedTree = i
					g.dragTreeStartX = float64(cursorX) - tree.x
					break
				}
			}
		}
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if g.isDraggingSun {
			// Update sun position while dragging
			g.sunX = float64(cursorX) - g.dragStartX
			g.sunY = float64(cursorY) - g.dragStartY

			// Keep sun within screen bounds
			g.sunX = math.Max(sunRadius, math.Min(float64(screenWidth)-sunRadius, g.sunX))
			g.sunY = math.Max(sunRadius, math.Min(float64(screenHeight)*0.5, g.sunY))
		} else if g.draggedTree != -1 {
			// Update tree position while dragging
			newX := float64(cursorX) - g.dragTreeStartX
			newY := float64(cursorY)
			groundY := float64(screenHeight - groundHeight + groundOffset)

			// Allow free movement but keep tree below ground line
			if newY >= groundY {
				g.trees[g.draggedTree].x = newX
				g.trees[g.draggedTree].y = newY
			}
		}
	} else {
		g.isDraggingSun = false
		g.draggedTree = -1
	}

	return nil
}

func (g *Game) updateTreeCount() {
	// Update tree count based on density setting
	oldTrees := g.trees
	g.trees = make([]Tree, g.menu.treeDensity)

	// Keep existing trees if possible
	for i := range g.trees {
		if i < len(oldTrees) {
			g.trees[i] = oldTrees[i]
		} else {
			// Initialize new tree with random position
			baseY := float64(screenHeight-groundHeight+groundOffset) + treeDepth
			g.trees[i] = Tree{
				x:     50 + rand.Float64()*float64(screenWidth-100), // Random position with margin
				y:     baseY,
				size:  50 + rand.Float64()*30,
				shade: 0.7 + rand.Float64()*0.3,
			}
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
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

// --- Modify drawTree to accept the shadow factor ---
func drawTree(screen *ebiten.Image, tree Tree, sunX, sunY, treeShadow float64) {
	trunkWidth := tree.size * 0.2
	trunkHeight := tree.size * 0.4

	// Calculate distance and angle to sun
	dx := tree.x - sunX
	dy := tree.y - sunY
	distanceToSun := math.Sqrt(dx*dx + dy*dy)
	shadowAngle := math.Atan2(tree.y-sunY, tree.x-sunX)

	// Calculate distance factor (shadows get longer when sun is closer)
	maxDistance := math.Sqrt(float64(screenWidth*screenWidth + screenHeight*screenHeight))
	distanceFactor := math.Max(0.5, 1.0-distanceToSun/maxDistance) * 2.0

	// Calculate shadow length based on sun height and distance
	sunHeight := screenHeight - sunY
	heightFactor := math.Max(0.2, sunHeight/screenHeight) // Prevents extremely short shadows when sun is at bottom
	baseShadowLength := tree.size * 2.0                   // Base shadow length

	// Shadow gets longer as sun gets lower and closer to horizon
	shadowLength := baseShadowLength * (1 / heightFactor) * distanceFactor

	// Shadow gets shorter when sun is directly overhead
	verticalAngleFactor := math.Abs(math.Sin(shadowAngle))
	shadowLength *= (0.3 + 0.7*verticalAngleFactor) // Maintains minimum shadow length

	// Calculate shadow length and apply treeShadow factor
	shadowLength *= treeShadow // new scaling for tree shadows

	// Draw shadow with dynamic length and width
	for i := 0.0; i < shadowLength; i++ {
		progress := i / shadowLength
		alpha := uint8(50 * (1 - progress))
		shadowWidth := trunkWidth * 0.6 * (1 - progress*0.8) // Maintain some minimum width

		ebitenutil.DrawCircle(
			screen,
			tree.x+math.Cos(shadowAngle)*i*0.8, // Increased step size for longer shadows
			tree.y+math.Sin(shadowAngle)*i*0.8-2,
			shadowWidth,
			color.RGBA{0, 0, 0, alpha},
		)
	}

	// Draw trunk with 3D effect
	// Main trunk
	ebitenutil.DrawRect(
		screen,
		// Main trunk
		tree.x-trunkWidth/2,
		tree.y-trunkHeight,
		trunkWidth,
		trunkHeight,
		color.RGBA{139, 69, 19, 255}, // Brown
	)

	// Trunk right shading
	ebitenutil.DrawRect(
		screen,
		tree.x+trunkWidth/2-2,
		tree.y-trunkHeight,
		4,
		trunkHeight,
		color.RGBA{110, 50, 15, 255}, // Darker brown for depth
	)

	// Draw triangular tree top with 3D effect (3 segments for fuller look)
	for i := 0; i < 3; i++ {
		segment := float64(i)
		segmentHeight := tree.size * 0.4
		segmentWidth := tree.size * (1.0 - segment*0.2)

		top := tree.y - trunkHeight - segmentHeight*(segment+1)
		bottom := tree.y - trunkHeight - segmentHeight*segment

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
	var activeClouds int
	if g.menu.visible {
		activeClouds = g.menu.cloudCount
	} else {
		activeClouds = int(math.Floor(g.density * float64(len(g.clouds))))
	}

	for i := 0; i < activeClouds && i < len(g.clouds); i++ {
		cloud := g.clouds[i]
		g.drawCloudShadow(screen, cloud)
	}

	// Sort trees by Y position so trees closer to bottom are drawn last (appear on top)
	sortedTrees := make([]Tree, len(g.trees))
	copy(sortedTrees, g.trees)
	sort.Slice(sortedTrees, func(i, j int) bool {
		return sortedTrees[i].y < sortedTrees[j].y
	})

	// Draw trees with current shadow factor
	for _, tree := range sortedTrees {
		drawTree(screen, tree, g.sunX, g.sunY, g.menu.treeShadow)
	}

	// Draw clouds after trees
	for i := 0; i < activeClouds && i < len(g.clouds); i++ {
		cloud := g.clouds[i]
		g.drawCloud(screen, cloud)
	}

	if g.menu.visible {
		// Draw semi-transparent overlay
		ebitenutil.DrawRect(
			screen,
			10,
			10,
			240,
			180,
			color.RGBA{0, 0, 0, 180},
		)

		// Draw menu content
		y := 20
		ebitenutil.DebugPrintAt(screen, "=== Environment Controls ===", 15, y)
		y += 20
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Tree Count: %d (Up/Down)", g.menu.treeDensity), 15, y)
		y += 20
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Cloud Count: %d (Left/Right)", g.menu.cloudCount), 15, y)
		y += 20
		ebitenutil.DebugPrintAt(screen, "Controls:", 15, y)
		y += 20
		ebitenutil.DebugPrintAt(screen, "- M: Toggle Menu", 15, y)
		y += 20
		ebitenutil.DebugPrintAt(screen, "- LMB: Drag Sun/Trees", 15, y)
		y += 20
		ebitenutil.DebugPrintAt(screen, "- S/D: Change Tree Shadow intensity", 15, y)
		y += 20
		ebitenutil.DebugPrintAt(screen, "- ESC: Exit", 15, y)
	} else {
		// Draw basic controls when menu is hidden
		ebitenutil.DebugPrint(screen, "Press M for environment controls\nLMB to drag sun/trees\nPress ESC to exit")
	}
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

func (g *Game) drawCloud(screen *ebiten.Image, cloud Cloud) {
	// Calculate distance from sun to cloud
	dx := cloud.x - g.sunX
	dy := cloud.y - g.sunY
	distanceToSun := math.Sqrt(dx*dx + dy*dy)
	maxDistance := math.Sqrt(float64(screenWidth*screenWidth + screenHeight*screenHeight))
	sunlightFactor := math.Max(0, 1-(distanceToSun/maxDistance)) // 1 when close to sun, 0 when far

	// Calculate angle to sun for directional lighting
	angleToSun := math.Atan2(dy, dx)

	// Draw multiple overlapping circles to create a cloud shape
	circles := []struct{ dx, dy float64 }{
		{0, 0},
		{cloud.size * 0.5, cloud.size * 0.1},
		{cloud.size * 0.3, -cloud.size * 0.1},
		{cloud.size * 0.7, cloud.size * 0.05},
	}

	for _, c := range circles {
		// Calculate how lit this part of the cloud is based on its position relative to the sun
		relativeAngle := math.Atan2(c.dy, c.dx) - angleToSun
		lightingFactor := 0.7 + 0.3*math.Cos(relativeAngle) // Creates subtle variation based on position relative to sun

		// Calculate base color with slight yellow tint from sun
		baseR := uint8(255)
		baseG := uint8(255)
		baseB := uint8(255)

		// Add yellow tint based on sun proximity
		yellowTint := uint8(25 * sunlightFactor) // Max yellow tint of 25
		baseR = uint8(math.Min(float64(baseR+yellowTint), 255))
		baseG = uint8(math.Min(float64(baseG+yellowTint), 255))
		baseB = uint8(math.Min(float64(baseB), 255)) // Keep blue unchanged for slight yellow effect

		// Apply lighting factor
		finalR := uint8(float64(baseR) * lightingFactor)
		finalG := uint8(float64(baseG) * lightingFactor)
		finalB := uint8(float64(baseB) * lightingFactor)

		ebitenutil.DrawCircle(
			screen,
			cloud.x+c.dx,
			cloud.y+c.dy,
			cloud.size*0.3,
			color.RGBA{
				finalR,
				finalG,
				finalB,
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
		if err != ebiten.Termination {
			panic(err)
		}
	}
}
