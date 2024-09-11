/*
A basic implemenetation of a sandbox.
id like to flesh this out much more
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

/*
	Handle the gravity for sand.
	Takes an x,y coordinate as ints for an argument
	to get the sand coordinates
*/

func (g *Game) HandleSand(x, y int) {
	// Check if we're at the bottom of the grid
	if y+1 >= len(g.grid[0]) {
		return
	}

	// Function to move water to an adjacent empty cell
	moveWater := func(waterX, waterY int) bool {
		for _, dx := range []int{-1, 1} {
			newX := waterX + dx
			if newX >= 0 && newX < len(g.grid) && g.grid[newX][waterY] == Empty {
				g.grid[newX][waterY] = Water
				return true
			}
		}
		return false
	}

	// Try to move straight down
	if g.grid[x][y+1] == Empty {
		g.grid[x][y+1] = Sand
		g.grid[x][y] = Empty
		return
	} else if g.grid[x][y+1] == Water {
		if moveWater(x, y+1) {
			g.grid[x][y+1] = Sand
			g.grid[x][y] = Empty
		} else {
			g.grid[x][y+1] = Sand
			g.grid[x][y] = Water
		}
		return
	}

	// If can't move straight down, try diagonal movement
	leftX, rightX := x-1, x+1
	canMoveLeft := leftX >= 0 && y+1 < len(g.grid[0])
	canMoveRight := rightX < len(g.grid) && y+1 < len(g.grid[0])

	if canMoveLeft && (g.grid[leftX][y+1] == Empty || g.grid[leftX][y+1] == Water) {
		if g.grid[leftX][y+1] == Water && !moveWater(leftX, y+1) {
			g.grid[leftX][y+1] = Sand
			g.grid[x][y] = Water
		} else {
			g.grid[leftX][y+1] = Sand
			g.grid[x][y] = Empty
		}
	} else if canMoveRight && (g.grid[rightX][y+1] == Empty || g.grid[rightX][y+1] == Water) {
		if g.grid[rightX][y+1] == Water && !moveWater(rightX, y+1) {
			g.grid[rightX][y+1] = Sand
			g.grid[x][y] = Water
		} else {
			g.grid[rightX][y+1] = Sand
			g.grid[x][y] = Empty
		}
	}
	// If no movement is possible, the sand stays where it is
}

/*
Handle the gravity for water
Takes x & y as int for location of water
*/

func (g *Game) HandleWater(x, y int) {
	// Check if we're at the bottom of the grid
	if y+1 >= len(g.grid[0]) {
		return
	}

	// Try to move down
	if g.grid[x][y+1] == Empty {
		g.grid[x][y] = Empty
		g.grid[x][y+1] = Water
		return
	}

	// If can't move down, try to spread horizontally
	leftX, rightX := x-1, x+1
	canMoveLeft := leftX >= 0 && g.grid[leftX][y] == Empty
	canMoveRight := rightX < len(g.grid) && g.grid[rightX][y] == Empty

	if canMoveLeft && canMoveRight {
		// Randomly choose left or right
		if rand.Float32() < 0.5 {
			g.grid[x][y] = Empty
			g.grid[leftX][y] = Water
		} else {
			g.grid[x][y] = Empty
			g.grid[rightX][y] = Water
		}
	} else if canMoveLeft {
		g.grid[x][y] = Empty
		g.grid[leftX][y] = Water
	} else if canMoveRight {
		g.grid[x][y] = Empty
		g.grid[rightX][y] = Water
	}
	// If can't move in any direction, the water stays where it is
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
		for x := range g.grid {
			for y := len(g.grid[x]) - 1; y >= 0; y-- {
				switch g.grid[x][y] {
				case Sand:
					g.HandleSand(x, y)
				case Water:
					g.HandleWater(x, y)
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
				vector.DrawFilledRect(screen, float32(i), float32(j), 1, 1, color.RGBA{128, 128, 128, 255}, false)
			case Water:
				vector.DrawFilledRect(screen, float32(i), float32(j), 1, 1, color.RGBA{228, 228, 228, 255}, false)
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
