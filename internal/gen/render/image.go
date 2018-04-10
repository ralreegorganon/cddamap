package render

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/ralreegorganon/cddamap/internal/gen/world"
	"golang.org/x/image/font"
)

var dpi = 72.0
var size = 24.0
var spacing = 1.0
var cellWidth = 21.3594
var cellHeight = 24
var cellOverprintWidth = 24
var mapFont *truetype.Font
var colorCache map[color.RGBA]*image.Uniform

func init() {
	fontBytes, err := Asset("Topaz-8.ttf")
	if err != nil {
		panic(err)
	}

	mapFont, err = freetype.ParseFont(fontBytes)
	if err != nil {
		panic(err)
	}

	colorCache = make(map[color.RGBA]*image.Uniform)
}

func Image(w world.World, outputRoot, overmapFilter string, includeLayers []int, terrain, seen, seenSolid, skipEmpty, cities bool) error {
	err := os.MkdirAll(outputRoot, os.ModePerm)
	if err != nil {
		return err
	}

	e := &png.Encoder{
		BufferPool: &pool{},
	}

	if len(includeLayers) == 0 {
		return nil
	}

	l := w.TerrainLayers[includeLayers[0]]

	width := int(cellOverprintWidth * len(l.TerrainRows[0].TerrainCellKeys))
	height := cellHeight * len(l.TerrainRows)

	fullImage := image.NewRGBA(image.Rect(0, 0, width, height))

	c := freetype.NewContext()
	c.SetDPI(dpi)
	c.SetFont(mapFont)
	c.SetFontSize(size)
	c.SetClip(fullImage.Bounds())
	c.SetDst(fullImage)
	c.SetHinting(font.HintingNone)

	for _, layerID := range includeLayers {
		if terrain {
			err := terrainToImage(e, fullImage, c, w, outputRoot, overmapFilter, layerID, skipEmpty)
			if err != nil {
				return err
			}
		}

		if seen {
			err := seenToImage(e, fullImage, c, w, outputRoot, overmapFilter, layerID, skipEmpty)
			if err != nil {
				return err
			}
		}

		if seenSolid {
			err := seenToImageSolid(e, fullImage, c, w, outputRoot, overmapFilter, layerID, skipEmpty)
			if err != nil {
				return err
			}
		}
	}

	if cities {
		err := citiesToImage(e, fullImage, c, w, outputRoot, overmapFilter)
		if err != nil {
			return err
		}
	}
	return nil
}

func terrainToImage(e *png.Encoder, fullImage *image.RGBA, c *freetype.Context, w world.World, outputRoot, overmapFilter string, layerID int, skipEmpty bool) error {
	l := w.TerrainLayers[layerID]

	if l.Empty && skipEmpty {
		return nil
	}

	draw.Draw(fullImage, fullImage.Bounds(), image.Black, image.ZP, draw.Src)

	pt := freetype.Pt(0, 0+int(c.PointToFixed(size)>>6))
	for _, r := range l.TerrainRows {
		for _, k := range r.TerrainCellKeys {
			cell := w.TerrainCellLookup[k]
			bg, ok := colorCache[cell.ColorBG]
			if !ok {
				bg = image.NewUniform(cell.ColorBG)
				colorCache[cell.ColorBG] = bg
			}

			fg, ok := colorCache[cell.ColorFG]
			if !ok {
				fg = image.NewUniform(cell.ColorFG)
				colorCache[cell.ColorFG] = fg
			}

			draw.Draw(fullImage, image.Rect(int(pt.X>>6), int(pt.Y>>6), int(pt.X>>6)+cellOverprintWidth, int(pt.Y>>6)-cellHeight), bg, image.ZP, draw.Src)
			c.SetSrc(fg)
			c.DrawString(cell.Symbol, pt)
			pt.X += c.PointToFixed(float64(cellOverprintWidth))
		}
		pt.X = c.PointToFixed(0)
		pt.Y += c.PointToFixed(size * spacing)
	}

	filename := filepath.Join(outputRoot, fmt.Sprintf("o%v_%v.png", overmapFilter, layerID))
	err := write(filename, e, fullImage)
	if err != nil {
		return err
	}

	return nil
}

func write(filename string, e *png.Encoder, fullImage *image.RGBA) error {
	outFile, err := os.Create(filename)
	if err != nil {
		return err
	}

	b := bufio.NewWriter(outFile)
	err = e.Encode(b, fullImage)
	if err != nil {
		outFile.Close()
		return err
	}

	err = b.Flush()
	if err != nil {
		outFile.Close()
		return err
	}

	outFile.Close()
	return nil
}

