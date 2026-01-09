package front

import (
	"image"
	"image/color"
	"log"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/text/language"
)

func errHandle(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

var frontFaceSources = func() *text.GoTextFaceSource {
	fd, err := os.Open("assets/wenq.ttf")
	errHandle(err)
	defer fd.Close()
	face, err := text.NewGoTextFaceSource(fd)
	errHandle(err)
	return face
}()

func DrawText(screen *ebiten.Image, str string, x float64, y float64, size float64, vertical bool) {
	direction := text.DirectionLeftToRight
	if vertical {
		direction = text.DirectionTopToBottomAndLeftToRight
	}
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	text.Draw(screen, str, &text.GoTextFace{
		Source:    frontFaceSources,
		Size:      size,
		Direction: direction,
		Language:  language.Chinese,
	}, op)
}

// 文本对象
type TextItem struct {
	x, y          float64
	Width, Height float64
	Str           string
	textFace      *text.GoTextFace
	lineSpace     float64
	r, g, b, a    float32
}

func NewTextItem(str string, x float64, y float64, size float64, vertical bool, lineSpace float64, col color.Color) *TextItem {
	t := &TextItem{}
	t.Str = str
	direction := text.DirectionLeftToRight
	if vertical {
		direction = text.DirectionTopToBottomAndLeftToRight
	}
	t.textFace = &text.GoTextFace{
		Source:    frontFaceSources,
		Size:      size,
		Direction: direction,
		Language:  language.Chinese,
	}
	t.x, t.y = x, y
	t.Width, t.Height = text.Measure(str, t.textFace, lineSpace)
	t.lineSpace = lineSpace
	r, g, b, a := col.RGBA()
	convert := func(x uint32) float32 {
		return float32(x) / float32(0xffff)
	}
	t.r, t.g, t.b, t.a = convert(r), convert(g), convert(b), convert(a)
	return t
}

func (t *TextItem) SetStr(str string) {
	t.Width, t.Height = text.Measure(str, t.textFace, t.lineSpace)
	t.Str = str
}

func (t *TextItem) SetPos(x, y float64) {
	t.x, t.y = x, y
}

func (t *TextItem) Draw(screen *ebiten.Image) {
	op := &text.DrawOptions{}
	op.LineSpacing = t.lineSpace
	op.ColorScale.Scale(t.r, t.g, t.b, t.a)
	op.GeoM.Translate(t.x, t.y)
	text.Draw(screen, t.Str, t.textFace, op)
}

var (
	whiteImage = ebiten.NewImage(3, 3)

	// whiteSubImage is an internal sub image of whiteImage.
	// Use whiteSubImage at DrawTriangles instead of whiteImage in order to avoid bleeding edges.
	whiteSubImage = whiteImage.SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)
)

func init() {
	whiteImage.Fill(color.White)
}

type colSetting struct {
	pos        int
	r, g, b    float32
	r1, g1, b1 float32
}

type TextItem2 struct {
	x, y         float64
	dx, dy       []float64
	cols         []colSetting
	Width, Heigh float64
	str          string
	runes        []rune
	vertices     [][]ebiten.Vertex
	indices      [][]uint16
	strokeWide   float32
	visibility   float32 //可见度
	textFace     *text.GoTextFace
	lineSpace    float64
}

func NewTextItem2(str string, x, y, size, strokeWide, lineSpace float64) *TextItem2 {
	t := &TextItem2{x: x, y: y, strokeWide: float32(strokeWide), lineSpace: lineSpace, visibility: 1}
	t.textFace = &text.GoTextFace{
		Source:   frontFaceSources,
		Size:     size,
		Language: language.Chinese,
	}
	t.SetText(str)
	return t
}

