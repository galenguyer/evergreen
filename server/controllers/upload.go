package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/c2h5oh/datasize"
	"github.com/galenguyer/evergreen/core"
	"github.com/galenguyer/evergreen/utils"
)

type UploadController struct {
	MaxFileSize int
	MaxLifetime int
}

// TODO: Add better logging for failure diagnoses
func (c UploadController) HandleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "{\"error\":\"method not allowed\"}", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, int64(c.MaxFileSize))
	if err := r.ParseMultipartForm(int64(c.MaxFileSize)); err != nil {
		http.Error(
			w,
			"{\"error\":\"uploaded file larger than "+datasize.ByteSize(c.MaxFileSize).String()+"\"}",
			http.StatusBadRequest,
		)
		return
	}

	lifetime, err := strconv.Atoi(r.FormValue("lifetime"))
	if err != nil {
		http.Error(
			w,
			"{\"error\":\"no form attribute \\\"lifetime\\\" or lifetime is not a valid number\"}",
			http.StatusBadRequest,
		)
		return
	}
	if lifetime > c.MaxLifetime {
		http.Error(
			w,
			fmt.Sprintf(
				"{\"error\":\"%ds is greater than max lifetime of %ds\"}",
				lifetime,
				c.MaxLifetime,
			),
			http.StatusBadRequest,
		)
		return
	}
	if lifetime <= 0 {
		http.Error(
			w,
			fmt.Sprintf(
				"{\"error\":\"%ds is less than 0s\"}",
				lifetime,
			),
			http.StatusBadRequest,
		)
		return
	}

	// The argument to FormFile must match the name attribute
	// of the file input on the frontend
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(
			w,
			"{\"error\":\"no form attribute \\\"file\\\"\"}",
			http.StatusBadRequest,
		)
		return
	}
	defer file.Close()

	err = os.MkdirAll("./uploads", os.ModePerm)
	if err != nil {
		http.Error(
			w,
			"{\"error\":\"internal server error\"}",
			http.StatusInternalServerError,
		)
		return
	}
	err = os.MkdirAll("./metadata", os.ModePerm)
	if err != nil {
		http.Error(
			w,
			"{\"error\":\"internal server error\"}",
			http.StatusInternalServerError,
		)
		return
	}

	fileName := utils.GenerateName(6)

	metadata := &core.Metadata{
		Filename: fmt.Sprintf("%s%s", fileName, filepath.Ext(fileHeader.Filename)),
		Expiry:   time.Now().Add(time.Duration(lifetime) * time.Second),
	}

	jsonMetadata, _ := json.Marshal(metadata)
	ioutil.WriteFile(fmt.Sprintf("./metadata/%s.json", fileName), jsonMetadata, os.ModePerm)

	dst, err := os.Create(fmt.Sprintf("./uploads/%s%s", fileName, filepath.Ext(fileHeader.Filename)))
	if err != nil {
		http.Error(
			w,
			"{\"error\":\"internal server error\"}",
			http.StatusInternalServerError,
		)
		return
	}

	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(
			w,
			"{\"error\":\"internal server error\"}",
			http.StatusInternalServerError,
		)
		return
	}

	log.Println("[upload] saved file", string(jsonMetadata))
	fmt.Fprintf(
		w,
		"{\"message\":\"file upload successful\",\"file\":\"%s\",\"expiry\":%d}\n",
		filepath.Base(dst.Name()),
		metadata.Expiry.Unix(),
	)
}
