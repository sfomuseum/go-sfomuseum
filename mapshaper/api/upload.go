package api

import (
	"io"
	"io/ioutil"
	_ "log"
	"net/http"
	"os"
)

type uploadOptions struct {
	MaxSize int64
}

func uploadWithRequest(rsp http.ResponseWriter, req *http.Request, opts *uploadOptions) (*os.File, error) {

	req.Body = http.MaxBytesReader(rsp, req.Body, opts.MaxSize)

	defer req.Body.Close()

	tmp_fh, err := ioutil.TempFile("", "mapshaper-*.geojson")

	if err != nil {
		return nil, err
	}

	_, err = io.Copy(tmp_fh, req.Body)

	if err != nil {
		return nil, err
	}

	return tmp_fh, nil
}
