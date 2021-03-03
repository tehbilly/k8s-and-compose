package main

import (
	"context"
	"fmt"
	redis "github.com/go-redis/redis/v8"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

type RedisServer struct {
	client *redis.Client
}

func NewRedisServer(client *redis.Client) *RedisServer {
	// Completely contrived example, just storing environment variables in redis
	env := os.Environ()
	sort.Strings(env)

	for _, e := range env {
		ep := strings.Split(e, "=")
		fmt.Printf("Saving redis entry for '%s': %s\n", ep[0], ep[1])
		if err := client.Set(context.Background(), fmt.Sprintf("env:%s", ep[0]), ep[1], time.Duration(0)).Err(); err != nil {
			fmt.Printf("Unable to set 'env:%s' to %s: %v\n", ep[0], ep[1], err)
			os.Exit(1)
		}
	}

	return &RedisServer{client: client}
}

func (r *RedisServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	keys, err := r.client.Keys(context.Background(), "env:*").Result()
	if err != nil {
		fmt.Fprintln(rw, "Unable to load data from redis:", err)
		return
	}

	for _, key := range keys {
		val, err := r.client.Get(context.Background(), key).Result()
		if err != nil {
			fmt.Fprintf(rw, "Unable to load value for '%s' from redis: %v\n", key, err)
			return
		}
		name := strings.TrimPrefix(key, "env:")
		fmt.Fprintf(rw, "%s = %s\n", name, val)
	}
}

func main() {
	client, err := redisClient()
	if err != nil {
		fmt.Println("Unable to connect to redis:", err)
		os.Exit(1)
	}

	server := NewRedisServer(client)

	addr := listenAddr()
	fmt.Println("Listening on:", addr)

	if err := http.ListenAndServe(addr, server); err != nil {
		panic(err)
	}
}

func redisClient() (*redis.Client, error) {
	redisHost, ok := os.LookupEnv("REDIS_HOST")
	if !ok {
		redisHost = "redis"
	}
	redisPort, ok := os.LookupEnv("REDIS_PORT")
	if !ok {
		redisPort = "6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", redisHost, redisPort),
		OnConnect: func(ctx context.Context, cn *redis.Conn) error {
			fmt.Println("Connected to redis!")
			return nil
		},
	})

	if err := client.FlushDB(context.Background()).Err(); err != nil {
		fmt.Println("Unable to flush DB:", err)
		return nil, err
	}

	return client, nil
}

func listenAddr() string {
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8000"
	}
	return fmt.Sprintf("0.0.0.0:%s", port)
}
