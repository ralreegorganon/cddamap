package metadata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/color"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/imdario/mergo"
	"github.com/ralreegorganon/cddamap/internal/gen/save"
	log "github.com/sirupsen/logrus"
)

type overmapTerrain struct {
	internalID  uint32
	ID          string   `json:"id"`
	Type        string   `json:"type"`
	Abstract    string   `json:"abstract"`
	Sym         string   `json:"sym"`
	Color       string   `json:"color"`
	LandUseCode string   `json:"land_use_code"`
	CopyFrom    string   `json:"copy-from"`
	SeeCost     int      `json:"see_cost"`
	Extras      string   `json:"extras"`
	MonDensity  int      `json:"mondensity"`
	Flags       []string `json:"flags"`
	Spawns      spawns   `json:"spawns"`
	Delete      deleteit `json:"delete"`
}

type deleteit struct {
	Flags []string `json:"flags"`
}

type spawns struct {
	Group      string `json:"group"`
	Population []int  `json:"population"`
	Chance     int    `json:"chance"`
}

type modInfo struct {
	Ident string `json:"ident"`
}

type overmapLandUseCode struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Sym   string `json:"sym"`
	Color string `json:"color"`
}

const overmapTerrainTypeID = "overmap_terrain"
const overmapLandUseCodeTypeID = "overmap_land_use_code"

type inLoadOrder []string

func (s inLoadOrder) Len() int {
	return len(s)
}

func (s inLoadOrder) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s inLoadOrder) Less(i, j int) bool {
	c1 := strings.Count(s[i], "/")
	c2 := strings.Count(s[j], "/")

	if c1 == c2 {
		return s[i] < s[j]
	}
	return c1 < c2
}

func indexOf(slice []string, item string) int {
	for i := range slice {
		if slice[i] == item {
			return i
		}
	}
	return -1
}

var linearSuffixes = []string{
	"_isolated",
	"_end_south",
	"_end_west",
	"_ne",
	"_end_north",
	"_ns",
	"_es",
	"_nes",
	"_end_east",
	"_wn",
	"_ew",
	"_new",
	"_sw",
	"_nsw",
	"_esw",
	"_nesw"}

var linearSuffixSymbols = map[string]string{
	"_isolated":  "",
	"_end_south": "\u2502",
	"_end_west":  "\u2500",
	"_ne":        "\u2514",
	"_end_north": "\u2502",
	"_ns":        "\u2502",
	"_es":        "\u250c",
	"_nes":       "\u251c",
	"_end_east":  "\u2500",
	"_wn":        "\u2518",
	"_ew":        "\u2500",
	"_new":       "\u2534",
	"_sw":        "\u2510",
	"_nsw":       "\u2524",
	"_esw":       "\u252c",
	"_nesw":      "\u253c",
}

var rotationSuffixes = []string{
	"_north",
	"_east",
	"_south",
	"_west"}

var rotations [][]string

type ColorPair struct {
	FG color.RGBA
	BG color.RGBA
}

var colors map[string]ColorPair

