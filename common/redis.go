package common

import (
	"context"
	"one-api/common/config"
	"one-api/common/logger"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

var RDB *redis.Client
var RedisEnabled = false

// InitRedisClient This function is called after init()
func InitRedisClient() (err error) {
	redisConn := viper.GetString("redis_conn_string")

	if redisConn == "" {
		logger.SysLog("REDIS_CONN_STRING not set, Redis is not enabled")
		return nil
	}
	if viper.GetInt("sync_frequency") == 0 {
		logger.SysLog("SYNC_FREQUENCY not set, Redis is disabled")
		return nil
	}
	logger.SysLog("Redis is enabled")
	opt, err := redis.ParseURL(redisConn)
	if err != nil {
		logger.FatalLog("failed to parse Redis connection string: " + err.Error())
		return
	}
	RDB = redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = RDB.Ping(ctx).Result()
	if err != nil {
		logger.FatalLog("Redis ping test failed: " + err.Error())
	} else {
		RedisEnabled = true
		// for compatibility with old versions
		config.MemoryCacheEnabled = true
	}

	return err
}

func ParseRedisOption() *redis.Options {
	opt, err := redis.ParseURL(viper.GetString("redis_conn_string"))
	if err != nil {
		logger.FatalLog("failed to parse Redis connection string: " + err.Error())
	}
	return opt
}

func RedisSet(key string, value string, expiration time.Duration) error {
	ctx := context.Background()
	return RDB.Set(ctx, key, value, expiration).Err()
}

func RedisGet(key string) (string, error) {
	ctx := context.Background()
	return RDB.Get(ctx, key).Result()
}

func RedisDel(key string) error {
	ctx := context.Background()
	return RDB.Del(ctx, key).Err()
}

func RedisDecrease(key string, value int64) error {
	ctx := context.Background()
	return RDB.DecrBy(ctx, key, value).Err()
}
