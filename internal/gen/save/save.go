package save

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type Save struct {
	Name    string
	Mods    []string
	Overmap Overmap
	Seen    map[string]Seen
}

type Overmap struct {
	Chunks []OvermapChunk
}

type OvermapChunk struct {
	X      int
	Y      int
	Layers [][]TerrainGroup `json:"layers"`
	//RegionID string           `json:"region_id"`
	//MonsterGroups   string `json:"monster_groups"`
	Cities []City `json:"cities"`
	//RoadsOut        string `json:"roads_out"`
	//Radios          string `json:"radios"`
	//MonsterMap      string `json:"monster_map"`
	//TrackedVehicles string `json:"tracked_vehicles"`
	//ScentTraces     string `json:"scent_traces"`
	//NPCs            string `json:"npcs"`
}

type TerrainGroup struct {
	OvermapTerrainID string
	Count            float64
}

type City struct {
	Name string `json:"name"`
	X    int    `json:"x"`
	Y    int    `json:"y"`
	Size int    `json:"size"`
}

type Seen struct {
	Character string
	Chunks    []SeenChunk
}

type SeenChunk struct {
	X        int
	Y        int
	Visible  [][]SeenGroup `json:"visible"`
	Explored [][]SeenGroup `json:"explored"`
}

type SeenGroup struct {
	Seen  bool
	Count float64
}

func (tg *TerrainGroup) UnmarshalJSON(bs []byte) error {
	arr := []interface{}{}
	json.Unmarshal(bs, &arr)
	tg.OvermapTerrainID = arr[0].(string)
	tg.Count = arr[1].(float64)
	return nil
}

func (sg *SeenGroup) UnmarshalJSON(bs []byte) error {
	arr := []interface{}{}
	json.Unmarshal(bs, &arr)
	sg.Seen = arr[0].(bool)
	sg.Count = arr[1].(float64)
	return nil
}

func Build(save, filter string) (Save, error) {
	s := Save{}

	o, err := overmapFromSave(save, filter)
	if err != nil {
		return s, err
	}

	cs, err := characterSeenFromSave(save, filter)
	if err != nil {
		return s, err
	}

	saveModsPath := filepath.Join(save, "mods.json")
	b, err := ioutil.ReadFile(saveModsPath)
	if err != nil {
		return s, err
	}
	var mods []string
	err = json.Unmarshal(b, &mods)
	if err != nil {
		return s, err
	}

	name := filepath.Base(save)

	s = Save{
		Name:    name,
		Overmap: o,
		Mods:    mods,
		Seen:    cs,
	}

	return s, nil
}

func overmapFromSave(save, filter string) (Overmap, error) {
	o := Overmap{}
	chunkFiles, err := overmapChunkFiles(save, filter)
	if err != nil {
		return o, err
	}

	chunks := make([]OvermapChunk, 0)

	for _, f := range chunkFiles {
		t, err := ioutil.ReadFile(f)
		if err != nil {
			return o, err
		}

		lines := strings.Split(string(t), "\n")

		if !strings.HasPrefix(lines[0], "# version 33") {
			return o, fmt.Errorf("unsupported version: %v", lines[0])
		}

		var buffer bytes.Buffer
		for i := 1; i < len(lines); i++ {
			buffer.WriteString(lines[i])
		}

		pruned := buffer.Bytes()

		var chunk OvermapChunk
		err = json.Unmarshal(pruned, &chunk)
		if err != nil {
			return o, err
		}

		x, y, err := chunkFileNameToCoordinates(f)
		if err != nil {
			return o, err
		}
		chunk.X = x
		chunk.Y = y

		chunks = append(chunks, chunk)
	}

	o = Overmap{
		Chunks: chunks,
	}
	return o, nil
}

func overmapChunkFiles(root, filter string) ([]string, error) {
	files := []string{}
	re := regexp.MustCompile(`o\.-?\d+\.-?\d+$`)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filter != "" {
			if !strings.HasSuffix(path, filter) {
				return nil
			}
		}

		isOvermapChunk := re.MatchString(path)
		if isOvermapChunk {
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

func chunkFileNameToCoordinates(chunkFile string) (int, int, error) {
	_, file := filepath.Split(chunkFile)
	parts := strings.Split(file, ".")
	x, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, err
	}
	y, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, err
	}
	return x, y, nil
}

func characterSeenFromSave(save, filter string) (map[string]Seen, error) {
	s := make(map[string]Seen)

	chunkFiles, err := characterSeenChunkFiles(save, filter)
	if err != nil {
		return s, err
	}

	for _, f := range chunkFiles {
		t, err := ioutil.ReadFile(f)
		if err != nil {
			return s, err
		}

		lines := strings.Split(string(t), "\n")

		if !strings.HasPrefix(lines[0], "# version 33") {
			return s, fmt.Errorf("unsupported version: %v", lines[0])
		}

		var buffer bytes.Buffer
		for i := 1; i < len(lines); i++ {
			buffer.WriteString(lines[i])
		}

		pruned := buffer.Bytes()

		var chunk SeenChunk
		err = json.Unmarshal(pruned, &chunk)
		if err != nil {
			return s, err
		}

		parts := strings.Split(filepath.Base(f), ".")
		name := parts[0]

		if _, ok := s[name]; !ok {
			s[name] = Seen{
				Character: name,
				Chunks:    make([]SeenChunk, 0),
			}
		}

		x, y, err := characterSeenFileNameToCoordinates(f)
		if err != nil {
			return s, err
		}
		chunk.X = x
		chunk.Y = y

		seen := s[name]
		seen.Chunks = append(seen.Chunks, chunk)
		s[name] = seen
	}

	return s, nil
}

func characterSeenChunkFiles(root, filter string) ([]string, error) {
	files := []string{}
	re := regexp.MustCompile(`\.seen\.-?\d\.-?\d$`)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filter != "" {
			if !strings.HasSuffix(path, filter) {
				return nil
			}
		}

		isCharacterSeenChunk := re.MatchString(path)
		if isCharacterSeenChunk {
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

func characterSeenFileNameToCoordinates(chunkFile string) (int, int, error) {
	_, file := filepath.Split(chunkFile)
	parts := strings.Split(file, ".")
	x, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, err
	}
	y, err := strconv.Atoi(parts[3])
	if err != nil {
		return 0, 0, err
	}
	return x, y, nil
}

func HashTerrainID(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
