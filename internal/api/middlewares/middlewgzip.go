package middlewares

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func MidlewZip(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		uri := r.RequestURI
		method := r.Method

		// check if client send compressed content in the body (gzip only)
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			var reader io.Reader
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				zap.S().Errorln("Error unzip reques body")
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			reader = gz
			defer func() {
				err := gz.Close()
				zap.S().Errorln("Can't close *gzip.Writer", err)
			}()

			body, err := io.ReadAll(reader)
			if err != nil {
				zap.S().Errorln("Error read data from unzipped reques body")
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// update body with unzipped file
			read := bytes.NewReader(body)
			readCloser := io.NopCloser(read)

			// send to ServeHTTP without encoding
			r.Header.Del("Content-Encoding")
			r.Body = readCloser
		}

		// check if client support gzip
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}

		//Send compressed with gzip unsver
		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			zap.S().Errorln("error during gzip compression")
			_, err = io.WriteString(w, err.Error())
			if err != nil {
				zap.S().Errorln("error during gzip compression: can't write error to Response")
			}
			return
		}
		defer func() {
			err := gz.Close()
			zap.S().Errorln("Can't close *gzip.Writer", err)
		}()

		w.Header().Set("Content-Encoding", "gzip")
		h.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)

		duration := time.Since(start)

		zap.S().Infoln(
			"URI: ", uri,
			"Method: ", method,
			"Duration zip json: ", duration,
		)
	})
}
