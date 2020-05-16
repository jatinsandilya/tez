package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/jatinsandilya/tez/config"
	"github.com/gomodule/redigo/redis"
	"github.com/nitishm/go-rejson"
	"github.com/nitishm/go-rejson/rjs"
)

// Pool is the shared redis connection
// object to be used across this service
type Pool struct {
	rh *rejson.Handler

	rp redis.Pool
}

// NewPool returns a new redis pool at startup
func NewPool() *Pool {

	config := config.GetConfig()

	var addr = config.RC.Host + ":" + strconv.Itoa(config.RC.Port)

	return &Pool{
		rh: rejson.NewReJSONHandler(),
		rp: redis.Pool{

			MaxIdle:     config.RC.MaxIdle,
			IdleTimeout: config.RC.IdleTimeout,

			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial("tcp", addr)
				if err != nil {
					return nil, err
				}

				return c, err
			},

			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				_, err := c.Do("PING")
				return err
			},
		},
	}
}

// Set sets json data to a particular key
func (p *Pool) Set(ctx context.Context, done chan bool, key string, data Request) {
	conn := p.rp.Get()
	defer conn.Close()
	p.rh.SetRedigoClient(conn)

	err := p.JSONSetWithExpiry(ctx, key, ".", data)
	if err != nil {
		log.Printf("Failed to JSONSet : %s", err)
		done <- false
	}
	done <- true

}

// Get gets the json data present against a key
func (p *Pool) Get(ctx context.Context, data chan interface{}, key string) {
	conn := p.rp.Get()
	p.rh.SetRedigoClient(conn)
	defer conn.Close()

	var dataRead interface{}
	var err error
	done := make(chan bool)
	go func() {
		dataRead, err = p.rh.JSONGet(key, ".")
		if err != nil {
			log.Printf("Failed to JSONGet %v", err)
			done <- false
			return
		}
		done <- true
	}()

	select {
	case x := <-done:
		if !x {
			resp := Response{
				Status:  "failure",
				Code:    500,
				Message: "Internal Server Error",
			}
			data <- resp
			return
		}
		if dataRead == nil {
			resp := Response{
				Status:  "ok",
				Code:    200,
				Message: "Key unavailable.",
			}
			data <- resp
			return
		}
		var value interface{}
		json.Unmarshal(dataRead.([]byte), &value)
		data <- Response{
			Status:  "ok",
			Code:    200,
			Message: "Key available",
			Payload: value.(interface{}),
		}

	case <-ctx.Done():
		log.Println("Closing connection since cancelled.")
		conn.Close()
	}
}

//Delete deletes a key from redis
func (p *Pool) Delete(ctx context.Context, count chan int64, key string) {
	conn := p.rp.Get()
	p.rh.SetRedigoClient(conn)
	defer conn.Close()

	var res interface{}
	var err error
	done := make(chan bool)
	go func() {
		res, err = p.rh.JSONDel(key, ".")
		if err != nil {
			log.Printf("Failed to JSONDel %v", err)
			done <- false
			return
		}
		log.Println("Successfully deleted key : "+key+" : count : %v", res)
		done <- true
	}()

	select {
	case x := <-done:
		if !x {
			count <- -1 //signifies error
			return
		}
		if 0 == res.(int64) {
			count <- 0
			return
		}
		count <- 1
	case <-ctx.Done():
		log.Println("Closing connection since cancelled.")
		conn.Close()
	}
}

// Close closes the connection
// Used in a deferred manner
func (p *Pool) Close() {

	log.Println("Closing redis connection.")
	err := p.rp.Close()
	if err != nil {
		log.Printf("Failed to communicate to redis-server @ %v", err)
	}
}

//JSONSetWithExpiry sets the expiry flag on a json valued key
func (p *Pool) JSONSetWithExpiry(ctx context.Context, key string, path string, obj Request, opts ...rjs.SetOption) error {
	conn := p.rp.Get()
	p.rh.SetRedigoClient(conn)
	defer conn.Close()

	done := make(chan bool)
	var err error
	if obj.Options["expiry"] != nil {
		go func() {
			fmt.Println("Options provided : ", obj.Options)
			data, err := json.Marshal(obj.Value)
			conn.Send("MULTI")
			conn.Send("JSON.SET", key, path, data)
			conn.Send("EXPIRE", key, obj.Options["expiry"])
			// TODO : Add other options as and when required.
			out, err := conn.Do("EXEC")
			if err != nil {
				log.Printf("Error : %s", err)
				done <- false
			}
			fmt.Println(out)
			done <- true

		}()
	} else {
		go func() {
			_, err = p.rh.JSONSet(key, path, obj.Value, opts...)
			if err != nil {
				done <- false
			}
			done <- true

		}()
	}

	select {
	case x := <-done:
		if !x {
			return err
		}
		return nil
	case <-ctx.Done():
		log.Println("Closing connection since cancelled.")
		conn.Close()
		return ctx.Err()
	}
}

// DeleteWithPattern deletes the keys
// with a certain pattern
// Uses unlink for async deletion.
func (p *Pool) DeleteWithPattern(ctx context.Context, countChannel chan int64, pattern string) {

	conn := p.rp.Get()
	p.rh.SetRedigoClient(conn)
	defer conn.Close()
	iter := 0
	count := int64(0)
	done := make(chan bool)
	log.Printf("Pattern used for deletion : %s \n", pattern)

	go func() {
		for {
			arr, err := redis.Values(conn.Do("SCAN", iter, "MATCH", pattern))
			if err != nil {
				done <- false
				log.Errorf("Error while scanning for deletion of '%s' keys : %s", pattern, err)
				return
			}
			iter, _ = redis.Int(arr[0], nil)
			k, _ := redis.Strings(arr[1], nil)
			log.Println("Keys being unlinked: %s", k)
			for _, key := range k {
				resp, err := conn.Do("UNLINK", key)
				if err != nil {
					done <- false
					log.Errorf("Error deleting '%s' keys : %s", pattern, err)
					return
				}
				count += resp.(int64)
			}

			if iter == 0 {
				done <- true
				break
			}
		}
	}()
	select {
	case x := <-done:
		if !x {
			countChannel <- -1 //signifies error
			return
		}
		countChannel <- count
	case <-ctx.Done():
		log.Println("Closing connection since cancelled.")
		conn.Close()
	}
}
