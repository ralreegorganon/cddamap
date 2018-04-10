package server

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func CreateRouter(server *HTTPServer) (*mux.Router, error) {
	r := mux.NewRouter()
	m := map[string]map[string]HttpApiFunc{
		"GET": {
			"/api/worlds":                                                                                     server.GetWorlds,
			"/api/worlds/{worldID:[0-9]+}":                                                                    server.GetWorldLayerInfo,
			"/api/worlds/{worldID:[0-9]+}/layers/{layerID:[0-9]+}/cells/{x}/{y}":                              server.GetCells,
			"/api/worlds/{worldID:[0-9]+}/layers/{layerID:[0-9]+}/tiles/{z:[0-9]+}/{x:[0-9]+}/{y:[0-9]+}.png": server.GetTile,
		},
		"POST": {},
		"PUT":  {},
		"OPTIONS": {
			"": options,
		},
	}

	for method, routes := range m {
		for route, handler := range routes {
			localRoute := route
			localHandler := handler
			localMethod := method
			f := makeHttpHandler(localMethod, localRoute, localHandler)

			if localRoute == "" {
				r.Methods(localMethod).HandlerFunc(f)
			} else {
				r.Path(localRoute).Methods(localMethod).HandlerFunc(f)
			}
		}
	}

	return r, nil
}

func makeHttpHandler(localMethod string, localRoute string, handlerFunc HttpApiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeCorsHeaders(w, r)
		if err := handlerFunc(w, r, mux.Vars(r)); err != nil {
			httpError(w, err)
		}
	}
}

func writeCorsHeaders(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	w.Header().Add("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT, OPTIONS")
}

type HttpApiFunc func(w http.ResponseWriter, r *http.Request, vars map[string]string) error

type HTTPServer struct {
	DB       *DB
	tileRoot string
}

func NewHTTPServer(db *DB, tileRoot string) *HTTPServer {
	s := &HTTPServer{
		DB:       db,
		tileRoot: tileRoot,
	}

	return s
}

func writeJSON(w http.ResponseWriter, code int, thing interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	val, err := json.Marshal(thing)
	w.Write(val)
	return err
}

func writeJSONDirect(w http.ResponseWriter, code int, thing []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(thing)
}

func httpError(w http.ResponseWriter, err error) {
	statusCode := http.StatusInternalServerError

	if err != nil {
		log.WithField("err", err).Error("http error")
		http.Error(w, err.Error(), statusCode)
	}
}

func options(w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	w.WriteHeader(http.StatusOK)
	return nil
}

func (s *HTTPServer) GetWorlds(w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	worlds, err := s.DB.GetWorlds()

	if err != nil {
		return err
	}

	writeJSON(w, http.StatusOK, worlds)

	return nil
}

func (s *HTTPServer) GetWorldLayerInfo(w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	worldID, err := strconv.Atoi(vars["worldID"])
	if err != nil {
		return err
	}

	worldInfo, err := s.DB.GetWorldInfo(worldID)

	if err != nil {
		return err
	}

	writeJSON(w, http.StatusOK, worldInfo)

	return nil
}

func (s *HTTPServer) GetCells(w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	layerID, err := strconv.Atoi(vars["layerID"])
	if err != nil {
		return err
	}

	x, err := strconv.ParseFloat(vars["x"], 64)
	if err != nil {
		return err
	}

	y, err := strconv.ParseFloat(vars["y"], 64)
	if err != nil {
		return err
	}

	json, err := s.DB.GetCellJson(layerID, x, y)
	if err != nil {
		return err
	}
	writeJSONDirect(w, http.StatusOK, json)
	return nil
}

func (s *HTTPServer) GetTile(w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	layerID, err := strconv.Atoi(vars["layerID"])
	if err != nil {
		return err
	}

	x, err := strconv.Atoi(vars["x"])
	if err != nil {
		return err
	}

	y, err := strconv.Atoi(vars["y"])
	if err != nil {
		return err
	}

	z, err := strconv.Atoi(vars["z"])
	if err != nil {
		return err
	}

	t, err := s.DB.GetTileRoot(layerID)
	if err != nil {
		log.WithField("err", err).Error("db error")
		w.WriteHeader(http.StatusNotFound)
		return nil
	}

	tile := filepath.Join(s.tileRoot, t, strconv.Itoa(z), strconv.Itoa(x), strconv.Itoa(y)+".png")

	f, err := os.Open(tile)
	defer f.Close()
	if err != nil {
		log.WithField("err", err).Error("file error")
		w.WriteHeader(http.StatusNotFound)
		return nil
	}

	h := make([]byte, 512)
	f.Read(h)
	ct := http.DetectContentType(h)

	stat, _ := f.Stat()
	fs := strconv.FormatInt(stat.Size(), 10)

	w.Header().Set("Content-Type", ct)
	w.Header().Set("Content-Length", fs)

	f.Seek(0, 0)
	io.Copy(w, f)

	return nil
}
