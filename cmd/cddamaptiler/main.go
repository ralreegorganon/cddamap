package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"

	flags "github.com/jessevdk/go-flags"
	"github.com/ralreegorganon/cddamap/internal/tile"
	log "github.com/sirupsen/logrus"
)

var opts struct {
	ImageDirectory string   `short:"I" long:"imageDirectory" description:"Image directory to tile"`
	ImageFiles     []string `short:"i" long:"images" description:"Images to tile"`
	Resume         bool     `short:"z" long:"resume" description:"Resume tile building, instead of overwriting"`
}

func init() {
	f := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(f)
}

func main() {
	go func() {
		http.ListenAndServe(":8080", nil)
	}()

	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}

	if opts.ImageDirectory != "" {
		files, err := filepath.Glob(filepath.Join(opts.ImageDirectory, "*.png"))
		if err != nil {
			log.Fatal(err)
		}

		for _, f := range files {
			err := tile.ChopChop(f, opts.Resume)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	for _, f := range opts.ImageFiles {
		err := tile.ChopChop(f, opts.Resume)
		if err != nil {
			log.Fatal(err)
		}
	}
}