func (t *TextItem2) SetText(str string) {
	t.str = str
	t.dx = nil
	t.dy = nil
	t.vertices = nil
	t.indices = nil
	t.cols = nil
	t.runes = nil
	if len(str) == 0 {
		t.Width, t.Heigh = 0, 0
		return
	}
	var cleanStr = str
	for _, pattent := range []string{"[bule]", "[yellow]", "[white]", "[black]", "[green]", "[grey]"} {
		cleanStr = strings.ReplaceAll(cleanStr, pattent, "")
	}
	exp, err := regexp.Compile(`\[#[a-f0-9]{6}\]`)
	errHandle(err)
	cleanStr = exp.ReplaceAllString(cleanStr, "")
	t.Width, t.Heigh = text.Measure(cleanStr, t.textFace, t.lineSpace)
	runes := []rune(str)
	for i, index, dx, dy := 0, 0, 0.0, 0.0; i < len(runes); i++ {
		if runes[i] == '[' {
			switch {
			case len(runes) >= i+9 && runes[i+1] == '#' && runes[i+8] == ']':
				var col []int64
				for j := i + 2; j < i+8; j += 2 {
					v, err := strconv.ParseInt(string(runes[j:j+2]), 16, 32)
					if err != nil {
						goto normalPhase
					}
					col = append(col, v)
				}
				t.cols = append(t.cols, colSetting{pos: index, r: float32(col[0]) / 0xff,
					g: float32(col[1]) / 0xff, b: float32(col[2]) / 0xff,
					r1: (1 - float32(col[0])/0xff), g1: (1 - float32(col[1])/0xff), b1: (1 - float32(col[2])/0xff)})
				i += 8
				continue
			case len(runes) >= i+8 && string(runes[i+1:i+8]) == "yellow]":
				i += 7
				t.cols = append(t.cols, colSetting{pos: index,
					r: 220. / 0xff, g: 220. / 0xff, b: 130. / 0xff,
					r1: 100. / 0xff, g1: 100. / 0xff, b1: 40. / 0xff})
				continue
			case len(runes) >= i+6 && string(runes[i+1:i+6]) == "blue]":
				i += 5
				t.cols = append(t.cols, colSetting{pos: index,
					r: 84. / 0xff, g: 204. / 0xff, b: 255. / 0xff,
					r1: 40. / 0xff, g1: 40. / 0xff, b1: 40. / 0xff})
				continue
			case len(runes) >= i+7 && string(runes[i+1:i+7]) == "white]":
				i += 6
				t.cols = append(t.cols, colSetting{pos: index,
					r: 245. / 0xff, g: 245. / 0xff, b: 230. / 0xff,
					r1: 0, g1: 0, b1: 0})
				continue
			case len(runes) >= i+7 && string(runes[i+1:i+7]) == "black]":
				i += 6
				t.cols = append(t.cols, colSetting{pos: index,
					r: 20. / 0xff, g: 20. / 0xff, b: 20. / 0xff,
					r1: 200. / 0xff, g1: 200. / 0xff, b1: 200. / 0xff})
				continue
			case len(runes) >= i+7 && string(runes[i+1:i+7]) == "green]":
				i += 6
				t.cols = append(t.cols, colSetting{pos: index,
					r: 210. / 0xff, g: 250. / 0xff, b: 100. / 0xff,
					r1: 40. / 0xff, g1: 40. / 0xff, b1: 40. / 0xff})
				continue
			case len(runes) >= i+6 && string(runes[i+1:i+6]) == "grey]":
				i += 5
				t.cols = append(t.cols, colSetting{pos: index,
					r: 100. / 0xff, g: 100. / 0xff, b: 100. / 0xff,
					r1: 40. / 0xff, g1: 40. / 0xff, b1: 40. / 0xff})
				continue
			}

		}
	normalPhase:
		if runes[i] == '\n' {
			dx = 0
			dy += t.lineSpace
			continue
		}
		t.dx = append(t.dx, dx)
		t.dy = append(t.dy, dy)
		width, _ := text.Measure(string(runes[i]), t.textFace, t.lineSpace)
		dx += width
		t.runes = append(t.runes, runes[i])
		index++
	}
	if len(t.cols) == 0 {
		t.cols = append(t.cols, colSetting{pos: 0,
			r: 250. / 0xff, g: 250. / 0xff, b: 200. / 0xff,
			r1: 0, g1: 0, b1: 0})
	}
	if t.strokeWide <= 0 {
		return
	}
	colPos := 0
	var r, g, b float32
	t.vertices = nil
	t.indices = nil
	for i, s := range []rune(t.runes) {
		var vertices []ebiten.Vertex
		var indices []uint16
		if len(t.cols) > colPos && t.cols[colPos].pos == i {
			r = t.cols[colPos].r1
			g = t.cols[colPos].g1
			b = t.cols[colPos].b1
			colPos++
		}
		path := &vector.Path{}
		text.AppendVectorPath(path, string(s), t.textFace, &text.LayoutOptions{})
		op := &vector.StrokeOptions{}
		op.Width = t.strokeWide
		vertices, indices = path.AppendVerticesAndIndicesForStroke(vertices, indices, op)
		for j := 0; j < len(vertices); j++ {
			vertices[j].DstX += float32(t.x + t.dx[i])
			vertices[j].DstY += float32(t.y + t.dy[i])
			vertices[j].ColorR = r
			vertices[j].ColorG = g
			vertices[j].ColorB = b
		}
		t.vertices = append(t.vertices, vertices)
		t.indices = append(t.indices, indices)
	}
}

func (t *TextItem2) SetPos(x, y float64) {
	if math.Abs(t.x-x) < 1.0 && math.Abs(t.y-y) < 1.0 {
		return
	}
	for _, vertices := range t.vertices {
		for j := 0; j < len(vertices); j++ {
			vertices[j].DstX += float32(x - t.x)
			vertices[j].DstY += float32(y - t.y)
		}
	}
	t.x, t.y = x, y
}

func (t *TextItem2) SetVisibility(v float32) {
	if math.Abs(float64(t.visibility-v)) < 0.01 {
		return
	}
	for _, vertices := range t.vertices {
		for j := 0; j < len(vertices); j++ {
			vertices[j].ColorA = v
			vertices[j].ColorA = v
		}
	}
	t.visibility = v
}

func (t *TextItem2) Draw(screen *ebiten.Image) {
	if len(t.str) == 0 {
		return
	}
	op := &ebiten.DrawTrianglesOptions{}
	// op.FillRule = ebiten.NonZero
	op.AntiAlias = true
	for i := 0; i < len(t.vertices); i++ {
		screen.DrawTriangles(t.vertices[i], t.indices[i], whiteSubImage, op)
	}
	var r, g, b float32
	a := t.visibility
	colPos := 0
	for i := 0; i < len(t.runes); i++ {
		if len(t.cols) > colPos && t.cols[colPos].pos == i {
			r = t.cols[colPos].r
			g = t.cols[colPos].g
			b = t.cols[colPos].b
			colPos++
		}
		op1 := &text.DrawOptions{}
		op1.LineSpacing = t.lineSpace
		op1.ColorScale.Scale(r*a, g*a, b*a, a)
		op1.GeoM.Translate(t.x+t.dx[i], t.y+t.dy[i])
		text.Draw(screen, string(t.runes[i]), t.textFace, op1)
	}
}

func (t *TextItem2) GetStr() string {
	return t.str
}
