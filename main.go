package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand/v2"

	"github.com/hajimehoshi/bitmapfont/v3"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jakecoffman/cp"
)

const (
	width        = 32.0
	height       = 14.0
	screenWidth  = 600
	screenHeight = 480
	UNIT         = 4
)

var (
	fontFace        = text.NewGoXFace(bitmapfont.Face)
	big_brick       = LoadImageFromPath("assets/big_brick.png")
	laser           = LoadImageFromPath("assets/laser.png")
	dynamite        = LoadImageFromPath("assets/dynamite.png")
	dynamite_object = LoadImageFromPath("assets/dynamite_object.png")
	brick_red       = LoadImageFromPath("assets/brick_red.png")
	brick_green     = LoadImageFromPath("assets/brick_green.png")
	brick_yellow    = LoadImageFromPath("assets/brick_yellow.png")
	brick_black     = LoadImageFromPath("assets/brick_black.png")
	brick_white     = LoadImageFromPath("assets/brick_white.png")
	brick_blue      = LoadImageFromPath("assets/brick_blue.png")
	brick_grey      = LoadImageFromPath("assets/brick_gray.png")
	title           = LoadImageFromPath("assets/title.png")
	endtitle1       = LoadImageFromPath("assets/endtitle1.png")
	endtitle2       = LoadImageFromPath("assets/endtitle2.png")
	endtitle3       = LoadImageFromPath("assets/endtitle3.png")
	endtitle4       = LoadImageFromPath("assets/endtitle4.png")
	brickTextures   = map[string]*ebiten.Image{
		"red":    brick_red,
		"green":  brick_green,
		"yellow": brick_yellow,
		"black":  brick_black,
		"white":  brick_white,
		"blue":   brick_blue,
		"grey":   brick_grey,
	}
	smallBrickTextures = map[string]*ebiten.Image{
		"red":    makeBrickSmall(brick_red),
		"green":  makeBrickSmall(brick_green),
		"yellow": makeBrickSmall(brick_yellow),
		"black":  makeBrickSmall(brick_black),
		"white":  makeBrickSmall(brick_white),
		"blue":   makeBrickSmall(brick_blue),
		"grey":   makeBrickSmall(brick_grey),
	}
	item_slot         = LoadImageFromPath("assets/item_slot.png")
	target            = LoadImageFromPath("assets/target.png")
	bg                = LoadImageFromPath("assets/bg.png")
	restart           = LoadImageFromPath("assets/restart.png")
	fire_button       = ebiten.NewImage(64, 36)
	fire_button_hover = ebiten.NewImage(64, 36)
	brick1            = ReadOggBytesFromPath("assets/brick1.ogg")
	brick2            = ReadOggBytesFromPath("assets/brick2.ogg")
	brick3            = ReadOggBytesFromPath("assets/brick3.ogg")
	brick4            = ReadOggBytesFromPath("assets/brick4.ogg")
	brick5            = ReadOggBytesFromPath("assets/brick5.ogg")
	success           = ReadOggBytesFromPath("assets/jingles_PIZZI10.ogg")
	fail              = ReadOggBytesFromPath("assets/jingles_PIZZI07.ogg")
	laser_sound       = ReadOggBytesFromPath("assets/laserLarge_002.ogg")
	explode_sound     = ReadOggBytesFromPath("assets/explosionCrunch_002.ogg")
	brick_sounds      = [][]byte{
		brick1,
		brick2,
		brick3,
		brick4,
		brick5,
	}
	impactTypeBrick cp.CollisionType = 1
)

type Game struct {
	space             *cp.Space
	level             int
	editor            bool
	editorleveldata   *LevelData
	leveldata         *LevelData
	levelSuccess      bool
	levelComplete     bool
	levelCompleteTime int
	objects           []*Object
	mouseX, mouseY    int
	canFire           bool
	objectIndex       int
	isMainMenu        bool
	walls             bool
	t                 int
	audio_context     *audio.Context
	in_ui             bool
	gameOver          bool
}

type Object struct {
	objectType string
	body       *cp.Body
	lifetime   int
}

func makeBrickSmall(image *ebiten.Image) *ebiten.Image {
	new := ebiten.NewImage(brick_green.Bounds().Dx()/2, brick_black.Bounds().Dy())

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-float64(brick_black.Bounds().Dx())/2, 0)
	new.DrawImage(image, op)
	return new
}

