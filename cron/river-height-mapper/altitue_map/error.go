package altitue_map

import "fmt"

type RasterNotFound struct {
	lat int
	lon int
	firstTime bool
}

func (this *RasterNotFound) FirstTime() bool {
	return this.firstTime
}

func (this *RasterNotFound) Error() string {
	return fmt.Sprintf("Raster not found for lat %d lon %d", this.lat, this.lon)
}