func init() {
	rotations = make([][]string, 0)
	rotations = append(rotations, []string{"<", "^", ">", "v"})
	rotations = append(rotations, []string{"\u2518", "\u2514", "\u250c", "\u2510"})
	rotations = append(rotations, []string{"\u2500", "\u2502", "\u2500", "\u2502"})
	rotations = append(rotations, []string{"\u251c", "\u252c", "\u2524", "\u2534"})

	white := color.RGBA{150, 150, 150, 255}
	black := color.RGBA{0, 0, 0, 255}
	red := color.RGBA{255, 0, 0, 255}
	green := color.RGBA{0, 110, 0, 255}
	brown := color.RGBA{92, 51, 23, 255}
	blue := color.RGBA{0, 0, 200, 255}
	magenta := color.RGBA{139, 58, 98, 255}
	cyan := color.RGBA{0, 150, 180, 255}
	gray := color.RGBA{150, 150, 150, 255}
	darkGray := color.RGBA{99, 99, 99, 255}
	lightRed := color.RGBA{255, 150, 150, 255}
	lightGreen := color.RGBA{0, 255, 0, 255}
	yellow := color.RGBA{255, 255, 0, 255}
	lightBlue := color.RGBA{100, 100, 255, 255}
	lightMagenta := color.RGBA{254, 0, 254, 255}
	lightCyan := color.RGBA{0, 240, 255, 255}

	colors = make(map[string]ColorPair)

	colors["black_yellow"] = ColorPair{FG: black, BG: yellow}
	colors["blue_white"] = ColorPair{FG: blue, BG: white}
	colors["blue"] = ColorPair{FG: blue, BG: black}
	colors["brown_green"] = ColorPair{FG: brown, BG: green}
	colors["brown"] = ColorPair{FG: brown, BG: black}
	colors["c_blue_white"] = ColorPair{FG: blue, BG: white}
	colors["c_light_cyan_magenta"] = ColorPair{FG: lightCyan, BG: magenta}
	colors["c_red_white"] = ColorPair{FG: red, BG: white}
	colors["c_yellow_green"] = ColorPair{FG: yellow, BG: green}
	colors["c_yellow_white"] = ColorPair{FG: yellow, BG: white}
	colors["cyan"] = ColorPair{FG: cyan, BG: black}
	colors["dark_gray_magenta"] = ColorPair{FG: darkGray, BG: magenta}
	colors["dark_gray"] = ColorPair{FG: darkGray, BG: black}
	colors["green_cyan"] = ColorPair{FG: green, BG: cyan}
	colors["green_yellow"] = ColorPair{FG: green, BG: yellow}
	colors["green"] = ColorPair{FG: green, BG: black}
	colors["h_dark_gray"] = ColorPair{FG: darkGray, BG: black}
	colors["h_yellow"] = ColorPair{FG: yellow, BG: black}
	colors["i_black"] = ColorPair{FG: black, BG: white}
	colors["i_blue"] = ColorPair{FG: black, BG: blue}
	colors["i_brown"] = ColorPair{FG: black, BG: brown}
	colors["i_cyan"] = ColorPair{FG: black, BG: cyan}
	colors["i_green"] = ColorPair{FG: black, BG: green}
	colors["i_light_blue"] = ColorPair{FG: black, BG: lightBlue}
	colors["i_light_cyan"] = ColorPair{FG: black, BG: lightCyan}
	colors["i_light_gray"] = ColorPair{FG: black, BG: gray}
	colors["i_light_green"] = ColorPair{FG: black, BG: lightGreen}
	colors["i_light_red"] = ColorPair{FG: black, BG: lightRed}
	colors["i_magenta"] = ColorPair{FG: black, BG: magenta}
	colors["i_pink"] = ColorPair{FG: black, BG: lightMagenta}
	colors["i_red"] = ColorPair{FG: black, BG: red}
	colors["i_white"] = ColorPair{FG: white, BG: black}
	colors["i_yellow"] = ColorPair{FG: black, BG: yellow}
	colors["light_blue"] = ColorPair{FG: lightBlue, BG: black}
	colors["light_cyan_magenta"] = ColorPair{FG: lightCyan, BG: magenta}
	colors["light_cyan"] = ColorPair{FG: lightCyan, BG: black}
	colors["light_gray"] = ColorPair{FG: gray, BG: black}
	colors["light_green_cyan"] = ColorPair{FG: lightGreen, BG: cyan}
	colors["light_green_green"] = ColorPair{FG: lightGreen, BG: green}
	colors["light_green_red"] = ColorPair{FG: lightGreen, BG: red}
	colors["light_green_yellow"] = ColorPair{FG: lightGreen, BG: yellow}
	colors["light_green"] = ColorPair{FG: lightGreen, BG: black}
	colors["light_red"] = ColorPair{FG: lightRed, BG: black}
	colors["magenta"] = ColorPair{FG: magenta, BG: black}
	colors["pink_magenta"] = ColorPair{FG: lightMagenta, BG: magenta}
	colors["pink"] = ColorPair{FG: lightMagenta, BG: black}
	colors["red_white"] = ColorPair{FG: red, BG: white}
	colors["red"] = ColorPair{FG: red, BG: black}
	colors["unset"] = ColorPair{FG: white, BG: black}
	colors["white_magenta"] = ColorPair{FG: white, BG: magenta}
	colors["white_white"] = ColorPair{FG: white, BG: white}
	colors["white"] = ColorPair{FG: white, BG: black}
	colors["yellow_cyan"] = ColorPair{FG: yellow, BG: cyan}
	colors["yellow_green"] = ColorPair{FG: yellow, BG: green}
	colors["yellow_magenta"] = ColorPair{FG: yellow, BG: magenta}
	colors["yellow_white"] = ColorPair{FG: yellow, BG: white}
	colors["yellow"] = ColorPair{FG: yellow, BG: black}
}

