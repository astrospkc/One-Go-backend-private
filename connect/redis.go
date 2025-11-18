package connect

import (
	"context"
	"fmt"
	"gobackend/env"

	"github.com/valkey-io/valkey-go"
)

var (
	RedisClient valkey.Client  
	Rctx        = context.Background() 
)

func InitRedis() {
	envs:=env.NewEnv()
	redisURI:=envs.AIVEN_KEY
	
	if redisURI == "" {
		panic("AIVEN_SERVICE_URI not set")
	}

	// Parse the URI directly (handles TLS, password, username, host, port)
	client, err := valkey.NewClient(valkey.MustParseURL(redisURI))
	if err != nil {
		panic("Failed to initialize Redis client: " + err.Error())
	}

	RedisClient = client

	// Ping test to confirm successful connection
	err = RedisClient.Do(Rctx, RedisClient.B().Ping().Build()).Error()
	if err != nil {
		panic(fmt.Sprintf("Redis connection ping failed: %v", err))
	}

	fmt.Println("✅ Connected to Redis")
}