package server

import (
	"bytes"
	"context"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
)

type contextKey string

const requestBody contextKey = "body"
const jrpcMethod contextKey = "method"

type jRPCRequest struct {
	Method string `json:"method"`
}

func setContextValue(r *http.Request, key contextKey, val interface{}) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), key, val))
}
func parseBodyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// NOTE: This is necessary, as we are expecting to reread the body later
		// on in the middleware / request chain
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.WithError(err).Error("could not read request body")
			return
		}
		r = setContextValue(r, requestBody, body)
		r.Body = io.NopCloser(bytes.NewReader(body))
		next.ServeHTTP(w, r)
	})
}

func augmentRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// set RPC method
		var j jRPCRequest
		body, ok := r.Context().Value(requestBody).([]byte)
		if !ok {
			log.Error("request body not set on context")
			next.ServeHTTP(w, r)
			return
		}
		err := bijson.Unmarshal(body, &j)
		if err != nil {
			log.WithField("body", string(body)).WithError(err).Error("could not Unmarshal body getJRPCMethod")
			next.ServeHTTP(w, r)
			return
		}
		r = setContextValue(r, jrpcMethod, j.Method)
		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methodStr := r.Context().Value(jrpcMethod)
		if methodStr != "" {
			log.WithFields(log.Fields{
				"RemoteAddr": r.RemoteAddr,
				"RequestURI": r.RequestURI,
				"method":     methodStr,
			}).Info("JRPC Method Requested")
		} else {
			log.WithFields(log.Fields{
				"RemoteAddr": r.RemoteAddr,
				"RequestURI": r.RequestURI,
			}).Info("JRPC Method Requested")
		}
		next.ServeHTTP(w, r)
	})
}
