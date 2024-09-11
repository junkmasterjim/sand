/*
A basic implemenetation of a sandbox.
Fluid dynamics testing
*/

package main

import (
	"image/color"
	"log"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	MaxWaterPressure     = 7
	SurfaceTensionFactor = 0.3
)

const (
	Empty = iota // Empty is 0
	Sand         // Sand is 1
	Water        // Water is 2
)

type Game struct {
	grid  [][]int
	count int
}

const GAME_WIDTH int = 320
const GAME_HEIGHT int = 240
const WINDOW_SCALE int = 3

func (g *Game) HandleSand(x, y int) {
	// Check if we're at the bottom of the grid
	if y+1 >= len(g.grid[0]) {
		return
	}

	// Try to move straight down
	if g.grid[x][y+1] == Empty {
		g.grid[x][y+1] = Sand
		g.grid[x][y] = Empty
		return
	} else if g.grid[x][y+1] == Water {
		// Swap positions with water
		g.grid[x][y+1] = Sand
		g.grid[x][y] = Water
		return
	}

	// If can't move straight down, try diagonal movement
	leftX, rightX := x-1, x+1
	canMoveLeft := leftX >= 0 && y+1 < len(g.grid[0])
	canMoveRight := rightX < len(g.grid) && y+1 < len(g.grid[0])

	if canMoveLeft && (g.grid[leftX][y+1] == Empty || g.grid[leftX][y+1] == Water) {
		// Move sand to the left-down cell
		g.grid[leftX][y+1], g.grid[x][y] = Sand, g.grid[leftX][y+1]
	} else if canMoveRight && (g.grid[rightX][y+1] == Empty || g.grid[rightX][y+1] == Water) {
		// Move sand to the right-down cell
		g.grid[rightX][y+1], g.grid[x][y] = Sand, g.grid[rightX][y+1]
	}
	// If no movement is possible, the sand stays where it is
}

func (g *Game) HandleWater(x, y int) {
	if y+1 >= len(g.grid[0]) {
		return
	}

	// Calculate water pressure
	pressure := 0
	for i := y; i >= 0 && g.grid[x][i] == Water && pressure < MaxWaterPressure; i-- {
		pressure++
	}

	// Move down with pressure
	for i := 1; i <= pressure && y+i < len(g.grid[0]); i++ {
		if g.grid[x][y+i] == Empty {
			g.grid[x][y+i] = Water
			g.grid[x][y] = Empty
			return
		} else if g.grid[x][y+i] != Water {
			break
		}
	}

	// Horizontal flow with surface tension
	leftX, rightX := x-1, x+1
	canFlowLeft := leftX >= 0 && g.grid[leftX][y] == Empty
	canFlowRight := rightX < len(g.grid) && g.grid[rightX][y] == Empty

	if canFlowLeft && canFlowRight {
		if rand.Float64() < 0.5 {
			g.flowHorizontal(x, y, leftX)
		} else {
			g.flowHorizontal(x, y, rightX)
		}
	} else if canFlowLeft {
		g.flowHorizontal(x, y, leftX)
	} else if canFlowRight {
		g.flowHorizontal(x, y, rightX)
	}
}

func (g *Game) flowHorizontal(x, y, newX int) {
	// Check for surface tension
	if rand.Float64() < SurfaceTensionFactor {
		return
	}

	// Flow to the side
	g.grid[newX][y] = Water
	g.grid[x][y] = Empty

	// Equalizing - create a smoother water surface
	if y > 0 && g.grid[x][y-1] == Empty && g.grid[newX][y-1] == Water {
		g.grid[x][y-1] = Water
		g.grid[newX][y-1] = Empty
	}
}

func (g *Game) Pour(x, y int) {
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		particleType := Sand // Default to pouring sand
		if ebiten.IsKeyPressed(ebiten.KeyW) {
			particleType = Water // Pour water if 'W' key is pressed
		} else if ebiten.IsKeyPressed(ebiten.KeyE) {
			particleType = Empty // Clear if E is pressed
		}
		if x >= 0 && x < len(g.grid) && y >= 0 && y < len(g.grid[0]) {
			g.grid[x][y] = particleType
			// Create a small cluster of particles
			for dx := -3; dx <= 3; dx++ {
				for dy := 0; dy <= 4; dy++ {
					newX, newY := x+dx, y+dy
					if newX >= 0 && newX < len(g.grid) && newY >= 0 && newY < len(g.grid[0]) {
						// Add some randomness to create a more natural pour
						if rand.Float32() < 0.7 {
							g.grid[newX][newY] = particleType
						}
					}
				}
			}
		}
	}
}

func (g *Game) Update() error {
	mouseX, mouseY := ebiten.CursorPosition()
	g.Pour(mouseX, mouseY)

	// Multiple passes for smoother movement
	for pass := 0; pass < 5; pass++ {
		// Update water from bottom to top, and alternating left-to-right and right-to-left
		for y := len(g.grid[0]) - 1; y >= 0; y-- {
			if pass%2 == 0 {
				for x := 0; x < len(g.grid); x++ {
					if g.grid[x][y] == Water {
						g.HandleWater(x, y)
					}
				}
			} else {
				for x := len(g.grid) - 1; x >= 0; x-- {
					if g.grid[x][y] == Water {
						g.HandleWater(x, y)
					}
				}
			}
		}

		// Handle sand after water
		// alternate left and right to flow more evenly
		for y := len(g.grid[0]) - 1; y >= 0; y-- {
			if pass%2 == 0 {
				for x := 0; x < len(g.grid); x++ {
					if g.grid[x][y] == Sand {
						g.HandleSand(x, y)
					}
				}
			} else {
				for x := len(g.grid) - 1; x >= 0; x-- {
					if g.grid[x][y] == Sand {
						g.HandleSand(x, y)
					}
				}
			}
		}

	}

	return nil
}

func Init() *Game {
	g := make([][]int, GAME_WIDTH)
	for i := range g {
		g[i] = make([]int, GAME_HEIGHT)
	}

	for x := range g {
		for y := range g[x] {
			switch rand.IntN(10) {
			case 1:
				g[x][y] = Sand
			}
		}
	}

	return &Game{grid: g}
}

func (g *Game) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "Sandbox Simulation")

	for i := range g.grid {
		for j := range g.grid[i] {
			switch g.grid[i][j] {
			case Sand:
				vector.DrawFilledRect(screen, float32(i), float32(j), 1, 1, color.RGBA{203, 189, 147, 255}, false)
			case Water:
				vector.DrawFilledRect(screen, float32(i), float32(j), 1, 1, color.RGBA{173, 216, 255, 255}, false)
			}
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return GAME_WIDTH, GAME_HEIGHT
}

func main() {
	g := Init()

	ebiten.SetWindowSize(GAME_WIDTH*WINDOW_SCALE, GAME_HEIGHT*WINDOW_SCALE)
	ebiten.SetWindowTitle("Sandbox Simulation")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
