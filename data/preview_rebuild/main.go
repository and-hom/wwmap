package main

import (
	"github.com/and-hom/wwmap/lib/config"
	"github.com/and-hom/wwmap/lib/blob"
	log "github.com/Sirupsen/logrus"
	"image"
	"bytes"
	"image/png"
	"github.com/and-hom/wwmap/lib/util"
	"github.com/and-hom/wwmap/backend/handler"
	"golang.org/x/image/draw"
)

func main() {
	configuration := config.Load("")
	configuration.ChangeLogLevel()

	imgStorage := blob.BasicFsStorage{
		BaseDir:configuration.ImgStorage.Full.Dir,
	}
	imgPreviewStorage := blob.BasicFsStorage{
		BaseDir:configuration.ImgStorage.Preview.Dir,
	}
	ids, err := imgStorage.ListIds()
	if err != nil {
		log.Fatalf("Can not list images: %v", err)
	}

	for _, id := range ids {
		sourceReader, err := imgStorage.Read(id)
		if err != nil {
			log.Errorf("Can not read image %s: %v", id, err)
			continue
		}
		sourceImage, _, err := image.Decode(sourceReader)
		if err != nil {
			log.Errorf("Can not decode image file %s: %v", id, err)
			continue
		}

		rect, _ := util.PreviewRect(sourceImage.Bounds(), handler.PREVIEW_MAX_WIDTH, handler.PREVIEW_MAX_HEIGHT)

		resized := image.NewRGBA(rect)
		draw.ApproxBiLinear.Scale(resized, rect, sourceImage, sourceImage.Bounds(), draw.Over, nil)
		var b bytes.Buffer
		err = png.Encode(&b, resized)
		if err != nil {
			log.Errorf("Can not encode image file %s: %v", id, err)
			continue
		}
		err = imgPreviewStorage.Store(id, &b)
		if err != nil {
			log.Errorf("Can not store image file %s: %v", id, err)
			continue
		}
		log.Infof("Image for %s successfully stored", id)
	}

}
