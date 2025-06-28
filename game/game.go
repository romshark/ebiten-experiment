package game

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Config struct {
	MinZoom, MaxZoom        int
	ZoomSpeed, ZoomFriction int
	PanSpeed, PanFriction   int
}

type Camera struct {
	X, Y             float64 // World position
	Zoom             float64
	VelX, VelY       float64 // Velocity for smooth movement
	ZoomVel          float64 // Zoom velocity
	screenW, screenH int
}

type World struct {
	width, height int
	tiles         [][]Tile
}

type Input struct {
	mouseX, mouseY         int
	prevMouseX, prevMouseY int
	mousePressed           bool
	wasdVelX, wasdVelY     float64
	// For anchor-based dragging
	dragStartMouseX, dragStartMouseY   int
	dragStartCameraX, dragStartCameraY float64
	isDragging                         bool
	// For momentum calculation
	dragVelX, dragVelY float64
}

type Game struct {
	world                     *World
	camera                    *Camera
	input                     *Input
	config                    *Config
	screenWidth, screenHeight int
}

func NewCamera(screenW, screenH int, worldW, worldH int) *Camera {
	return &Camera{
		X:       float64(worldW*TileSize) / 2, // Start at center of world
		Y:       float64(worldH*TileSize) / 2,
		Zoom:    1.0,
		screenW: screenW,
		screenH: screenH,
	}
}

func (c *Camera) Update(config *Config, input *Input, worldW, worldH int) {
	// Handle WASD movement (smooth with velocity)
	if !input.isDragging {
		c.VelX += input.wasdVelX * float64(config.PanSpeed) * 0.5
		c.VelY += input.wasdVelY * float64(config.PanSpeed) * 0.5

		// Apply friction to velocity
		friction := float64(config.PanFriction) * 0.01
		c.VelX *= (1.0 - friction)
		c.VelY *= (1.0 - friction)

		// Update position with velocity
		c.X += c.VelX
		c.Y += c.VelY
	}

	// Handle anchor-based mouse dragging
	if input.isDragging {
		if input.dragStartMouseX == input.mouseX && input.dragStartMouseY == input.mouseY {
			// Just started dragging - record the camera position
			input.dragStartCameraX = c.X
			input.dragStartCameraY = c.Y
		} else {
			// Calculate current camera position based on total drag distance
			mouseDeltaX := float64(input.mouseX - input.dragStartMouseX)
			mouseDeltaY := float64(input.mouseY - input.dragStartMouseY)

			worldDeltaX := -mouseDeltaX / c.Zoom
			worldDeltaY := -mouseDeltaY / c.Zoom

			newX := input.dragStartCameraX + worldDeltaX
			newY := input.dragStartCameraY + worldDeltaY

			// Calculate drag velocity for momentum (difference from last frame)
			input.dragVelX = newX - c.X
			input.dragVelY = newY - c.Y

			c.X = newX
			c.Y = newY
		}
	} else if input.dragVelX != 0 || input.dragVelY != 0 {
		// Just stopped dragging - transfer drag velocity to camera velocity for momentum
		c.VelX = input.dragVelX * 8.0 // Amplify for nice momentum feel
		c.VelY = input.dragVelY * 8.0
		input.dragVelX = 0
		input.dragVelY = 0
	}

	// Handle zoom
	c.ZoomVel *= (1.0 - float64(config.ZoomFriction)*0.01)
	c.Zoom += c.ZoomVel

	// Clamp zoom
	minZoom := float64(config.MinZoom) * 0.1
	maxZoom := float64(config.MaxZoom) * 0.1
	if c.Zoom < minZoom {
		c.Zoom = minZoom
	}
	if c.Zoom > maxZoom {
		c.Zoom = maxZoom
	}

	// Simple bounds checking - keep camera center within world
	worldWidth := float64(worldW * TileSize)
	worldHeight := float64(worldH * TileSize)

	if c.X < 0 {
		c.X = 0
		c.VelX = 0
	}
	if c.X > worldWidth {
		c.X = worldWidth
		c.VelX = 0
	}
	if c.Y < 0 {
		c.Y = 0
		c.VelY = 0
	}
	if c.Y > worldHeight {
		c.Y = worldHeight
		c.VelY = 0
	}
}

