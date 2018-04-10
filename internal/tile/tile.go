package tile

import (
	"bufio"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
)

var tileSize = 256

func nativeZoom(xCount, yCount int) int {
	return int(math.Max(math.Ceil(math.Log2(float64(xCount))), math.Ceil(math.Log2(float64(yCount)))))
}

func ChopChop(imgfile string, resume bool) error {
	f, err := os.Open(imgfile)
	if err != nil {
		return err
	}

	img, err := png.Decode(f)
	if err != nil {
		f.Close()
		return err
	}
	f.Close()

	bnd := img.Bounds()

	width := bnd.Dx()
	height := bnd.Dy()

	tileXCount := int(math.Ceil(float64(width) / float64(tileSize)))
	tileYCount := int(math.Ceil(float64(height) / float64(tileSize)))
	xPaddingRequired := tileXCount*tileSize - width
	yPaddingRequired := tileYCount*tileSize - height

	if xPaddingRequired > 0 {
		width += xPaddingRequired
	}

	if yPaddingRequired > 0 {
		height += yPaddingRequired
	}

	zCount := nativeZoom(tileXCount, tileYCount)

	layerFolder := strings.TrimSuffix(imgfile, filepath.Ext(imgfile)) + "_tiles"

	for z := 0; z <= zCount; z++ {
		zFolder := filepath.Join(layerFolder, strconv.Itoa(z))
		cover := int(math.Pow(2, float64(zCount-z))) * tileSize
		txc := int(math.Ceil(float64(width) / float64(cover)))
		tyc := int(math.Ceil(float64(height) / float64(cover)))
		fmt.Printf("allocating %v x %v\n", cover, cover)
		tile := image.NewRGBA(image.Rect(0, 0, cover, cover))
		tileBounds := tile.Bounds()

		for x := 0; x < txc; x++ {
			xFolder := filepath.Join(zFolder, strconv.Itoa(x))
			err := os.MkdirAll(xFolder, os.ModePerm)
			if err != nil {
				return err
			}

			for y := 0; y < tyc; y++ {
				filename := filepath.Join(xFolder, fmt.Sprintf("%v.png", y))

				if _, err := os.Stat(filename); resume && !os.IsNotExist(err) {
					continue
				}

				draw.Draw(tile, tileBounds, image.Transparent, image.ZP, draw.Src)
				clipRect := image.Rect(x*cover, y*cover, x*cover+cover, y*cover+cover)
				draw.Draw(tile, tileBounds, img, clipRect.Min, draw.Src)

				outFile, err := os.Create(filename)
				if err != nil {
					return err
				}

				b := bufio.NewWriter(outFile)
				if tileSize == cover {
					err = png.Encode(b, tile)
					if err != nil {
						outFile.Close()
						return err
					}
				} else {
					resizedTile := imaging.Resize(tile, tileSize, tileSize, imaging.Lanczos)
					err = png.Encode(b, resizedTile)
					if err != nil {
						outFile.Close()
						return err
					}
				}
				err = b.Flush()
				if err != nil {
					return err
				}
				outFile.Close()
			}
		}
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
