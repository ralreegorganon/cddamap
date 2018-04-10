package world

import (
	"fmt"
	"image/color"
	"math"

	"github.com/ralreegorganon/cddamap/internal/gen/metadata"
	"github.com/ralreegorganon/cddamap/internal/gen/save"
)

func keyExists(decoded map[string]interface{}, key string) bool {
	val, ok := decoded[key]
	return ok && val != nil
}

func indexOf(slice []int, item int) int {
	for i := range slice {
		if slice[i] == item {
			return i
		}
	}
	return -1
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

type World struct {
	Name              string
	TerrainLayers     []TerrainLayer
	SeenLayers        map[string][]SeenLayer
	TerrainCellLookup map[uint32]TerrainCell
	SeenCellLookup    map[bool]SeenCell
	CityLayer         CityLayer
}

type TerrainLayer struct {
	Empty       bool
	TerrainRows []TerrainRow
}

type TerrainRow struct {
	TerrainCellKeys []uint32
}

type TerrainCell struct {
	Symbol  string
	ColorFG color.RGBA
	ColorBG color.RGBA
	Name    string
	ID      string
}

type SeenLayer struct {
	Empty    bool
	SeenRows []SeenRow
}

type SeenRow struct {
	SeenCellKeys []bool
}

type SeenCell struct {
	ID      string
	Symbol  string
	Seen    bool
	ColorFG color.RGBA
	ColorBG color.RGBA
}

type CityLayer struct {
	CityRows []CityRow
	Cities   []City
}

type CityRow struct {
	CityCell []string
}

type City struct {
	Name string
	X    int
	Y    int
	Size int
}

func Build(m metadata.Overmap, s save.Save) (World, error) {

	terrainCellLookup := make(map[uint32]TerrainCell)

	seenCellLookup := map[bool]SeenCell{
		true: SeenCell{
			Symbol:  " ",
			Seen:    true,
			ColorFG: color.RGBA{0, 0, 0, 0},
			ColorBG: color.RGBA{0, 0, 0, 0},
		},
		false: SeenCell{
			Symbol:  "#",
			Seen:    false,
			ColorFG: color.RGBA{44, 44, 44, 255},
			ColorBG: color.RGBA{0, 0, 0, 255},
		},
	}

	terrainLayers := buildTerrainLayers(m, s, terrainCellLookup)
	characterSeenLayers := buildCharacterSeenLayers(m, s)
	cityLayer := buildCityLayer(m, s)

	world := World{
		Name:              s.Name,
		TerrainLayers:     terrainLayers,
		SeenLayers:        characterSeenLayers,
		TerrainCellLookup: terrainCellLookup,
		SeenCellLookup:    seenCellLookup,
		CityLayer:         cityLayer,
	}

	return world, nil
}

type worldChunkDimensions struct {
	XSize int
	YSize int
	XMin  int
	XMax  int
	YMin  int
	YMax  int
}

func calculateWorldChunkDimensions(m metadata.Overmap, s save.Save) worldChunkDimensions {
	cXMax := math.MinInt64
	cXMin := math.MaxInt64
	cYMax := math.MinInt64
	cYMin := math.MaxInt64

	for _, c := range s.Overmap.Chunks {
		if c.X > cXMax {
			cXMax = c.X
		}
		if c.Y > cYMax {
			cYMax = c.Y
		}
		if c.X < cXMin {
			cXMin = c.X
		}
		if c.Y < cYMin {
			cYMin = c.Y
		}
	}

	cXSize := abs(cXMax-cXMin) + 1
	cYSize := abs(cYMax-cYMin) + 1

	wcd := worldChunkDimensions{
		XSize: cXSize,
		YSize: cYSize,
		XMin:  cXMin,
		XMax:  cXMax,
		YMin:  cYMin,
		YMax:  cYMax,
	}
	return wcd
}

func buildCityLayer(m metadata.Overmap, s save.Save) CityLayer {
	wcd := calculateWorldChunkDimensions(m, s)
	chunkCapacity := wcd.XSize * wcd.YSize

	cities := make([]City, 0)
	cells := make([]string, 32400*chunkCapacity)

	for _, c := range s.Overmap.Chunks {
		ci := c.X + (0 - wcd.XMin) + wcd.XSize*(c.Y+0-wcd.YMin)
		for _, city := range c.Cities {

			ce := City{
				Name: city.Name,
				Size: city.Size,
				X:    (c.X+(0-wcd.XMin))*180 + city.X,
				Y:    (c.Y+0-wcd.YMin)*180 + city.Y,
			}

			cities = append(cities, ce)

			nameStart := ci*32400 + city.Y*180 + city.X - len(city.Name)/2
			for i := 0; i < len(city.Name); i++ {
				cells[nameStart+i] = string(city.Name[i])
			}
		}
	}

	layer := CityLayer{
		Cities: cities,
	}
	layer.CityRows = make([]CityRow, 180*wcd.YSize)
	for r := 0; r < 180*wcd.YSize; r++ {
		layer.CityRows[r].CityCell = make([]string, 180*wcd.XSize)
	}

	for xi := 0; xi < wcd.XSize; xi++ {
		for yi := 0; yi < wcd.YSize; yi++ {
			for ri := 0; ri < 180; ri++ {
				for ci := 0; ci < 180; ci++ {
					cell := cells[(xi+yi*wcd.XSize)*32400+ri*180+ci]
					layer.CityRows[yi*180+ri].CityCell[xi*180+ci] = cell
				}
			}
		}
	}

	return layer
}

func buildCharacterSeenLayers(m metadata.Overmap, s save.Save) map[string][]SeenLayer {
	wcd := calculateWorldChunkDimensions(m, s)
	chunkCapacity := wcd.XSize * wcd.YSize

	seen := make(map[string][]SeenLayer)

	for name, chunks := range s.Seen {
		doneChunks := make(map[int]bool)
		cells := make([]bool, 680400*chunkCapacity)
		for _, c := range chunks.Chunks {
			ci := c.X + (0 - wcd.XMin) + wcd.XSize*(c.Y+0-wcd.YMin)
			doneChunks[ci] = true
			for li, l := range c.Visible {
				lzp := 0
				for _, e := range l {
					for i := 0; i < int(e.Count); i++ {
						tmi := ci*680400 + li*32400 + lzp
						cells[tmi] = e.Seen
						lzp++
					}
				}
			}
		}

		for i := 0; i < chunkCapacity; i++ {
			if _, ok := doneChunks[i]; !ok {
				for e := 0; e < 680400; e++ {
					cells[i*680400+e] = false
				}
			}
		}

		layers := make([]SeenLayer, 21)
		for l := 0; l < 21; l++ {
			layers[l].SeenRows = make([]SeenRow, 180*wcd.YSize)
			for r := 0; r < 180*wcd.YSize; r++ {
				layers[l].SeenRows[r].SeenCellKeys = make([]bool, 180*wcd.XSize)
			}
		}

		for li := 0; li < 21; li++ {
			empty := true
			for xi := 0; xi < wcd.XSize; xi++ {
				for yi := 0; yi < wcd.YSize; yi++ {
					for ri := 0; ri < 180; ri++ {
						for ci := 0; ci < 180; ci++ {
							cell := cells[(xi+yi*wcd.XSize)*680400+li*32400+ri*180+ci]
							layers[li].SeenRows[yi*180+ri].SeenCellKeys[xi*180+ci] = cells[(xi+yi*wcd.XSize)*680400+li*32400+ri*180+ci]
							if empty && cell != false {
								empty = false
							}
						}
					}
				}
			}
			layers[li].Empty = empty
		}
		seen[name] = layers
	}

	return seen
}

func buildTerrainLayers(m metadata.Overmap, s save.Save, tcl map[uint32]TerrainCell) []TerrainLayer {
	missingTerrain := make(map[string]int)

	for _, c := range s.Overmap.Chunks {
		for _, l := range c.Layers {
			for _, e := range l {
				if exists := m.Exists(e.OvermapTerrainID); !exists {
					if _, ok := missingTerrain[e.OvermapTerrainID]; !ok {
						missingTerrain[e.OvermapTerrainID] = 0
					}
					missingTerrain[e.OvermapTerrainID]++
				}
			}
		}
	}

	for k, v := range missingTerrain {
		fmt.Printf("missing terrain: %v x %v\n", k, v)
	}

	wcd := calculateWorldChunkDimensions(m, s)
	chunkCapacity := wcd.XSize * wcd.YSize

	doneChunks := make(map[int]bool)
	cells := make([]uint32, 680400*chunkCapacity)
	for _, c := range s.Overmap.Chunks {
		ci := c.X + (0 - wcd.XMin) + wcd.XSize*(c.Y+0-wcd.YMin)
		doneChunks[ci] = true
		for li, l := range c.Layers {
			lzp := 0
			for _, e := range l {
				h := save.HashTerrainID(e.OvermapTerrainID)
				_, ok := tcl[h]
				if !ok {
					s := m.Symbol(e.OvermapTerrainID)
					cfg, cbg := m.Color(e.OvermapTerrainID)
					n := m.Name(e.OvermapTerrainID)
					tc := TerrainCell{
						ID:      e.OvermapTerrainID,
						Symbol:  s,
						ColorFG: cfg,
						ColorBG: cbg,
						Name:    n,
					}
					tcl[h] = tc
				}

				for i := 0; i < int(e.Count); i++ {
					tmi := ci*680400 + li*32400 + lzp
					cells[tmi] = h
					lzp++
				}
			}
		}
	}

	dfg, dbg := m.Color("default")
	tc := TerrainCell{
		ID:      "",
		Symbol:  " ",
		ColorFG: dfg,
		ColorBG: dbg,
	}
	h := save.HashTerrainID(tc.ID)
	tcl[h] = tc

	for i := 0; i < chunkCapacity; i++ {
		if _, ok := doneChunks[i]; !ok {
			for e := 0; e < 680400; e++ {
				cells[i*680400+e] = h
			}
		}
	}

	layers := make([]TerrainLayer, 21)
	for l := 0; l < 21; l++ {
		layers[l].TerrainRows = make([]TerrainRow, 180*wcd.YSize)
		for r := 0; r < 180*wcd.YSize; r++ {
			layers[l].TerrainRows[r].TerrainCellKeys = make([]uint32, 180*wcd.XSize)
		}
	}

	emptyRockHash := save.HashTerrainID("empty_rock")
	openAirHash := save.HashTerrainID("open_air")
	blankHash := save.HashTerrainID("")

	for li := 0; li < 21; li++ {
		empty := true
		for xi := 0; xi < wcd.XSize; xi++ {
			for yi := 0; yi < wcd.YSize; yi++ {
				for ri := 0; ri < 180; ri++ {
					for ci := 0; ci < 180; ci++ {
						cell := cells[(xi+yi*wcd.XSize)*680400+li*32400+ri*180+ci]
						layers[li].TerrainRows[yi*180+ri].TerrainCellKeys[xi*180+ci] = cell
						if empty && cell != emptyRockHash && cell != openAirHash && cell != blankHash {
							empty = false
						}
					}
				}
			}
		}
		layers[li].Empty = empty
	}

	return layers
}
