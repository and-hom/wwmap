package main

import (
	"net/http"
	"io/ioutil"
	"os"
	"io"
)

type PictureHandler struct {
	Handler
}

func (this *PictureHandler) PictureMetadataHandler(w http.ResponseWriter, r *http.Request) {
	corsHeaders(w, "POST")

	requestBody := r.Body
	defer requestBody.Close()

	imgUrl, err := ioutil.ReadAll(requestBody)
	if err != nil {
		this.onError(w, err, "Can not read request body", http.StatusBadRequest)
		return
	}

	imgResp, err := http.Get(string(imgUrl))
	if err != nil {
		this.onError(w, err, "Can not fetch image", 422)
		return
	}

	defer imgResp.Body.Close()

	tmpFile, err := ioutil.TempFile(os.TempDir(), "img")
	if err != nil {
		this.onError500(w, err, "Can not create temp file")
		return
	}
	defer this.CloseAndRemove(tmpFile)

	_, err = io.Copy(tmpFile, imgResp.Body)
	if err != nil {
		this.onError500(w, err, "Can not fetch image from server: " + string(imgUrl))
		return
	}
	_, err = tmpFile.Seek(0, os.SEEK_SET)
	if err != nil {
		this.onError500(w, err, "Can not seek on img file")
		return
	}

	imgData, err := GetImgProperties(tmpFile)
	if err != nil {
		this.onError500(w, err, "Can not get img properties")
		return
	}

	w.Write([]byte(this.JsonStr(imgData, "{}")))
}

