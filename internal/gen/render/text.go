package render

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ralreegorganon/cddamap/internal/gen/world"
)

func Text(w world.World, outputRoot, overmapFilter string, includeLayers []int, terrain, seen, skipEmpty, cities bool) error {
	err := os.MkdirAll(outputRoot, os.ModePerm)
	if err != nil {
		return err
	}

	for _, layerID := range includeLayers {
		if terrain {
			err := terrainToText(w, outputRoot, overmapFilter, layerID, skipEmpty)
			if err != nil {
				return err
			}
		}
		if seen {
			err = seenToText(w, outputRoot, overmapFilter, layerID, skipEmpty)
			if err != nil {
				return err
			}
		}
	}

	if cities {
		err = cityToText(w, outputRoot, overmapFilter)
		if err != nil {
			return err
		}
	}
	return nil
}

func terrainToText(w world.World, outputRoot, overmapFilter string, layerID int, skipEmpty bool) error {
	l := w.TerrainLayers[layerID]

	if l.Empty && skipEmpty {
		return nil
	}

	var b strings.Builder
	for _, r := range l.TerrainRows {
		for _, k := range r.TerrainCellKeys {
			c := w.TerrainCellLookup[k]
			b.WriteString(c.Symbol)
		}
		b.WriteString("\n")
	}

	filename := filepath.Join(outputRoot, fmt.Sprintf("o%v_%v", overmapFilter, layerID))
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	f.WriteString(b.String())
	return nil
}

func seenToText(w world.World, outputRoot, overmapFilter string, layerID int, skipEmpty bool) error {
	for name, layers := range w.SeenLayers {
		l := layers[layerID]

		if l.Empty && skipEmpty {
			continue
		}

		var b strings.Builder
		for _, r := range l.SeenRows {
			for _, k := range r.SeenCellKeys {
				cell := w.SeenCellLookup[k]
				b.WriteString(cell.Symbol)

			}
			b.WriteString("\n")
		}

		filename := filepath.Join(outputRoot, fmt.Sprintf("%v%v_visible_%v", name, overmapFilter, layerID))
		f, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer f.Close()

		f.WriteString(b.String())
	}
	return nil
}

func cityToText(w world.World, outputRoot, overmapFilter string) error {
	var b strings.Builder
	for _, r := range w.CityLayer.CityRows {
		for _, k := range r.CityCell {
			if k == "" {
				b.WriteString(" ")
			} else {
				b.WriteString(k)
			}
		}
		b.WriteString("\n")
	}

	filename := filepath.Join(outputRoot, fmt.Sprintf("%vcities", overmapFilter))
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	f.WriteString(b.String())
	return nil
}
