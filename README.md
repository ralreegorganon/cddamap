# cddamap

Mapping the Cataclysm, one world at a time...

## Generating an image from a world save
 
1. Install cddamapgen: `go get -u github.com/ralreegorganon/cddamap/cmd/cddamapgen`
2. Generate an image for ground level: `cddamapgen -g ~/code/Cataclysm-DDA -s ~/code/Cataclysm-DDA/save/Bruce -o ~/Desktop -ir -l 10`

```
Usage:
  cddamapgen [OPTIONS]

Application Options:
  -g, --game=             Cataclysm: DDA game root directory
  -s, --save=             Game save directory to process
  -o, --output=           Output folder
  -t, --text              Render to text files
  -i, --images            Render to images
  -l, --layer=            Layer to render, 0-20. Repeat flag for multiple layers or omit for all.
  -c, --connectionString= PostGIS database connection string
  -r, --terrain           Render terrain
  -e, --seen              Render seen
  -d, --seensolid         Render seen as a solid overlay
  -C, --cities            Render city names
  -k, --skipempty         Skip rendering empty layers
  -O, --overmap=          Overmap filter to limit included overmaps

Help Options:
  -h, --help              Show this help message
```