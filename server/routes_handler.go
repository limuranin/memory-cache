package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"memory-cache/logger"
	"memory-cache/msgtypes"

	"github.com/gorilla/mux"
)

const (
	keyParam    = "key"
	mapKeyParam = "mapKey"
	indexParam  = "index"
)

type routesHandler struct {
	router *mux.Router
	cacher Cacher
}

func newRoutesHandler(router *mux.Router, cacher Cacher) *routesHandler {
	return &routesHandler{
		router: router,
		cacher: cacher,
	}
}

func (rh *routesHandler) registerRoutes() {
	rh.router.
		Name("Set").
		Path("/set").
		Methods(http.MethodPost, http.MethodOptions).
		Handler(rh.SetHandler())

	rh.router.
		Name("Get").
		Path(fmt.Sprintf("/get/{%v}", keyParam)).
		Methods(http.MethodGet).
		Handler(rh.GetHandler())

	rh.router.
		Name("GetListElem").
		Path(fmt.Sprintf("/getListElem/{%v}/{%v:[0-9]+}", keyParam, indexParam)).
		Methods(http.MethodGet).
		Handler(rh.GetListElemHandler())

	rh.router.
		Name("GetMapElemValue").
		Path(fmt.Sprintf("/getMapElemValue/{%v}/{%v}", keyParam, mapKeyParam)).
		Methods(http.MethodGet).
		HandlerFunc(rh.GetMapElemHandler())

	rh.router.
		Name("Remove").
		Path(fmt.Sprintf("/remove/{%v}", keyParam)).
		Methods(http.MethodDelete, http.MethodOptions).
		HandlerFunc(rh.RemoveHandler())

	rh.router.
		Name("Keys").
		Path("/keys").
		Methods(http.MethodGet).
		HandlerFunc(rh.KeysHandler())

	rh.router.Use(requestLoggingMiddleware)
	rh.router.Use(mux.CORSMethodMiddleware(rh.router))
	rh.router.Use(corsMiddleware)
}

func (rh *routesHandler) SetHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil {
			responseError(w, errors.New("nil request body"), http.StatusBadRequest)
			return
		}

		setReq := &msgtypes.SetReq{}
		if err := json.NewDecoder(r.Body).Decode(setReq); err != nil {
			responseError(w, err, http.StatusBadRequest)
			return
		}

		logger.Debugf("Set key '%v' and value '%+v' with ttl '%v'",
			setReq.Key, setReq.Value, time.Duration(setReq.Ttl))
		if err := rh.cacher.Set(setReq.Key, setReq.Value, time.Duration(setReq.Ttl)); err != nil {
			responseError(w, err, http.StatusInternalServerError)
			return
		}

		responseSuccessStatus(w)
	}
}

func (rh *routesHandler) GetHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		key := params[keyParam]

		value, err := rh.cacher.Get(key)
		if err != nil {
			responseError(w, err, http.StatusInternalServerError)
			return
		}

		resp := &msgtypes.ValueResp{
			Value: value,
		}
		responseSuccess(w, resp)
	}
}

func (rh *routesHandler) GetListElemHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		key := params[keyParam]

		indexStr := params[indexParam]
		index, err := strconv.Atoi(indexStr)
		if err != nil {
			responseError(w, err, http.StatusBadRequest)
			return
		}

		value, err := rh.cacher.GetListElem(key, index)
		if err != nil {
			responseError(w, err, http.StatusInternalServerError)
			return
		}

		resp := &msgtypes.ValueResp{
			Value: value,
		}
		responseSuccess(w, resp)
	}
}

func (rh *routesHandler) GetMapElemHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		key := params[keyParam]
		mapKey := params[mapKeyParam]

		value, err := rh.cacher.GetMapElemValue(key, mapKey)
		if err != nil {
			responseError(w, err, http.StatusInternalServerError)
			return
		}

		resp := &msgtypes.ValueResp{
			Value: value,
		}
		responseSuccess(w, resp)
	}
}

func (rh *routesHandler) RemoveHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		key := params[keyParam]

		if err := rh.cacher.Remove(key); err != nil {
			responseError(w, err, http.StatusInternalServerError)
			return
		}

		responseSuccessStatus(w)
	}
}

func (rh *routesHandler) KeysHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		keys, err := rh.cacher.Keys()
		if err != nil {
			responseError(w, err, http.StatusInternalServerError)
			return
		}

		resp := &msgtypes.KeysResp{
			Keys: keys,
		}
		responseSuccess(w, resp)
	}
}

func requestLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
			logger.Infof("Request method: '%s',  request URI: '%s'",
				r.Method, r.RequestURI)
		})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			if r.Method == http.MethodOptions {
				return
			}

			next.ServeHTTP(w, r)
		})
}

func responseSuccessStatus(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func responseSuccess(w http.ResponseWriter, resp interface{}) {
	responseSuccessStatus(w)

	data, err := json.Marshal(resp)
	if err != nil {
		logger.Errorf("marshal response error: %v", err)
		return
	}

	if _, err := w.Write(data); err != nil {
		logger.Errorf("write data error: %v", err)
		return
	}
}

func responseError(w http.ResponseWriter, err error, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp := msgtypes.ErrorResp{
		Error: err.Error(),
	}

	data, err := json.Marshal(resp)
	if err != nil {
		logger.Errorf("marshal response error: %v", err)
		return
	}

	if _, err := w.Write(data); err != nil {
		logger.Errorf("write data error: %v", err)
		return
	}
}