type Overmap struct {
	built        map[string]overmapTerrain
	landusecodes map[string]overmapLandUseCode
}

func (o Overmap) UID(id string) uint32 {
	if t, tok := o.built[id]; tok {
		return t.internalID
	}
	return 0
}

func (o Overmap) Exists(id string) bool {
	_, ok := o.built[id]
	return ok
}

func (o Overmap) Symbol(id string, landUseCode bool) string {
	if t, tok := o.built[id]; tok {
		if !landUseCode {
			return t.Sym
		}
		if luc, lucok := o.landusecodes[t.LandUseCode]; lucok {
			return luc.Sym
		}
	}
	return "?"
}

func (o Overmap) Color(id string, landUseCode bool) (color.RGBA, color.RGBA) {
	if c, tok := o.built[id]; tok {
		if !landUseCode {
			if cp, ok := colors[c.Color]; ok {
				return cp.FG, cp.BG
			}
			fmt.Printf("missing terrain color for: %#v from %v\n", c.Color, id)
		}
		if luc, lucok := o.landusecodes[c.LandUseCode]; lucok {
			if cp, ok := colors[luc.Color]; ok {
				return cp.FG, cp.BG
			}
			fmt.Printf("missing landusecode color for: %#v\n", luc.Color)
		}
	}

	unset := colors["unset"]
	return unset.FG, unset.BG
}

func Build(save save.Save, gameRoot string) (Overmap, error) {
	o := Overmap{}

	jsonRoot := filepath.Join(gameRoot, "data", "json")
	modsRoot := filepath.Join(gameRoot, "data", "mods")
	files, err := overmapTerrainSourceFiles(jsonRoot, modsRoot, save.Mods)
	if err != nil {
		return o, err
	}

	templates := make(map[string]overmapTerrain)
	landusecodes := make(map[string]overmapLandUseCode)
	for _, f := range files {
		err = loadTemplates(f, templates, landusecodes)
		if err != nil {
			return o, err
		}
	}

	built, err := buildTemplates(templates)
	if err != nil {
		return o, err
	}

	o = Overmap{
		built:        built,
		landusecodes: landusecodes,
	}

	return o, nil
}

func overmapTerrainSourceFiles(jsonRoot, modsRoot string, saveMods []string) ([]string, error) {
	files := []string{}

	err := filepath.Walk(jsonRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".json") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	activeMods := map[string]string{}
	for _, m := range saveMods {
		activeMods[m] = m
	}

	mods, err := ioutil.ReadDir(modsRoot)
	if err != nil {
		return nil, err
	}

	for _, f := range mods {
		if !f.IsDir() {
			continue
		}

		modInfoPath := filepath.Join(modsRoot, f.Name(), "modinfo.json")
		b, err := ioutil.ReadFile(modInfoPath)
		if err != nil {
			return nil, err
		}
		var modinfo []modInfo
		err = json.Unmarshal(b, &modinfo)
		if err != nil {
			return nil, err
		}

		if _, ok := activeMods[modinfo[0].Ident]; !ok {
			continue
		}

		err = filepath.Walk(filepath.Join(modsRoot, f.Name()), func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if strings.HasSuffix(path, ".json") {
				files = append(files, path)
			}
			return nil
		})
	}

	sort.Sort(inLoadOrder(files))

	return files, nil
}