func (c *Camera) UpdateScreenSize(w, h int) {
	c.screenW = w
	c.screenH = h
}

func NewWorld(width, height int) *World {
	// Ensure textures are initialized
	if grassTextures == nil {
		InitGrassTextures()
	}

	w := &World{
		width:  width,
		height: height,
		tiles:  make([][]Tile, height),
	}

	// Generate tiles with random texture indices
	for y := 0; y < height; y++ {
		w.tiles[y] = make([]Tile, width)
		for x := 0; x < width; x++ {
			w.tiles[y][x] = Tile{
				textureIndex: rand.Intn(NumGrassVariations),
			}
		}
	}

	return w
}

func NewInput() *Input {
	return &Input{}
}

func (i *Input) Update() {
	// Update mouse
	i.prevMouseX, i.prevMouseY = i.mouseX, i.mouseY
	i.mouseX, i.mouseY = ebiten.CursorPosition()

	// Handle drag start/stop
	mouseJustPressed := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	mouseJustReleased := inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)
	i.mousePressed = ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)

	if mouseJustPressed {
		// Start dragging - record anchor point
		i.isDragging = true
		i.dragStartMouseX = i.mouseX
		i.dragStartMouseY = i.mouseY
		// dragStartCamera will be set in Camera.Update
	}

	if mouseJustReleased {
		i.isDragging = false
	}

	// Update WASD
	i.wasdVelX = 0
	i.wasdVelY = 0

	if ebiten.IsKeyPressed(ebiten.KeyA) {
		i.wasdVelX = -1
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		i.wasdVelX = 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		i.wasdVelY = -1
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		i.wasdVelY = 1
	}
}

func (i *Input) HandleZoom(camera *Camera, config *Config) {
	_, yOffset := ebiten.Wheel()
	if yOffset != 0 {
		zoomFactor := float64(config.ZoomSpeed) * 0.001
		camera.ZoomVel += yOffset * zoomFactor
	}
}

func (g *Game) Update() error {
	// Update input
	g.input.Update()
	g.input.HandleZoom(g.camera, g.config)

	// Update camera
	g.camera.Update(g.config, g.input, g.world.width, g.world.height)

	return nil
}