func seenToImage(e *png.Encoder, fullImage *image.RGBA, c *freetype.Context, w world.World, outputRoot, overmapFilter string, layerID int, skipEmpty bool) error {
	for name, layers := range w.SeenLayers {
		l := layers[layerID]

		if l.Empty && skipEmpty {
			continue
		}

		draw.Draw(fullImage, fullImage.Bounds(), image.Black, image.ZP, draw.Src)

		pt := freetype.Pt(0, 0+int(c.PointToFixed(size)>>6))
		for _, r := range l.SeenRows {
			for _, k := range r.SeenCellKeys {
				cell := w.SeenCellLookup[k]
				bg, ok := colorCache[cell.ColorBG]
				if !ok {
					bg = image.NewUniform(cell.ColorBG)
					colorCache[cell.ColorBG] = bg
				}

				fg, ok := colorCache[cell.ColorFG]
				if !ok {
					fg = image.NewUniform(cell.ColorFG)
					colorCache[cell.ColorFG] = fg
				}

				draw.Draw(fullImage, image.Rect(int(pt.X>>6), int(pt.Y>>6), int(pt.X>>6)+cellOverprintWidth, int(pt.Y>>6)-cellHeight), bg, image.ZP, draw.Src)
				c.SetSrc(fg)
				c.DrawString(cell.Symbol, pt)
				pt.X += c.PointToFixed(float64(cellOverprintWidth))
			}
			pt.X = c.PointToFixed(0)
			pt.Y += c.PointToFixed(size * spacing)
		}

		filename := filepath.Join(outputRoot, fmt.Sprintf("%v%v_visible_%v.png", name, overmapFilter, layerID))
		err := write(filename, e, fullImage)
		if err != nil {
			return err
		}
	}

	return nil
}

func seenToImageSolid(e *png.Encoder, fullImage *image.RGBA, c *freetype.Context, w world.World, outputRoot, overmapFilter string, layerID int, skipEmpty bool) error {
	for name, layers := range w.SeenLayers {
		l := layers[layerID]

		if l.Empty && skipEmpty {
			continue
		}

		draw.Draw(fullImage, fullImage.Bounds(), image.Black, image.ZP, draw.Src)

		pt := freetype.Pt(0, 0+int(c.PointToFixed(size)>>6))
		for _, r := range l.SeenRows {
			for _, k := range r.SeenCellKeys {
				cell := w.SeenCellLookup[k]
				bg, ok := colorCache[cell.ColorBG]
				if !ok {
					bg = image.NewUniform(cell.ColorBG)
					colorCache[cell.ColorBG] = bg
				}

				draw.Draw(fullImage, image.Rect(int(pt.X>>6), int(pt.Y>>6), int(pt.X>>6)+cellOverprintWidth, int(pt.Y>>6)-cellHeight), bg, image.ZP, draw.Src)
				pt.X += c.PointToFixed(float64(cellOverprintWidth))
			}
			pt.X = c.PointToFixed(0)
			pt.Y += c.PointToFixed(size * spacing)
		}

		filename := filepath.Join(outputRoot, fmt.Sprintf("%v%v_visible_solid_%v.png", name, overmapFilter, layerID))
		err := write(filename, e, fullImage)
		if err != nil {
			return err
		}
	}

	return nil
}

func citiesToImage(e *png.Encoder, fullImage *image.RGBA, c *freetype.Context, w world.World, outputRoot, overmapFilter string) error {
	draw.Draw(fullImage, fullImage.Bounds(), image.Transparent, image.ZP, draw.Src)

	bg := image.NewUniform(color.RGBA{255, 255, 0, 255})
	fg := image.NewUniform(color.RGBA{0, 0, 0, 255})

	pt := freetype.Pt(0, 0+int(c.PointToFixed(size)>>6))
	for _, r := range w.CityLayer.CityRows {
		for _, k := range r.CityCell {
			if k != "" {
				draw.Draw(fullImage, image.Rect(int(pt.X>>6), int(pt.Y>>6)+2, int(pt.X>>6)+cellOverprintWidth, int(pt.Y>>6)-cellHeight), bg, image.ZP, draw.Src)
				c.SetSrc(fg)
				c.DrawString(k, pt)
			}
			pt.X += c.PointToFixed(float64(cellOverprintWidth))
		}
		pt.X = c.PointToFixed(0)
		pt.Y += c.PointToFixed(size * spacing)
	}

	filename := filepath.Join(outputRoot, fmt.Sprintf("%vcities.png", overmapFilter))
	err := write(filename, e, fullImage)
	if err != nil {
		return err
	}

	return nil
}

type pool struct {
	b *png.EncoderBuffer
}

func (p *pool) Get() *png.EncoderBuffer {
	return p.b
}

func (p *pool) Put(b *png.EncoderBuffer) {
	p.b = b
}
