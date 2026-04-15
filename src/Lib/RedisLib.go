package Lib

import (
	"encoding/base64"
	"encoding/json"
	"github.com/gomodule/redigo/redis"
	"log"
	"data-fetcher-api/src/Helper"
	"time"
)

const EXPIRE_SESSION = 60 * 60
const REDIS_NETWORK = "tcp"

func redisPool() (*redis.Pool, error) {
	parameter, err := Helper.GetParameter()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &redis.Pool{
		MaxIdle:     10,
		MaxActive:   100,
		IdleTimeout: 5 * time.Minute,
		Dial: func() (redis.Conn, error) {
			return redis.Dial(REDIS_NETWORK, parameter.Parameters.RedisAddress)
		},
	}, nil
}

func SetUser(value int, roles interface{}, ip string) (interface{}, error) {
	pool, err := redisPool()
	if err != nil {
		return "", err
	}
	conn := pool.Get()
	defer conn.Close()

	parameter, err := Helper.GetParameter()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	prefix := parameter.Parameters.DfSessionPrefix

	encodedIP := base64.StdEncoding.EncodeToString([]byte(ip))
	key := prefix + encodedIP

	redisMembers := []interface{}{value, roles}

	serializedRedisMembers, err := json.Marshal(redisMembers)
	if err != nil {
		return "", err
	}

	err = conn.Send("SET", key, serializedRedisMembers)
	if err != nil {
		return "", err
	}

	err = conn.Send("EXPIRE", key, EXPIRE_SESSION)
	if err != nil {
		return "", err
	}

	err = conn.Flush()
	if err != nil {
		return "", err
	}

	redisValue, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		return "", err
	}

	err = conn.Flush()
	if err != nil {
		return "", err
	}

	var decodedRedisValue []interface{}
	err = json.Unmarshal(redisValue, &decodedRedisValue)
	if err != nil {
		return "", err
	}
	return decodedRedisValue, nil
}

func GetValue(key string) (string, error) {

	pool, err := redisPool()
	if err != nil {
		return "", err
	}
	conn := pool.Get()
	defer conn.Close()

	redisValue, err := redis.String(conn.Do("GET", key))
	if err != nil {
		return "", err
	}
	return redisValue, nil
}
