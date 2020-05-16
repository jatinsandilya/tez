package server

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/jatinsandilya/tez/config"
	"github.com/jatinsandilya/tez/redis"
	"github.com/gorilla/mux"
)

var (
	requestTimeout = config.GetConfig().RequestTimeout
)

// handleSetKey godoc
// @Summary Set a value against a key
// @Description Sets a redis key with given value as json
// @Tags cache
// @Accept  json
// @Produce  json
// @Param cache body redis.Request true "Add Key"
// @Success 200 {object} redis.Response
// @Router /v1/cache [post]
func (s *Server) handleSetKey(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	reqBody, err := ioutil.ReadAll(r.Body)

	if err != nil {
		log.Printf("Failed to parse request", err)
		resp := redis.Response{
			Status:  "failure",
			Code:    400,
			Message: "Bad Request",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
	var value redis.Request

	err = json.Unmarshal(reqBody, &value)

	if err != nil {
		log.Printf("Failed to parse request", err)
		resp := redis.Response{
			Status:  "failure",
			Code:    400,
			Message: "Bad Request",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	log.Printf("Value recieved : %+v\n", value)

	done := make(chan bool)
	go s.rp.Set(ctx, done, value.Key, value)

	select {
	case x := <-done:
		resp := redis.Response{
			Status:  "ok",
			Code:    200,
			Message: "Key successfully set",
		}
		if !x {
			log.Printf("Failed to set key")
			resp = redis.Response{
				Status:  "failure",
				Code:    500,
				Message: "Internal Server Error",
			}
		}
		json.NewEncoder(w).Encode(resp)
		return

	case <-ctx.Done():
		log.Printf("Request was cancelled : %s ", ctx.Err().Error())
		resp := redis.Response{
			Status:  "failure",
			Code:    408,
			Message: "Request Timeout.",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
}

// handleGetKey godoc
// @Summary Get value against a key
// @Description Fetch a redis value with key
// @Tags cache
// @Accept  json
// @Produce  json
// @Param key path string true "key to be fetch"
// @Success 200 {object} redis.Response
// @Router /v1/cache/{key} [get]
func (s *Server) handleGetKey(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	params := mux.Vars(r)

	key := params["key"]

	log.Printf("Key recieved : %+v\n", key)

	data := make(chan interface{})
	go s.rp.Get(ctx, data, key)
	select {
	case c := <-data:
		json.NewEncoder(w).Encode(c)
		return

	case <-ctx.Done():
		log.Printf("Request was cancelled : %s ", ctx.Err().Error())
		resp := redis.Response{
			Status:  "failure",
			Code:    408,
			Message: "Request Timeout.",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
}

// handleDeleteKey godoc
// @Summary Delete a key
// @Description Deletes a redis key specified
// @Tags cache
// @Accept  json
// @Produce  json
// @Param key path string true "key to be deleted"
// @Success 200 {object} redis.Response
// @Router /v1/cache/{key} [delete]
func (s *Server) handleDeleteKey(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	params := mux.Vars(r)

	key := params["key"]

	count := make(chan int64)
	go s.rp.Delete(ctx, count, key)
	select {
	case c := <-count:

		if c == int64(0) {
			resp := redis.Response{
				Status:  "ok",
				Code:    200,
				Message: "Key unavailable.",
			}
			json.NewEncoder(w).Encode(resp)
			return
		}

		if c == int64(-1) {
			log.Printf("Failed to delete key")
			resp := redis.Response{
				Status:  "failure",
				Code:    500,
				Message: "Internal Server Error",
			}
			json.NewEncoder(w).Encode(resp)
			return
		}

		resp := redis.Response{
			Status:  "ok",
			Code:    200,
			Message: "Key Deleted",
		}

		json.NewEncoder(w).Encode(resp)
		return

	case <-ctx.Done():
		log.Printf("Request was cancelled : %s ", ctx.Err().Error())
		resp := redis.Response{
			Status:  "failure",
			Code:    408,
			Message: "Request Timeout.",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
}

// handleDeleteValuesWithPrefix godoc
// @Summary Deletes a set of keys with a pattern
// @Description Deletes a set of keys with a pattern
// @Tags cache
// @Accept  json
// @Produce  json
// @Param pattern path string true "keyPattern to be deleted"
// @Success 200 {object} redis.Response
// @Router /v1/cache/pattern/{pattern} [delete]
func (s *Server) handleDeleteKeysWithPattern(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	params := mux.Vars(r)

	key := params["pattern"]

	if strings.TrimSpace(key) == "*" {
		resp := redis.Response{
			Status:  "failure",
			Code:    403,
			Message: "Operation not allowed.",
		}

		json.NewEncoder(w).Encode(resp)
		return
	}

	countChannel := make(chan int64)
	go s.rp.DeleteWithPattern(ctx, countChannel, key)

	select {
	case c := <-countChannel:

		if c == int64(0) {
			resp := redis.Response{
				Status:  "ok",
				Code:    200,
				Message: "Keys unavailable.",
			}
			json.NewEncoder(w).Encode(resp)
			return
		}

		if c == int64(-1) {
			log.Printf("Failed to delete key")
			resp := redis.Response{
				Status:  "failure",
				Code:    500,
				Message: "Internal Server Error",
			}
			json.NewEncoder(w).Encode(resp)
			return
		}

		resp := redis.Response{
			Status:  "ok",
			Code:    200,
			Message: "Key Deleted",
			Payload: c,
		}

		json.NewEncoder(w).Encode(resp)
		return

	case <-ctx.Done():
		log.Printf("Request was cancelled : %s ", ctx.Err().Error())
		resp := redis.Response{
			Status:  "failure",
			Code:    408,
			Message: "Request Timeout.",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

}

func (s *Server) handleUpdateKey(w http.ResponseWriter, r *http.Request) {
	// Todo - Add here.
}

func (s *Server) handleUpdateSubKey(w http.ResponseWriter, r *http.Request) {
	// Todo - Add here.
}

func (s *Server) handleGetValuesWithPattern(w http.ResponseWriter, r *http.Request) {
	// Todo - Add here.
}
func (s *Server) handleSetKeysInBulk(w http.ResponseWriter, r *http.Request) {
	// Todo - Add here.
}

func (s *Server) handleGetKeysWithPattern(w http.ResponseWriter, r *http.Request) {
	// Todo - Add here.
}