func loadTemplates(file string, templates map[string]overmapTerrain, landusecodes map[string]overmapLandUseCode) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	hasContent := bytes.Contains(b, []byte(overmapTerrainTypeID)) || bytes.Contains(b, []byte(overmapLandUseCodeTypeID))
	if !hasContent {
		return nil
	}

	var temp []map[string]interface{}
	err = json.Unmarshal(b, &temp)
	if err != nil {
		return err
	}

	filteredObjects := make([]map[string]interface{}, 0)
	for _, t := range temp {
		if t["type"].(string) == overmapTerrainTypeID {
			filteredObjects = append(filteredObjects, t)
		}
	}

	expandedOvermapTerrains := make([]map[string]interface{}, 0)
	for _, t := range filteredObjects {
		switch x := t["id"].(type) {
		case string:
			expandedOvermapTerrains = append(expandedOvermapTerrains, t)
		case []interface{}:
			for _, id := range x {
				b := make(map[string]interface{})
				byt, err := json.Marshal(t)
				if err != nil {
					return err
				}
				json.Unmarshal(byt, &b)
				b["id"] = id
				expandedOvermapTerrains = append(expandedOvermapTerrains, b)
			}
		}
		switch x := t["abstract"].(type) {
		case string:
			expandedOvermapTerrains = append(expandedOvermapTerrains, t)
		case []interface{}:
			for _, id := range x {
				b := make(map[string]interface{})
				byt, err := json.Marshal(t)
				if err != nil {
					return err
				}
				json.Unmarshal(byt, &b)
				b["abstract"] = id
				expandedOvermapTerrains = append(expandedOvermapTerrains, b)
			}
		}
	}

	filteredText, err := json.Marshal(expandedOvermapTerrains)
	if err != nil {
		return err
	}

	var overmapTerrains []overmapTerrain
	err = json.Unmarshal(filteredText, &overmapTerrains)
	if err != nil {
		return err
	}

	for _, ot := range overmapTerrains {
		if ot.Type != overmapTerrainTypeID {
			continue
		}
		if ot.Abstract != "" {
			templates[ot.Abstract] = ot
		} else {
			templates[ot.ID] = ot
		}
	}

	filteredObjects = make([]map[string]interface{}, 0)
	for _, t := range temp {
		if t["type"].(string) == overmapLandUseCodeTypeID {
			filteredObjects = append(filteredObjects, t)
		}
	}

	filteredText, err = json.Marshal(filteredObjects)
	if err != nil {
		return err
	}

	var overmapLandUseCodes []overmapLandUseCode
	err = json.Unmarshal(filteredText, &overmapLandUseCodes)
	if err != nil {
		return err
	}

	for _, oluc := range overmapLandUseCodes {
		if oluc.Type != overmapLandUseCodeTypeID {
			continue
		}
		landusecodes[oluc.ID] = oluc
	}

	return nil
}

func buildTemplates(templates map[string]overmapTerrain) (map[string]overmapTerrain, error) {
	built := make(map[string]overmapTerrain)

	for _, ot := range templates {
		bt := make([]overmapTerrain, 0)
		t := ot
		bt = append(bt, t)
		for t.CopyFrom != "" {
			t = templates[t.CopyFrom]
			t.internalID = save.HashTerrainID(t.ID)
			bt = append(bt, t)
		}

		b := overmapTerrain{}
		for i := len(bt) - 1; i >= 0; i-- {
			if err := mergo.Merge(&b, bt[i], mergo.WithOverride); err != nil {
				return built, err
			}
		}

		if ot.Abstract == "" {
			b.Abstract = ""
			b.CopyFrom = ""
			b.internalID = save.HashTerrainID(b.ID)
			built[b.ID] = b

			rotate := true

			if b.Flags != nil {
				flagsMap := make(map[string]bool)
				for _, f := range b.Flags {
					flagsMap[f] = true
				}

				if b.Delete.Flags != nil {
					for _, f := range b.Delete.Flags {
						delete(flagsMap, f)
					}
				}

				if _, ok := flagsMap["NO_ROTATE"]; ok {
					rotate = false
				}

				if _, ok := flagsMap["LINEAR"]; ok {
					for _, suffix := range linearSuffixes {
						bs := overmapTerrain{}
						if err := mergo.Merge(&bs, b, mergo.WithOverride); err != nil {
							return built, err
						}
						bs.ID = b.ID + suffix
						bs.Sym = linearSuffixSymbols[suffix]
						b.internalID = save.HashTerrainID(b.ID)
						built[bs.ID] = bs
					}
				}
			}

			if rotate {
				for i, suffix := range rotationSuffixes {
					bs := overmapTerrain{}
					if err := mergo.Merge(&bs, b, mergo.WithOverride); err != nil {
						log.Fatal(err)
					}
					bs.ID = b.ID + suffix
					b.internalID = save.HashTerrainID(b.ID)

					for _, r := range rotations {
						index := indexOf(r, b.Sym)
						if index != -1 {
							bs.Sym = r[(i+index+4)%4]
							break
						}
					}
					built[bs.ID] = bs
				}
			}
		}
	}

	return built, nil
}
