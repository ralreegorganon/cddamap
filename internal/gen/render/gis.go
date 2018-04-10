package render

import (
	"database/sql"
	"fmt"
	"math"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/ralreegorganon/cddamap/internal/gen/save"
	"github.com/ralreegorganon/cddamap/internal/gen/world"
)

func GIS(w world.World, connectionString string, includeLayers []int, terrain, seen, seenSolid, skipEmpty, cities bool) error {
	tl := w.TerrainLayers[includeLayers[0]]
	width := int(cellWidth * float64(len(tl.TerrainRows[0].TerrainCellKeys)))
	height := cellHeight * len(tl.TerrainRows)

	tileXCount := int(math.Ceil(float64(width) / float64(tileSize)))
	tileYCount := int(math.Ceil(float64(height) / float64(tileSize)))

	maxz := nativeZoom(tileXCount, tileYCount)

	db, err := sqlx.Open("postgres", connectionString)
	if err != nil {
		return err
	}

	var worldID int
	err = db.QueryRow("insert into world (name, maxz) values ($1, $2) on conflict(name) do update set maxz = EXCLUDED.maxz returning world_id", w.Name, maxz).Scan(&worldID)
	if err != nil {
		return err
	}

	emptyRockHash := save.HashTerrainID("empty_rock")
	openAirHash := save.HashTerrainID("open_air")
	blankHash := save.HashTerrainID("")

	for _, i := range includeLayers {
		if seen || seenSolid {
			for name, layers := range w.SeenLayers {
				l := layers[i]

				if l.Empty && skipEmpty {
					continue
				}

				var characterID int
				err = db.QueryRow("select character_id from character where world_id = $1 and namehash = $2", worldID, name).Scan(&characterID)
				if err == sql.ErrNoRows {
					err = db.QueryRow("insert into character (world_id, namehash, name) values ($1, $2, $2) returning character_id", worldID, name).Scan(&characterID)
					if err != nil {
						return err
					}
				} else if err != nil {
					return err
				}

				if seen {
					var layerID int
					err = db.QueryRow("select layer_id from layer where world_id = $1 and z = $2 and character_id = $3 and type = 'seen'", worldID, i, characterID).Scan(&layerID)
					if err == sql.ErrNoRows {
						err = db.QueryRow("insert into layer (world_id, z, character_id, type) values ($1, $2, $3, 'seen') returning layer_id", worldID, i, characterID).Scan(&layerID)
						if err != nil {
							return err
						}
					} else if err != nil {
						return err
					}
				}
				if seenSolid {
					var layerID int
					err = db.QueryRow("select layer_id from layer where world_id = $1 and z = $2 and character_id = $3 and type = 'seen_solid'", worldID, i, characterID).Scan(&layerID)
					if err == sql.ErrNoRows {
						err = db.QueryRow("insert into layer (world_id, z, character_id, type) values ($1, $2, $3, 'seen_solid') returning layer_id", worldID, i, characterID).Scan(&layerID)
						if err != nil {
							return err
						}
					} else if err != nil {
						return err
					}
				}

			}
		}

		if terrain {
			l := w.TerrainLayers[i]

			if l.Empty && skipEmpty {
				continue
			}

			var layerID int
			err = db.QueryRow("select layer_id from layer where world_id = $1 and z = $2 and type = 'overmap'", worldID, i).Scan(&layerID)
			if err == sql.ErrNoRows {
				err = db.QueryRow("insert into layer (world_id, z, type) values ($1, $2, 'overmap') returning layer_id", worldID, i).Scan(&layerID)
				if err != nil {
					return err
				}
			} else if err != nil {
				return err
			}

			txn, err := db.Begin()
			if err != nil {
				return err
			}

			nukeCellsStmt, err := txn.Prepare("delete from cell where layer_id = $1")
			if err != nil {
				return err
			}

			_, err = nukeCellsStmt.Exec(layerID)
			if err != nil {
				return err
			}

			stmt, err := txn.Prepare(pq.CopyIn("cell", "layer_id", "id", "name", "the_geom"))
			if err != nil {
				return err
			}

			for ri, r := range l.TerrainRows {
				for ci, k := range r.TerrainCellKeys {
					if k == emptyRockHash || k == openAirHash || k == blankHash {
						continue
					}

					x := float64(ci) * cellWidth
					y := float64(ri) * float64(cellHeight)
					x2 := x + cellWidth
					y2 := y + float64(cellHeight)

					c := w.TerrainCellLookup[k]

					geom := fmt.Sprintf("POLYGON((%[1]f %[2]f, %[3]f %[4]f, %[5]f %[6]f, %[7]f %[8]f, %[1]f %[2]f))", x, y, x2, y, x2, y2, x, y2)
					_, err = stmt.Exec(layerID, c.ID, c.Name, geom)
					if err != nil {
						return err
					}
				}
			}
			_, err = stmt.Exec()
			if err != nil {
				return err
			}

			err = stmt.Close()
			if err != nil {
				return err
			}

			err = txn.Commit()
			if err != nil {
				return err
			}
		}
	}

	if cities {
		var layerID int
		err = db.QueryRow("select layer_id from layer where world_id = $1 and z = $2 and type = 'city'", worldID, 10).Scan(&layerID)
		if err == sql.ErrNoRows {
			err = db.QueryRow("insert into layer (world_id, z, type) values ($1, $2, 'city') returning layer_id", worldID, 10).Scan(&layerID)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		txn, err := db.Begin()
		if err != nil {
			return err
		}

		nukeCitiesStmt, err := txn.Prepare("delete from city where world_id = $1")
		if err != nil {
			return err
		}

		_, err = nukeCitiesStmt.Exec(worldID)
		if err != nil {
			return err
		}

		stmt, err := txn.Prepare(pq.CopyIn("city", "world_id", "name", "size", "the_geom"))
		if err != nil {
			return err
		}

		for _, c := range w.CityLayer.Cities {
			x := float64(c.X)*cellWidth + cellWidth/2
			y := float64(c.Y)*float64(cellHeight) + float64(cellWidth)/2

			geom := fmt.Sprintf("POINT(%[1]f %[2]f)", x, y)
			_, err = stmt.Exec(worldID, c.Name, c.Size, geom)
			if err != nil {
				return err
			}
		}
		_, err = stmt.Exec()
		if err != nil {
			return err
		}

		err = stmt.Close()
		if err != nil {
			return err
		}

		err = txn.Commit()
		if err != nil {
			return err
		}
	}

	return nil
}

func nativeZoom(xCount, yCount int) int {
	return int(math.Max(math.Ceil(math.Log2(float64(xCount))), math.Ceil(math.Log2(float64(yCount)))))
}

var tileSize = 256