func init() {
	fire_button.Fill(color.Black)

	op2 := &text.DrawOptions{}
	op2.GeoM.Scale(2, 2)
	op2.GeoM.Translate(5, 2)
	op2.ColorScale.ScaleWithColor(color.White)
	text.Draw(fire_button, "Fire!", fontFace, op2)

	fire_button_hover.Fill(color.RGBA{60, 60, 60, 255})

	op2 = &text.DrawOptions{}
	op2.GeoM.Scale(2, 2)
	op2.GeoM.Translate(5, 2)
	op2.ColorScale.ScaleWithColor(color.White)
	text.Draw(fire_button_hover, "Fire!", fontFace, op2)
}

func NewGame() *Game {
	g := &Game{
		level:         0,
		editor:        false,
		levelComplete: false,
		objects:       []*Object{},
		canFire:       false,
		objectIndex:   0,
		isMainMenu:    true,
		walls:         true,
		audio_context: audio.NewContext(sampleRate),
	}
	LoadLevel(g, LevelMainMenu)
	g.editorleveldata = &EditorLevel

	return g
}

func (g *Game) Update() error {
	g.mouseX, g.mouseY = ebiten.CursorPosition()
	g.t += 1

	if g.gameOver {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton0) {
			g.gameOver = false
			g.level = 0
			LoadRelevantLevel(g)
			return nil
		}
	}

	g.in_ui = false
	if !g.isMainMenu && g.mouseX > (screenWidth-40) && g.mouseY < 40 {
		g.in_ui = true
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton0) {
			LoadRelevantLevel(g)
		}
	}

	if g.editor {
		g.UpdateEditor()
		return nil
	}

	if g.levelComplete {
		g.levelCompleteTime += 1
	}

	if g.levelComplete && g.levelCompleteTime > 100 {
		g.levelComplete = false
		if g.isMainMenu {
			g.isMainMenu = false
		} else {
			if g.levelSuccess {
				g.level += 1
			}
		}
		LoadRelevantLevel(g)
	}

	if g.levelComplete {
		return nil
	}

	g.space.Step(1.0 / float64(ebiten.TPS()))

	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		LoadRelevantLevel(g)
	}

	for i, brick := range g.leveldata.bricks {
		body := brick.body
		if body.Position().Y > screenHeight+20 || body.Position().X < 0 || body.Position().X > screenWidth {
			if g.isMainMenu {
				removeBrick(g.leveldata.bricks, i)
				break
			} else {
				body.SetPosition(cp.Vector{X: screenWidth / 2, Y: -40})
				body.SetVelocity(0, 0)
			}
		}
	}

	attack_vector := getAttackVector(g)

	all_sleeping := true
	all_below_target_height := true
	for _, block := range g.leveldata.bricks {
		body := block.body
		body.EachShape(func(s *cp.Shape) {
			verts := getPolygonVertices(s)
			for _, v := range verts {
				if v.Y < float64(g.leveldata.targetHeight) {
					all_below_target_height = false
				}
			}
		})
		if !body.IsSleeping() {
			all_sleeping = false
		}
	}

	for objectindex, object := range g.objects {
		object.lifetime += 1
		body := object.body
		if !body.IsSleeping() {
			all_sleeping = false
		}

		if object.objectType == "dynamite" && object.lifetime > 50 {
			for _, brick := range g.leveldata.bricks {
				vec := brick.body.Position().Clone().Sub(body.Position()).Normalize()
				pow := 1 / brick.body.Position().Sub(body.Position()).Length()
				brick.body.SetVelocityVector(vec.Mult(pow * 10000))
			}
			g.objects = removeObject(g.objects, objectindex)

			sePlayer := g.audio_context.NewPlayerFromBytes(explode_sound)
			sePlayer.SetVolume(
				1,
			)
			sePlayer.Play()
		}
	}

	all_objects_fired := g.objectIndex >= len(g.leveldata.objects)

	if all_objects_fired {
		g.canFire = false
	}

	if all_sleeping && !g.isMainMenu {
		if all_below_target_height {
			g.levelCompleteTime = 0
			g.levelComplete = true
			g.levelSuccess = true
			sePlayer := g.audio_context.NewPlayerFromBytes(success)
			sePlayer.SetVolume(
				0.2,
			)

			sePlayer.Play()
		} else if all_objects_fired {
			g.levelCompleteTime = 0
			g.levelComplete = true
			g.levelSuccess = false
			sePlayer := g.audio_context.NewPlayerFromBytes(fail)
			sePlayer.SetVolume(
				0.2,
			)

			sePlayer.Play()
		}
	}

	if g.isMainMenu && g.t < 15000 {
		if rand.Float64() > 0.95 {
			var colour string
			rand2 := rand.Float64()
			if rand2 > 0.6 {
				colour = "red"
			} else if rand2 > 0.4 {
				colour = "green"
			} else if rand2 > 0.2 {
				colour = "blue"
			} else {
				colour = "yellow"
			}
			brick := &BrickData{
				x:         int(rand.Float64() * screenWidth),
				y:         -20,
				colour:    colour,
				rotation:  rand.Float64() * math.Pi * 2,
				bricktype: "normal",
			}
			createBrick(g.space, brick)
			brick.body.SetVelocity((rand.Float64()-0.5)*500, rand.Float64()*1000)
			g.leveldata.bricks = append(g.leveldata.bricks, brick)
		}
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton0) && !all_objects_fired && !g.in_ui {
		nextobj := g.leveldata.objects[g.objectIndex]

		if nextobj == "laser" {
			for _, block := range g.leveldata.bricks {
				body := block.body

				start := cp.Vector{X: screenWidth / 2, Y: 0}
				end := cp.Vector{X: screenWidth/2 + attack_vector.X, Y: attack_vector.Y}

				body.EachShape(func(shape *cp.Shape) {
					vertices := getPolygonVertices(shape)
					for _, vertex := range vertices {
						if pointOnLineSegment(vertex, start, end) {
							body.SetVelocity((rand.Float64()-0.5)*10000, (rand.Float64()-0.5)*10000)
							break
						}
					}
				})

			}
			sePlayer := g.audio_context.NewPlayerFromBytes(laser_sound)
			sePlayer.SetVolume(
				1,
			)
			sePlayer.Play()
		} else if nextobj == "big_brick" {
			g.objects = append(g.objects, &Object{
				objectType: "big_brick",
				body:       fireBigBrick(g.space, g),
			})
		} else if nextobj == "dynamite" {
			g.objects = append(g.objects, &Object{
				objectType: "dynamite",
				body:       fireDynamite(g.space, g),
			})
		}
		g.objectIndex += 1
	}

	if g.isMainMenu {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton0) {
			g.levelComplete = true
		}
	}

	return nil
}

