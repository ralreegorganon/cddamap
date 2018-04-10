package server

import (
	"github.com/guregu/null"
)

type World struct {
	ID   int    `json:"id" db:"world_id"`
	Name string `json:"name" db:"name"`
}

type WorldLayerInfo struct {
	WorldID       int         `json:"worldId" db:"world_id"`
	LayerID       int         `json:"layerId" db:"layer_id"`
	MaxZ          int         `json:"maxz" db:"maxz"`
	Z             int         `json:"z" db:"z"`
	Type          string      `json:"type" db:"type"`
	WorldName     string      `json:"worldName" db:"world_name"`
	CharacterName null.String `json:"characterName" db:"character_name"`
}

type ZLevel struct {
	TerrainLayer   null.Int       `json:"layerId"`
	SeenLayer      map[string]int `json:"seenLayers"`
	SeenSolidLayer map[string]int `json:"seenSolidLayers"`
}

type WorldInfo struct {
	ID   int             `json:"id"`
	Name string          `json:"name"`
	MaxZ int             `json:"maxz"`
	Z    map[int]*ZLevel `json:"z"`
}