// Update the Draw method to use the texture index
func (g *Game) Draw(screen *ebiten.Image) {
	// Clear screen
	screen.Fill(color.RGBA{50, 50, 50, 255})

	// Use current screen dimensions from Layout
	screenW := float64(g.screenWidth)
	screenH := float64(g.screenHeight)

	// Calculate visible tile range
	leftWorld := g.camera.X - screenW/(2*g.camera.Zoom)
	rightWorld := g.camera.X + screenW/(2*g.camera.Zoom)
	topWorld := g.camera.Y - screenH/(2*g.camera.Zoom)
	bottomWorld := g.camera.Y + screenH/(2*g.camera.Zoom)

	// Convert to tile coordinates
	startTileX := int(math.Floor(leftWorld / TileSize))
	endTileX := int(math.Ceil(rightWorld / TileSize))
	startTileY := int(math.Floor(topWorld / TileSize))
	endTileY := int(math.Ceil(bottomWorld / TileSize))

	// Clamp to world bounds
	if startTileX < 0 {
		startTileX = 0
	}
	if endTileX > g.world.width {
		endTileX = g.world.width
	}
	if startTileY < 0 {
		startTileY = 0
	}
	if endTileY > g.world.height {
		endTileY = g.world.height
	}

	// Draw visible tiles
	for y := startTileY; y < endTileY; y++ {
		for x := startTileX; x < endTileX; x++ {
			// World position of tile (top-left corner)
			worldX := float64(x * TileSize)
			worldY := float64(y * TileSize)

			// Convert to screen coordinates
			screenX := (worldX-g.camera.X)*g.camera.Zoom + screenW/2
			screenY := (worldY-g.camera.Y)*g.camera.Zoom + screenH/2

			// Get the texture for this tile
			textureIndex := g.world.tiles[y][x].textureIndex
			texture := grassTextures[textureIndex]

			// Draw tile
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(g.camera.Zoom, g.camera.Zoom)
			op.GeoM.Translate(screenX, screenY)

			screen.DrawImage(texture, op)
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	// Update screen size and camera - ensure this happens every time
	g.screenWidth = outsideWidth
	g.screenHeight = outsideHeight
	g.camera.UpdateScreenSize(outsideWidth, outsideHeight)

	// Return the exact dimensions requested
	return outsideWidth, outsideHeight
}

func DefaultConfig() *Config {
	return &Config{
		MinZoom:      5,  // 0.5x zoom
		MaxZoom:      50, // 5.0x zoom
		ZoomSpeed:    10,
		ZoomFriction: 20,
		PanSpeed:     1,
		PanFriction:  32,
	}
}

func NewGame(worldSizeX, worldSizeY int, c *Config) *Game {
	// Set default config if nil
	if c == nil {
		c = DefaultConfig()
	}

	return &Game{
		world:        NewWorld(worldSizeX, worldSizeY),
		camera:       NewCamera(800, 600, worldSizeX, worldSizeY), // Will be updated in Layout
		input:        NewInput(),
		config:       c,
		screenWidth:  800, // Will be updated in Layout
		screenHeight: 600, // Will be updated in Layout
	}
}

const (
	TileSize           = 32
	NumGrassVariations = 8 // Number of precomputed grass textures
)

// Global texture cache
var grassTextures []*ebiten.Image

type Tile struct {
	textureIndex int // Index into grassTextures array
}

// Initialize grass textures - call this once at startup
func InitGrassTextures() {
	grassTextures = make([]*ebiten.Image, NumGrassVariations)

	for i := 0; i < NumGrassVariations; i++ {
		grassTextures[i] = generateGrassVariation(i)
	}
}

func generateGrassVariation(seed int) *ebiten.Image {
	// Use seed for deterministic variation
	r := rand.New(rand.NewSource(int64(seed + 12345)))

	img := ebiten.NewImage(TileSize, TileSize)

	// Base grass color with slight variation per texture
	baseHue := 34 + r.Intn(10) - 5     // Slight red variation
	baseGreen := 139 + r.Intn(30) - 15 // Green variation
	baseBlue := 34 + r.Intn(8) - 4     // Slight blue variation

	baseColor := color.RGBA{
		R: uint8(clamp(baseHue, 0, 255)),
		G: uint8(clamp(baseGreen, 0, 255)),
		B: uint8(clamp(baseBlue, 0, 255)),
		A: 255,
	}

	// Fill with base color
	img.Fill(baseColor)

	// Add texture noise - vary amount per variation
	noiseAmount := 40 + r.Intn(20)
	for i := 0; i < noiseAmount; i++ {
		x := r.Intn(TileSize)
		y := r.Intn(TileSize)

		// Vary the green slightly
		variation := r.Intn(40) - 20
		green := int(baseColor.G) + variation

		pixelColor := color.RGBA{
			R: baseColor.R,
			G: uint8(clamp(green, 0, 255)),
			B: baseColor.B,
			A: 255,
		}

		// Draw small grass-like pixels
		vector.DrawFilledRect(img, float32(x), float32(y), 1, 1, pixelColor, false)
	}

	// Add grass blades - vary amount per variation
	bladeAmount := 15 + r.Intn(10)
	for i := 0; i < bladeAmount; i++ {
		x := r.Intn(TileSize-2) + 1
		y := r.Intn(TileSize-4) + 2

		// Vary blade color slightly
		darkGreen := color.RGBA{
			R: uint8(clamp(20+r.Intn(10), 0, 255)),
			G: uint8(clamp(100+r.Intn(20), 0, 255)),
			B: uint8(clamp(20+r.Intn(10), 0, 255)),
			A: 255,
		}

		// Vary blade size slightly
		bladeHeight := 2 + r.Intn(2)
		vector.DrawFilledRect(img, float32(x), float32(y), 1, float32(bladeHeight), darkGreen, false)
	}

	return img
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