func getAttackVector(g *Game) cp.Vector {
	attack_vector := cp.Vector{
		X: float64(g.mouseX) - screenWidth/2,
		Y: float64(g.mouseY),
	}

	vector_scale := math.Abs((screenWidth / 2) / attack_vector.X)

	// if math.Abs(attack_vector.Y*vector_scale) > screenHeight {
	// 	vector_scale := math.Abs(screenHeight / attack_vector.Y)

	// 	attack_vector.X = attack_vector.X * vector_scale
	// 	attack_vector.Y = attack_vector.Y * vector_scale
	// } else {

	attack_vector.X = attack_vector.X * vector_scale
	attack_vector.Y = attack_vector.Y * vector_scale
	// }

	return attack_vector
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.editor {
		g.DrawEditor(screen)
		return
	}

	if g.gameOver {
		screen.Fill(color.Black)

		if g.t > 100 {
			screen.DrawImage(endtitle1, nil)
		}
		if g.t > 200 {
			screen.DrawImage(endtitle2, nil)
		}
		if g.t > 300 {
			screen.DrawImage(endtitle3, nil)
		}
		if g.t > 400 {
			screen.DrawImage(endtitle4, nil)
		}

		return
	}

	if g.levelComplete {
		screen.Fill(color.Black)

		op2 := &text.DrawOptions{}
		op2.PrimaryAlign = text.AlignCenter
		op2.GeoM.Scale(2, 2)
		op2.GeoM.Translate(screenWidth/2, screenHeight/2-50)
		op2.ColorScale.ScaleWithColor(color.White)

		if g.isMainMenu {
			text.Draw(screen, "Level 1", fontFace, op2)
		} else if g.levelSuccess {
			text.Draw(screen, "Level Complete!", fontFace, op2)
		} else {
			text.Draw(screen, "Level Failed :(", fontFace, op2)
		}

		return
	}

	all_objects_fired := g.objectIndex >= len(g.leveldata.objects)

	screen.Fill(color.RGBA{39, 117, 166, 255})
	screen.DrawImage(bg, nil)

	op := &ebiten.DrawImageOptions{}
	op.ColorScale.Scale(230.0/255.0, 230.0/255.0, 230.0/255.0, 1)

	if g.canFire && !g.in_ui && len(g.leveldata.objects) > 0 && !all_objects_fired {
		attack_vector := getAttackVector(g)
		nextobj := g.leveldata.objects[g.objectIndex]

		var path vector.Path
		path.MoveTo(screenWidth/2, 0)
		path.LineTo(screenWidth/2+float32(attack_vector.X), float32(attack_vector.Y))
		path.Close()
		if nextobj == "laser" {
			StrokePath(screen, &path, color.RGBA{255, 0, 0, 255}, 4, 0, 0)
			StrokePath(screen, &path, color.RGBA{255, 255, 255, 255}, 2, 0, 0)
		} else {
			StrokePath(screen, &path, color.RGBA{255, 255, 255, 255}, 4, 0, 0)
		}

	}

	for _, block := range g.leveldata.bricks {
		body := block.body
		if block.body == nil {
			createBrick(g.space, block)
		} else {
			DrawBrick(
				screen,
				body.Position().X,
				body.Position().Y,
				body.Angle(),
				op,
				block,
			)
		}
	}

	for _, block := range g.objects {
		body := block.body

		var image *ebiten.Image

		if block.objectType == "dynamite" {
			image = dynamite_object
		} else {
			image = big_brick
		}

		op.GeoM.Reset()
		op.GeoM.Translate(-float64(image.Bounds().Dx())/2, -float64(image.Bounds().Dy())/2)

		op.GeoM.Rotate(body.Angle())
		op.GeoM.Translate(body.Position().X, body.Position().Y)

		screen.DrawImage(image, op)
	}

	if !g.isMainMenu {
		op = &ebiten.DrawImageOptions{}
		op.GeoM.Translate(5, 40)
		screen.DrawImage(item_slot, op)
	}

	for i, object := range g.leveldata.objects {
		if i < g.objectIndex {
			continue
		}
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(5, 40)
		op.GeoM.Translate(float64(i-g.objectIndex)*48, 0)
		op.GeoM.Translate(float64(item_slot.Bounds().Dx())/2, float64(item_slot.Bounds().Dy())/2)

		if object == "big_brick" {
			op.GeoM.Translate(-float64(big_brick.Bounds().Dx())/2, -float64(big_brick.Bounds().Dy())/2)
			screen.DrawImage(big_brick, op)
		} else if object == "laser" {
			op.GeoM.Translate(-float64(laser.Bounds().Dx())/2, -float64(laser.Bounds().Dy())/2)
			screen.DrawImage(laser, op)
		} else if object == "dynamite" {
			op.GeoM.Translate(-float64(dynamite.Bounds().Dx())/2, -float64(dynamite.Bounds().Dy())/2)
			screen.DrawImage(dynamite, op)
		}
	}

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, float64(g.leveldata.targetHeight))
	screen.DrawImage(target, op)

	if !g.isMainMenu {
		op2 := &text.DrawOptions{}
		op2.GeoM.Scale(2, 2)
		op2.GeoM.Translate(5, 5)
		op2.ColorScale.ScaleWithColor(color.White)
		text.Draw(screen, fmt.Sprint("Level ", g.level+1), fontFace, op2)

		op = &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(screenWidth-restart.Bounds().Dx()), 0)
		screen.DrawImage(restart, op)
	}

	if g.isMainMenu {
		screen.DrawImage(title, nil)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func DrawBrick(screen *ebiten.Image, x, y, angle float64, op *ebiten.DrawImageOptions, brickdata *BrickData) {
	var brick *ebiten.Image
	var exists bool
	colour := brickdata.colour

	if brickdata.bricktype == "small" {
		brick, exists = smallBrickTextures[colour]
		if !exists {
			brick = smallBrickTextures["red"]
		}
	} else {
		brick, exists = brickTextures[colour]
		if !exists {
			brick = brick_red
		}
	}

	op.GeoM.Reset()
	op.GeoM.Translate(-float64(brick.Bounds().Dx())/2, -float64(brick.Bounds().Dy())/2)

	op.GeoM.Rotate(angle)
	op.GeoM.Translate(x, y)

	screen.DrawImage(brick, op)
}

func LoadRelevantLevel(g *Game) {
	if g.level == 0 {
		LoadLevel(g, Level0)
	} else if g.level == 1 {
		LoadLevel(g, Level1)
	} else if g.level == 2 {
		LoadLevel(g, Level2)
	} else if g.level == 3 {
		LoadLevel(g, Level3)
	} else if g.level == 4 {
		LoadLevel(g, Level4)
	} else if g.level == 5 {
		LoadLevel(g, Level5)
	} else if g.level == 6 {
		LoadLevel(g, Level6)
	} else if g.level == 7 {
		LoadLevel(g, Level7)
	} else if g.level == 8 {
		LoadLevel(g, Level8)
	} else if g.level == 9 {
		LoadLevel(g, Level9)
	} else {
		g.gameOver = true
		g.t = 0
		g.level = 0
		LoadLevel(g, Level0)
	}
}

func getPolygonVertices(shape *cp.Shape) []cp.Vector {
	poly := shape.Class.(*cp.PolyShape) // Type assertion to PolyShape
	vertexCount := poly.Count()
	body := shape.Body()

	worldVertices := make([]cp.Vector, vertexCount)
	for i := 0; i < vertexCount; i++ {
		localVertex := poly.Vert(i)
		worldVertices[i] = body.LocalToWorld(localVertex)
	}
	return worldVertices
}

func LoadLevel(g *Game, level LevelData) {
	space := cp.NewSpace()
	space.Iterations = 30
	space.SetGravity(cp.Vector{X: 0, Y: 300})
	space.SleepTimeThreshold = 0.5
	space.SetCollisionSlop(0.5)

	floor := space.AddShape(cp.NewSegment(
		space.StaticBody,
		cp.Vector{X: -600, Y: screenHeight - 2},
		cp.Vector{X: 600, Y: screenHeight - 2},
		0,
	))
	floor.SetElasticity(1)
	floor.SetFriction(1)
	floor.SetCollisionType(impactTypeBrick)

	if g.walls {
		leftwall := space.AddShape(cp.NewSegment(
			space.StaticBody,
			cp.Vector{X: 0, Y: 0},
			cp.Vector{X: 0, Y: screenHeight},
			0,
		))
		leftwall.SetElasticity(1)
		leftwall.SetFriction(1)
		leftwall.SetCollisionType(impactTypeBrick)

		rightwall := space.AddShape(cp.NewSegment(
			space.StaticBody,
			cp.Vector{X: screenWidth, Y: 0},
			cp.Vector{X: screenWidth, Y: screenHeight},
			0,
		))
		rightwall.SetElasticity(1)
		rightwall.SetFriction(1)
		rightwall.SetCollisionType(impactTypeBrick)
	}

	// target := space.AddShape(cp.NewSegment(
	// 	space.StaticBody,
	// 	cp.Vector{X: -600, Y: float64(g.leveldata.targetHeight)},
	// 	cp.Vector{X: 600, Y: float64(g.leveldata.targetHeight)},
	// 	0,
	// ))
	// target.SetElasticity(1)
	// target.SetFriction(1)
	// target.SetCollisionType(impactTypeTargetSensor)

	g.space = space
	g.leveldata = &level
	g.objects = []*Object{}
	g.objectIndex = 0

	for _, brick := range g.leveldata.bricks {
		createBrick(space, brick)
	}
	g.canFire = true

	handler := space.NewCollisionHandler(impactTypeBrick, impactTypeBrick)
	handler.BeginFunc = func(arb *cp.Arbiter, space *cp.Space, userData interface{}) bool {
		// This function is called when two shapes first start colliding
		// Play a sound here
		// fmt.Println("Collision detected!")
		// Your code to play a sound

		b1, b2 := arb.Bodies()
		if math.Abs(b1.Velocity().Length()-b2.Velocity().Length()) > 50 {
			sePlayer := g.audio_context.NewPlayerFromBytes(brick_sounds[rand.IntN(len(brick_sounds))])
			sePlayer.SetVolume(
				math.Min(
					1,
					math.Abs(b1.Velocity().Length()-b2.Velocity().Length())/800,
				),
			)
			if g.isMainMenu {
				sePlayer.SetVolume(sePlayer.Volume() * 0.1)
			}

			sePlayer.Play()
		}

		// Return true to process the collision normally, false to ignore it
		return true
	}
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Needless Flattening")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}

func createBrick(space *cp.Space, data *BrickData) {
	w := width
	h := height
	if data.bricktype == "small" {
		w = w / 2
	}

	mass := 4.0
	moment := cp.MomentForBox(mass, w, h)

	body := space.AddBody(cp.NewBody(mass, moment))
	body.SetPosition(cp.Vector{X: float64(data.x), Y: float64(data.y)})
	body.SetAngle(data.rotation)

	data.body = body

	hw := w / 2.0
	hh := h / 2.0
	bb := &cp.BB{L: -hw, B: -hh, R: hw, T: hh}

	unit := 2.0
	verts := []cp.Vector{
		{X: bb.R, Y: bb.T - unit}, // bottom right
		{X: bb.L, Y: bb.T - unit}, // bottom left

		{X: bb.L, Y: bb.B}, // top left
		{X: bb.R, Y: bb.B}, // top right
	}

	shape := cp.NewPolyShapeRaw(body, len(verts), verts, 0)

	space.AddShape(shape)
	shape.SetElasticity(0)
	shape.SetFriction(0.5)

	shape.SetCollisionType(impactTypeBrick)
}

func fireBigBrick(space *cp.Space, g *Game) *cp.Body {
	attack_vector := getAttackVector(g)

	x_position := float64(screenWidth) / 2
	y_position := 0.0

	mass := 40.0
	height := float64(big_brick.Bounds().Dy())
	width := float64(big_brick.Bounds().Dx())

	moment := cp.MomentForBox(mass, width, height)

	body := space.AddBody(cp.NewBody(mass, moment))
	body.SetPosition(cp.Vector{X: x_position, Y: y_position})

	body.SetVelocityVector(cp.Vector{
		X: attack_vector.Normalize().X * 600,
		Y: attack_vector.Normalize().Y * 600,
	})

	hw := width / 2.0
	hh := height / 2.0
	bb := &cp.BB{L: -hw, B: -hh, R: hw, T: hh}

	unit := 2.0
	verts := []cp.Vector{
		{X: bb.R, Y: bb.T - unit}, // bottom right
		{X: bb.L, Y: bb.T - unit}, // bottom left

		{X: bb.L, Y: bb.B}, // top left
		{X: bb.R, Y: bb.B}, // top right
	}

	shape := cp.NewPolyShapeRaw(body, len(verts), verts, 0)

	space.AddShape(shape)
	shape.SetElasticity(0)
	shape.SetFriction(0.5)
	shape.SetCollisionType(impactTypeBrick)

	return body
}

func fireDynamite(space *cp.Space, g *Game) *cp.Body {
	attack_vector := getAttackVector(g)

	x_position := float64(screenWidth) / 2
	y_position := 0.0

	mass := 40.0
	height := float64(dynamite_object.Bounds().Dy())
	width := float64(dynamite_object.Bounds().Dx())

	moment := cp.MomentForBox(mass, width, height)

	body := space.AddBody(cp.NewBody(mass, moment))
	body.SetPosition(cp.Vector{X: x_position, Y: y_position})

	body.SetVelocityVector(cp.Vector{
		X: attack_vector.Normalize().X * 500,
		Y: attack_vector.Normalize().Y * 500,
	})

	hw := width / 2.0
	hh := height / 2.0
	bb := &cp.BB{L: -hw, B: -hh, R: hw, T: hh}

	unit := 2.0
	verts := []cp.Vector{
		{X: bb.R, Y: bb.T - unit}, // bottom right
		{X: bb.L, Y: bb.T - unit}, // bottom left

		{X: bb.L, Y: bb.B}, // top left
		{X: bb.R, Y: bb.B}, // top right
	}

	shape := cp.NewPolyShapeRaw(body, len(verts), verts, 0)

	space.AddShape(shape)
	shape.SetElasticity(0)
	shape.SetFriction(0.5)
	shape.SetCollisionType(impactTypeBrick)

	return body
}

func pointOnLineSegment(point, start, end cp.Vector) bool {
	d1 := distance(point, start)
	d2 := distance(point, end)
	lineLen := distance(start, end)
	buffer := 0.1

	return math.Abs((d1+d2)-lineLen) < buffer
}

func distance(p1, p2 cp.Vector) float64 {
	return math.Sqrt(math.Pow(p2.X-p1.X, 2) + math.Pow(p2.Y-p1.Y, 2))
}

func removeObject(s []*Object, i int) []*Object {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
