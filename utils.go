package main

import (
	"bytes"
	"image"
	"image/color"
	_ "image/png"
	"io"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"embed"
)

var (
	whiteImage = ebiten.NewImage(3, 3)

	// whiteSubImage is an internal sub image of whiteImage.
	// Use whiteSubImage at DrawTriangles instead of whiteImage in order to avoid bleeding edges.
	whiteSubImage = whiteImage.SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)
	sampleRate    = 48000
)

func init() {
	whiteImage.Fill(color.White)
}

//go:embed assets/**
var emb embed.FS

func LoadImageFromPath(path string) *ebiten.Image {
	file, err := emb.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		panic(err)
	}

	sheet := ebiten.NewImageFromImage(img)

	return sheet
}

func StrokePath(screen *ebiten.Image, path *vector.Path, colour color.RGBA, width float32, x float32, y float32) {
	op_s := &vector.StrokeOptions{}
	op_s.Width = width
	op_s.LineJoin = vector.LineJoinRound
	vs, is := path.AppendVerticesAndIndicesForStroke(nil, nil, op_s)

	for i := range vs {
		vs[i].DstX = (vs[i].DstX + x)
		vs[i].DstY = (vs[i].DstY + y)
		vs[i].SrcX = 1
		vs[i].SrcY = 1
		vs[i].ColorR = float32(colour.R) / float32(0xff)
		vs[i].ColorG = float32(colour.G) / float32(0xff)
		vs[i].ColorB = float32(colour.B) / float32(0xff)
		vs[i].ColorA = float32(colour.A) / float32(0xff)
	}

	op := &ebiten.DrawTrianglesOptions{}
	op.AntiAlias = false

	screen.DrawTriangles(vs, is, whiteSubImage, op)
}

func ReadOggBytesFromPath(path string) []byte {
	data, err := emb.ReadFile(path)

	if err != nil {
		panic(err)
	}

	s, err := vorbis.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
	if err != nil {
		panic(err)
	}
	b, err := io.ReadAll(s)
	if err != nil {
		panic(err)
	}

	return b
}
