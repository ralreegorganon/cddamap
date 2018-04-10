package server

import (
	"fmt"

	"github.com/guregu/null"
	"github.com/jmoiron/sqlx"
)

type DB struct {
	*sqlx.DB
}

func (db *DB) Open(connectionString string) error {
	d, err := sqlx.Connect("postgres", connectionString)
	if err != nil {
		return err
	}
	db.DB = d
	return nil
}

func (db *DB) GetWorlds() ([]World, error) {
	worlds := []World{}
	err := db.Select(&worlds, `
		select 
			world_id, 
			name
		from 
			world
	`)
	if err != nil {
		return nil, err
	}
	return worlds, nil
}

func (db *DB) GetWorldInfo(worldID int) (WorldInfo, error) {
	worldInfo := WorldInfo{
		Z: make(map[int]*ZLevel),
	}

	worldLayerInfos := []WorldLayerInfo{}
	err := db.Select(&worldLayerInfos, `
		select
			w.world_id,
			w.maxz,
			l.layer_id,
			l.z,
			l.type,
			c.name character_name,
			w.name world_name
		from 
			world w
			left outer join layer l 
				on w.world_id = l.world_id
			left outer join character c
				on l.character_id = c.character_id
		where
			w.world_id = $1
	`, worldID)
	if err != nil {
		return worldInfo, err
	}

	if len(worldLayerInfos) == 0 {
		return worldInfo, err
	}

	worldInfo.ID = worldLayerInfos[0].WorldID
	worldInfo.Name = worldLayerInfos[0].WorldName
	worldInfo.MaxZ = worldLayerInfos[0].MaxZ

	for _, wli := range worldLayerInfos {
		z, ok := worldInfo.Z[wli.Z]
		if !ok {
			z = &ZLevel{
				SeenLayer:      make(map[string]int),
				SeenSolidLayer: make(map[string]int),
			}
			worldInfo.Z[wli.Z] = z
		}

		switch wli.Type {
		case "overmap":
			z.TerrainLayer = null.IntFrom(int64(wli.LayerID))
			break
		case "seen":
			z.SeenLayer[wli.CharacterName.String] = wli.LayerID
			break
		case "seen_solid":
			z.SeenSolidLayer[wli.CharacterName.String] = wli.LayerID
			break
		}
	}

	return worldInfo, nil
}

func (db *DB) GetCellJson(layerID int, x, y float64) ([]byte, error) {
	sql := fmt.Sprintf(`
		select
			row_to_json(fc) geojson
		from
			(
				select
					'FeatureCollection' as type,
					array_to_json(array_agg(f)) as features
				from
				(
					select
						'Feature' as type,
						st_asgeojson(the_geom)::json as geometry,
						json_build_object(
							'id', id, 
							'name', name
						) as properties
					from
						v_cell
					where 
						layer_id = $1
						and ST_CoveredBy(ST_GeomFromText('POINT(%[1]f %[2]f)'), the_geom)
				) as f
			) as fc
		`, x, y)

	var json []byte
	err := db.QueryRow(sql, layerID).Scan(&json)
	if err != nil {
		return nil, err
	}
	return json, nil
}

func (db *DB) GetTileRoot(layerID int) (string, error) {
	var tileRoot string
	err := db.QueryRow("select tile_root from v_tile where layer_id = $1", layerID).Scan(&tileRoot)
	if err != nil {
		return "", err
	}
	return tileRoot, nil
}
