Assuming gdal and gdal-python bindings are installed, run the following to create the tiles for a given layer image
python3 gdal2tiles-leaflet.py -l -p raster -r lanczos -w none o_10.png tiles