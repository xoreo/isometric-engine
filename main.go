package main

import (
	"fmt"
	"image"
	_ "image/png"
	"os"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/golang/geo/r2"
	"golang.org/x/image/colornames"
)

type tileType int

const (
	blank      tileType = iota
	grass      tileType = iota
	stone      tileType = iota
	selected   tileType = iota
	stoneEdgeN tileType = iota
	stoneEdgeE tileType = iota
	stoneEdgeS tileType = iota
	stoneEdgeW tileType = iota
)

const (
	worldSizeX = 10
	worldSizeY = 10
)

var (
	worldSize = pixel.V(worldSizeX, worldSizeY)
	tileSize  = pixel.V(63, 32)
	origin    = pixel.V(5, 1)
	world     [worldSizeX][worldSizeY]tileType
)

// loadPicture loads a picture from memory and returns a pixel picture.
func loadPicture(path string) (pixel.Picture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil
}

// pointToScreenSpace takes coordinates from the world space and maps them to
// coordinates in the virtual screen space.
func pointToScreenSpace(x, y float64) pixel.Vec {
	return pixel.V(
		(origin.X*tileSize.X+(x-y)*(tileSize.X/2))+tileSize.X/2,
		(origin.Y*tileSize.Y+(x+y)*(tileSize.Y/2))+tileSize.Y/2,
	)
}

// run is the main game function.
func run() {
	// Create the window config
	cfg := pixelgl.WindowConfig{
		Title: "@xoreo isometric-engine",
		Bounds: pixel.R(
			0,
			0,
			(worldSizeX+2)*tileSize.X,
			(worldSizeY)*tileSize.X,
		),
	}

	// Create the window itself
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	// Initialize the sprites
	spriteSheet, err := loadPicture("resources/spritesheet.png")
	if err != nil {
		panic(err)
	}

	var tileSprites [6]*pixel.Sprite

	tileSprites[grass] = pixel.NewSprite(spriteSheet, pixel.R(257, 67, tileSize.X, tileSize.Y))
	tileSprites[stone] = pixel.NewSprite(spriteSheet, pixel.R(1, 34, tileSize.X, tileSize.Y))
	tileSprites[selected] = pixel.NewSprite(spriteSheet, pixel.R(
		1, 1, tileSize.X, tileSize.Y,
	))

	// Initialize the world map to blank tiles
	for y, _ := range world {
		for x, _ := range world[y] {
			world[y][x] = grass
		}
	}

	// Main loop
	for !win.Closed() {
		// Clear the screen
		win.Clear(colornames.White)

		mouseVec := win.MousePosition() // Get the position of the mouse
		boardSpaceCell := pixel.V(
			float64(int(mouseVec.X)/int(tileSize.X)), // x position
			float64(int(mouseVec.Y)/int(tileSize.Y)), // y position
		)

		// Map the cell coords in screen space to those in cell space
		cellSpaceCell := pixel.V(
			(boardSpaceCell.Y-origin.Y)+(boardSpaceCell.X-origin.X),
			(boardSpaceCell.Y-origin.Y)-(boardSpaceCell.X-origin.X),
		)

		// Render all of the tiles, y first to add depth
		for y := 0; y < worldSizeY; y++ {
			for x := 0; x < worldSizeX; x++ {
				// Map to screen space
				screenVec := pointToScreenSpace(float64(x), float64(y))
				switch world[x][y] {
				case grass:
					// Draw the grass tile sprite
					tileSprites[grass].Draw(win, pixel.IM.Moved(screenVec))
					break
				}
			}
		}

		imd := imdraw.New(nil)           // Initialize the mesh
		imd.Color = pixel.RGB(255, 0, 0) // Red

		// Calculate where the point is in relation to the border of the tile
		tx := tileSize.X
		ty := tileSize.Y
		P := r2.Point{mouseVec.X, mouseVec.Y}
		O := r2.Point{boardSpaceCell.X * tx, boardSpaceCell.Y * ty}
		A := r2.Point{
			O.X + tx/2,
			O.Y,
		}
		B := r2.Point{
			O.X,
			O.Y + ty/2,
		}
		C := r2.Point{
			O.X + tx/2,
			O.Y + ty,
		}
		D := r2.Point{
			O.X + tx,
			O.Y + ty/2,
		}

		// Calculate the cross products
		dAB := (P.X-A.X)*(B.Y-A.Y) - (P.Y-A.Y)*(B.X-A.X)
		dBC := (P.X-B.X)*(C.Y-B.Y) - (P.Y-B.Y)*(C.X-B.X)
		dCD := (P.X-C.X)*(D.Y-C.Y) - (P.Y-C.Y)*(D.X-C.X)
		dDA := (P.X-D.X)*(A.Y-D.Y) - (P.Y-D.Y)*(A.X-D.X)
		fmt.Printf("dAB: %f\ndBC: %f\ndCD: %f\ndDA: %f\n\n", dAB, dBC, dCD, dDA)

		// Change the cellSpaceCell accordingly
		if dAB < 0 { // Bottom left
			cellSpaceCell.X -= 1
		} else if dBC < 0 { // Top left
			cellSpaceCell.Y += 1
		} else if dCD < 0 { // Top right
			cellSpaceCell.X += 1
		} else if dDA < 0 { // Bottom right
			cellSpaceCell.Y -= 1
		}

		// Check that the cell is within the board
		if cellSpaceCell.X >= 0 && cellSpaceCell.X < worldSizeX { // Check x bounds
			if cellSpaceCell.Y >= 0 && cellSpaceCell.Y < worldSizeY { // Check y bounds
				tileSprites[selected].Draw(win, pixel.IM.Moved(
					pointToScreenSpace(cellSpaceCell.X, cellSpaceCell.Y),
				)) // Draw the highlighted sprite on the cell
			}
		}

		win.Update() // Update the window
	}
}

func main() {
	pixelgl.Run(run) // Set the run function = my run function
}
