package main

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (g *Game) DrawEditor(screen *ebiten.Image) {
	screen.Fill(color.Black)
	ebitenutil.DebugPrint(screen, "editor\n")

	op := &ebiten.DrawImageOptions{}
	for _, brick := range g.editorleveldata.bricks {
		DrawBrick(screen, float64(brick.x), float64(brick.y), brick.rotation, op, brick)
	}
}

func (g *Game) UpdateEditor() {
	x, y := ebiten.CursorPosition()
	if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		new_brick := &BrickData{
			x:         x,
			y:         y,
			rotation:  0,
			colour:    "red",
			bricktype: "normal",
		}
		g.editorleveldata.bricks = append(g.editorleveldata.bricks, new_brick)

		new_brick.x = int(math.Round(float64(new_brick.x)/UNIT)) * UNIT
		new_brick.y = int(math.Round(float64(new_brick.y)/UNIT)) * UNIT
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		for _, brick := range g.editorleveldata.bricks {
			if brick.bricktype == "" {
				brick.bricktype = "normal"
			}
			brick.x = int(math.Round(float64(brick.x)/UNIT)) * UNIT
			brick.y = int(math.Round(float64(brick.y)/UNIT)) * UNIT
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		for _, brick := range g.editorleveldata.bricks {
			fmt.Printf(`{x: %d, y: %d, rotation: %f, colour: "%s", bricktype: "%s"},`, brick.x, brick.y, brick.rotation, brick.colour, brick.bricktype)
		}
		fmt.Println("")
	}
	minDist := math.Inf(1)
	var closestBrick *BrickData
	var closestBrickIndex int
	for i, brick := range g.editorleveldata.bricks {
		distFromMouse := math.Sqrt(math.Pow(float64(brick.x-x), 2) + math.Pow(float64(brick.y-y), 2))
		if minDist > distFromMouse {
			minDist = distFromMouse
			closestBrick = brick
			closestBrickIndex = i
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.Key0) {
		closestBrick.x = int(math.Round(float64(closestBrick.x)/UNIT)) * UNIT
		closestBrick.y = int(math.Round(float64(closestBrick.y)/UNIT)) * UNIT
	}

	if inpututil.IsKeyJustPressed(ebiten.Key0) && ebiten.IsKeyPressed(ebiten.KeyControl) {
		closestBrick.x = int(math.Round(float64(closestBrick.x)/float64(brick_blue.Bounds().Dx())) * float64(brick_blue.Bounds().Dx()))
		closestBrick.y = int(math.Round(float64(closestBrick.y)/float64(brick_blue.Bounds().Dy()-UNIT)) * float64(brick_blue.Bounds().Dy()-UNIT))
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
		closestBrick.x -= UNIT
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
		closestBrick.x += UNIT
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		if ebiten.IsKeyPressed(ebiten.KeyControl) {
			closestBrick.y -= brick_blue.Bounds().Dy() - UNIT
		} else {
			closestBrick.y -= UNIT
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		if ebiten.IsKeyPressed(ebiten.KeyControl) {
			closestBrick.y += brick_blue.Bounds().Dy() - UNIT
		} else {
			closestBrick.y += UNIT
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyMinus) {
		closestBrick.rotation -= math.Pi / 2
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEqual) {
		closestBrick.rotation += math.Pi / 2
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyN) {
		closestBrick.rotation = 0
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		g.editorleveldata.bricks = removeBrick(g.editorleveldata.bricks, closestBrickIndex)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF5) {
		g.ResetEditor()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.editor = false
	}
	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		closestBrick.colour = "red"
	}
	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		closestBrick.colour = "yellow"
	}
	if inpututil.IsKeyJustPressed(ebiten.Key3) {
		closestBrick.colour = "green"
	}
	if inpututil.IsKeyJustPressed(ebiten.Key4) {
		closestBrick.colour = "blue"
	}
	if inpututil.IsKeyJustPressed(ebiten.Key5) {
		closestBrick.colour = "black"
	}
	if inpututil.IsKeyJustPressed(ebiten.Key6) {
		closestBrick.colour = "white"
	}
	if inpututil.IsKeyJustPressed(ebiten.Key7) {
		closestBrick.colour = "grey"
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyZ) {
		closestBrick.bricktype = "small"
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyX) {
		closestBrick.bricktype = "regular"
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyJ) {
		g.editorleveldata.bricks = append(g.editorleveldata.bricks,
			&BrickData{
				x:        closestBrick.x - brick_blue.Bounds().Dx(),
				y:        closestBrick.y,
				rotation: closestBrick.rotation,
				colour:   closestBrick.colour,
			},
		)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyK) {
		g.editorleveldata.bricks = append(g.editorleveldata.bricks,
			&BrickData{
				x:        closestBrick.x + brick_blue.Bounds().Dx(),
				y:        closestBrick.y,
				rotation: closestBrick.rotation,
				colour:   closestBrick.colour,
			},
		)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyI) {
		offset := brick_blue.Bounds().Dy() - UNIT
		if ebiten.IsKeyPressed(ebiten.KeyControl) {
			offset *= 2
		}
		g.editorleveldata.bricks = append(g.editorleveldata.bricks,
			&BrickData{
				x:        closestBrick.x,
				y:        closestBrick.y - offset,
				rotation: closestBrick.rotation,
				colour:   closestBrick.colour,
			},
		)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyM) {
		offset := brick_blue.Bounds().Dy() - UNIT
		if ebiten.IsKeyPressed(ebiten.KeyControl) {
			offset *= 2
		}
		g.editorleveldata.bricks = append(g.editorleveldata.bricks,
			&BrickData{
				x:        closestBrick.x,
				y:        closestBrick.y + offset,
				rotation: closestBrick.rotation,
				colour:   closestBrick.colour,
			},
		)
	}
}

func (g *Game) ResetEditor() {
	g.editorleveldata = &EditorLevel
}

func removeBrick(s []*BrickData, i int) []*BrickData {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
