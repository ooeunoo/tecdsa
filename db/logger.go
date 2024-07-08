package db

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type ResponseWriter struct {
	http.ResponseWriter
	body *bytes.Buffer
}

func (w ResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func LogRequest(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params, _ := json.Marshal(r.URL.Query())
		body, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewReader(body))
		headers, _ := json.Marshal(r.Header)

		responseBody := &bytes.Buffer{}
		responseWriter := &ResponseWriter{
			ResponseWriter: w,
			body:           responseBody,
		}

		handler.ServeHTTP(responseWriter, r)

		requestLog := RequestLog{
			Path:      r.URL.Path,
			Region:    r.Header.Get("X-Region"),
			Params:    string(params),
			Body:      string(body),
			Method:    r.Method,
			Headers:   string(headers),
			IPAddress: r.RemoteAddr,
			Response:  responseBody.String(), // 응답 결과 저장
		}

		result := DB.Create(&requestLog)
		if result.Error != nil {
			log.Printf("Failed to log request: %v", result.Error)
		}
	}
}
