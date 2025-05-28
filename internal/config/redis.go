// internal/config/redis.go
package config

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

// Redis client instances
var (
	RedisClient        *redis.Client
	RedisClusterClient *redis.ClusterClient
	RedisPubSub        *redis.PubSub
)

// RedisManager manages Redis connections and operations
type RedisManager struct {
	client        *redis.Client
	clusterClient *redis.ClusterClient
	pubSub        *redis.PubSub
	config        RedisConfig
	isCluster     bool
}

// NewRedisManager creates a new Redis manager instance
func NewRedisManager(config RedisConfig) *RedisManager {
	return &RedisManager{
		config:    config,
		isCluster: config.EnableCluster,
	}
}

// InitRedis initializes Redis connection
func InitRedis() error {
	config := GetConfig().Redis

	manager := NewRedisManager(config)
	if err := manager.Connect(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// Set global instances
	if manager.isCluster {
		RedisClusterClient = manager.clusterClient
	} else {
		RedisClient = manager.client
	}

	log.Println("Redis connected successfully")
	return nil
}

// Connect establishes Redis connection
func (rm *RedisManager) Connect() error {
	if rm.isCluster {
		return rm.connectCluster()
	}
	return rm.connectSingle()
}

// connectSingle connects to a single Redis instance
func (rm *RedisManager) connectSingle() error {
	var addr string
	if rm.config.URL != "" {
		// Parse Redis URL
		opts, err := redis.ParseURL(rm.config.URL)
		if err != nil {
			return fmt.Errorf("failed to parse Redis URL: %w", err)
		}
		rm.client = redis.NewClient(opts)
	} else {
		addr = rm.config.Host + ":" + rm.config.Port

		opts := &redis.Options{
			Addr:            addr,
			Password:        rm.config.Password,
			DB:              rm.config.Database,
			MaxRetries:      rm.config.MaxRetries,
			MinRetryBackoff: rm.config.MinRetryBackoff,
			MaxRetryBackoff: rm.config.MaxRetryBackoff,
			DialTimeout:     rm.config.DialTimeout,
			ReadTimeout:     rm.config.ReadTimeout,
			WriteTimeout:    rm.config.WriteTimeout,
			PoolSize:        rm.config.PoolSize,
			MinIdleConns:    rm.config.MinIdleConns,
			MaxConnAge:      rm.config.MaxConnAge,
			PoolTimeout:     rm.config.PoolTimeout,
			IdleTimeout:     rm.config.IdleTimeout,
		}

		rm.client = redis.NewClient(opts)
	}

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rm.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to ping Redis: %w", err)
	}

	return nil
}

// connectCluster connects to Redis cluster
func (rm *RedisManager) connectCluster() error {
	if len(rm.config.ClusterAddresses) == 0 {
		return fmt.Errorf("no cluster addresses provided")
	}

	opts := &redis.ClusterOptions{
		Addrs:           rm.config.ClusterAddresses,
		Password:        rm.config.Password,
		MaxRetries:      rm.config.MaxRetries,
		MinRetryBackoff: rm.config.MinRetryBackoff,
		MaxRetryBackoff: rm.config.MaxRetryBackoff,
		DialTimeout:     rm.config.DialTimeout,
		ReadTimeout:     rm.config.ReadTimeout,
		WriteTimeout:    rm.config.WriteTimeout,
		PoolSize:        rm.config.PoolSize,
		MinIdleConns:    rm.config.MinIdleConns,
		MaxConnAge:      rm.config.MaxConnAge,
		PoolTimeout:     rm.config.PoolTimeout,
		IdleTimeout:     rm.config.IdleTimeout,
	}

	rm.clusterClient = redis.NewClusterClient(opts)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rm.clusterClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to ping Redis cluster: %w", err)
	}

	return nil
}

// GetRedisClient returns the appropriate Redis client
func GetRedisClient() redis.Cmdable {
	if RedisClusterClient != nil {
		return RedisClusterClient
	}
	return RedisClient
}

// Cache operations

// Set stores data in Redis with expiration
func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	client := GetRedisClient()
	if client == nil {
		return fmt.Errorf("Redis client not initialized")
	}

	var data []byte
	var err error

	switch v := value.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		data, err = json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %w", err)
		}
	}

	return client.Set(ctx, key, data, expiration).Err()
}

// Get retrieves data from Redis
func Get(ctx context.Context, key string) (string, error) {
	client := GetRedisClient()
	if client == nil {
		return "", fmt.Errorf("Redis client not initialized")
	}

	result, err := client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("key not found")
	}
	return result, err
}

// GetJSON retrieves and unmarshals JSON data from Redis
func GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := Get(ctx, key)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(data), dest)
}

// SetJSON marshals and stores JSON data in Redis
func SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return Set(ctx, key, value, expiration)
}

// Delete removes keys from Redis
func Delete(ctx context.Context, keys ...string) error {
	client := GetRedisClient()
	if client == nil {
		return fmt.Errorf("Redis client not initialized")
	}

	return client.Del(ctx, keys...).Err()
}

// Exists checks if keys exist in Redis
func Exists(ctx context.Context, keys ...string) (int64, error) {
	client := GetRedisClient()
	if client == nil {
		return 0, fmt.Errorf("Redis client not initialized")
	}

	return client.Exists(ctx, keys...).Result()
}

// Expire sets expiration for a key
func Expire(ctx context.Context, key string, expiration time.Duration) error {
	client := GetRedisClient()
	if client == nil {
		return fmt.Errorf("Redis client not initialized")
	}

	return client.Expire(ctx, key, expiration).Err()
}

// TTL returns the time to live for a key
func TTL(ctx context.Context, key string) (time.Duration, error) {
	client := GetRedisClient()
	if client == nil {
		return 0, fmt.Errorf("Redis client not initialized")
	}

	return client.TTL(ctx, key).Result()
}

// Increment increments a key's value
func Increment(ctx context.Context, key string) (int64, error) {
	client := GetRedisClient()
	if client == nil {
		return 0, fmt.Errorf("Redis client not initialized")
	}

	return client.Incr(ctx, key).Result()
}

// IncrementBy increments a key's value by a specific amount
func IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	client := GetRedisClient()
	if client == nil {
		return 0, fmt.Errorf("Redis client not initialized")
	}

	return client.IncrBy(ctx, key, value).Result()
}

// Session management

// SetSession stores session data
func SetSession(ctx context.Context, sessionID string, data interface{}, expiration time.Duration) error {
	key := "session:" + sessionID
	return SetJSON(ctx, key, data, expiration)
}

// GetSession retrieves session data
func GetSession(ctx context.Context, sessionID string, dest interface{}) error {
	key := "session:" + sessionID
	return GetJSON(ctx, key, dest)
}

// DeleteSession removes session data
func DeleteSession(ctx context.Context, sessionID string) error {
	key := "session:" + sessionID
	return Delete(ctx, key)
}

// RefreshSession updates session expiration
func RefreshSession(ctx context.Context, sessionID string, expiration time.Duration) error {
	key := "session:" + sessionID
	return Expire(ctx, key, expiration)
}

// Rate limiting

// RateLimitCheck checks and increments rate limit counter
func RateLimitCheck(ctx context.Context, key string, limit int, window time.Duration) (bool, int, error) {
	client := GetRedisClient()
	if client == nil {
		return false, 0, fmt.Errorf("Redis client not initialized")
	}

	pipe := client.Pipeline()

	// Get current count
	countCmd := pipe.Get(ctx, key)

	// Increment counter
	incrCmd := pipe.Incr(ctx, key)

	// Set expiration if key is new
	pipe.Expire(ctx, key, window)

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return false, 0, fmt.Errorf("failed to execute rate limit pipeline: %w", err)
	}

	// Get current count
	currentCount := int(incrCmd.Val())

	// If this is the first request, the counter was just created
	if countCmd.Err() == redis.Nil {
		return true, currentCount, nil
	}

	// Check if limit exceeded
	allowed := currentCount <= limit
	return allowed, currentCount, nil
}

// ResetRateLimit resets rate limit counter
func ResetRateLimit(ctx context.Context, key string) error {
	return Delete(ctx, key)
}

// Cache utilities

// GetMultiple retrieves multiple keys
func GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	client := GetRedisClient()
	if client == nil {
		return nil, fmt.Errorf("Redis client not initialized")
	}

	values, err := client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for i, key := range keys {
		if values[i] != nil {
			result[key] = values[i].(string)
		}
	}

	return result, nil
}

// SetMultiple stores multiple key-value pairs
func SetMultiple(ctx context.Context, pairs map[string]interface{}, expiration time.Duration) error {
	client := GetRedisClient()
	if client == nil {
		return fmt.Errorf("Redis client not initialized")
	}

	pipe := client.Pipeline()

	for key, value := range pairs {
		var data []byte
		var err error

		switch v := value.(type) {
		case string:
			data = []byte(v)
		case []byte:
			data = v
		default:
			data, err = json.Marshal(value)
			if err != nil {
				return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
			}
		}

		pipe.Set(ctx, key, data, expiration)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// HashSet stores data in Redis hash
func HashSet(ctx context.Context, key string, fields map[string]interface{}) error {
	client := GetRedisClient()
	if client == nil {
		return fmt.Errorf("Redis client not initialized")
	}

	return client.HMSet(ctx, key, fields).Err()
}

// HashGet retrieves data from Redis hash
func HashGet(ctx context.Context, key string, fields ...string) (map[string]string, error) {
	client := GetRedisClient()
	if client == nil {
		return nil, fmt.Errorf("Redis client not initialized")
	}

	if len(fields) == 0 {
		return client.HGetAll(ctx, key).Result()
	}

	values, err := client.HMGet(ctx, key, fields...).Result()
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for i, field := range fields {
		if values[i] != nil {
			result[field] = values[i].(string)
		}
	}

	return result, nil
}

// List operations

// ListPush pushes elements to the left of a list
func ListPush(ctx context.Context, key string, values ...interface{}) error {
	client := GetRedisClient()
	if client == nil {
		return fmt.Errorf("Redis client not initialized")
	}

	return client.LPush(ctx, key, values...).Err()
}

// ListPop pops element from the left of a list
func ListPop(ctx context.Context, key string) (string, error) {
	client := GetRedisClient()
	if client == nil {
		return "", fmt.Errorf("Redis client not initialized")
	}

	return client.LPop(ctx, key).Result()
}

// ListRange returns a range of elements from a list
func ListRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	client := GetRedisClient()
	if client == nil {
		return nil, fmt.Errorf("Redis client not initialized")
	}

	return client.LRange(ctx, key, start, stop).Result()
}

// ListLength returns the length of a list
func ListLength(ctx context.Context, key string) (int64, error) {
	client := GetRedisClient()
	if client == nil {
		return 0, fmt.Errorf("Redis client not initialized")
	}

	return client.LLen(ctx, key).Result()
}

// Set operations

// SetAdd adds members to a set
func SetAdd(ctx context.Context, key string, members ...interface{}) error {
	client := GetRedisClient()
	if client == nil {
		return fmt.Errorf("Redis client not initialized")
	}

	return client.SAdd(ctx, key, members...).Err()
}

// SetRemove removes members from a set
func SetRemove(ctx context.Context, key string, members ...interface{}) error {
	client := GetRedisClient()
	if client == nil {
		return fmt.Errorf("Redis client not initialized")
	}

	return client.SRem(ctx, key, members...).Err()
}

// SetMembers returns all members of a set
func SetMembers(ctx context.Context, key string) ([]string, error) {
	client := GetRedisClient()
	if client == nil {
		return nil, fmt.Errorf("Redis client not initialized")
	}

	return client.SMembers(ctx, key).Result()
}

// SetIsMember checks if a member exists in a set
func SetIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	client := GetRedisClient()
	if client == nil {
		return false, fmt.Errorf("Redis client not initialized")
	}

	return client.SIsMember(ctx, key, member).Result()
}

// Pub/Sub operations

// InitPubSub initializes pub/sub connection
func InitPubSub(channels ...string) error {
	client := GetRedisClient()
	if client == nil {
		return fmt.Errorf("Redis client not initialized")
	}

	// For cluster, we need to use the cluster client
	if RedisClusterClient != nil {
		RedisPubSub = RedisClusterClient.Subscribe(context.Background(), channels...)
	} else {
		RedisPubSub = RedisClient.Subscribe(context.Background(), channels...)
	}

	// Test the subscription
	_, err := RedisPubSub.Receive(context.Background())
	return err
}

// Publish publishes a message to a channel
func Publish(ctx context.Context, channel string, message interface{}) error {
	client := GetRedisClient()
	if client == nil {
		return fmt.Errorf("Redis client not initialized")
	}

	var data string
	switch v := message.(type) {
	case string:
		data = v
	default:
		jsonData, err := json.Marshal(message)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}
		data = string(jsonData)
	}

	return client.Publish(ctx, channel, data).Err()
}

// Subscribe subscribes to channels
func Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	client := GetRedisClient()
	if client == nil {
		return nil
	}

	if RedisClusterClient != nil {
		return RedisClusterClient.Subscribe(ctx, channels...)
	}
	return RedisClient.Subscribe(ctx, channels...)
}

// Unsubscribe unsubscribes from channels
func Unsubscribe(pubsub *redis.PubSub, channels ...string) error {
	return pubsub.Unsubscribe(context.Background(), channels...)
}

// Cache invalidation

// InvalidatePattern deletes keys matching a pattern
func InvalidatePattern(ctx context.Context, pattern string) error {
	client := GetRedisClient()
	if client == nil {
		return fmt.Errorf("Redis client not initialized")
	}

	// For cluster, we need to handle this differently
	if RedisClusterClient != nil {
		return invalidatePatternCluster(ctx, pattern)
	}

	return invalidatePatternSingle(ctx, pattern)
}

// invalidatePatternSingle invalidates pattern for single Redis instance
func invalidatePatternSingle(ctx context.Context, pattern string) error {
	keys, err := RedisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return RedisClient.Del(ctx, keys...).Err()
	}

	return nil
}

// invalidatePatternCluster invalidates pattern for Redis cluster
func invalidatePatternCluster(ctx context.Context, pattern string) error {
	// In cluster mode, we need to scan all nodes
	err := RedisClusterClient.ForEachMaster(ctx, func(ctx context.Context, master *redis.Client) error {
		keys, err := master.Keys(ctx, pattern).Result()
		if err != nil {
			return err
		}

		if len(keys) > 0 {
			return master.Del(ctx, keys...).Err()
		}

		return nil
	})

	return err
}

// Health check

// HealthCheck checks Redis connection health
func HealthCheck(ctx context.Context) error {
	client := GetRedisClient()
	if client == nil {
		return fmt.Errorf("Redis client not initialized")
	}

	return client.Ping(ctx).Err()
}

// GetStats returns Redis connection statistics
func GetStats() map[string]interface{} {
	stats := make(map[string]interface{})

	if RedisClient != nil {
		poolStats := RedisClient.PoolStats()
		stats["single_instance"] = map[string]interface{}{
			"hits":        poolStats.Hits,
			"misses":      poolStats.Misses,
			"timeouts":    poolStats.Timeouts,
			"total_conns": poolStats.TotalConns,
			"idle_conns":  poolStats.IdleConns,
			"stale_conns": poolStats.StaleConns,
		}
	}

	if RedisClusterClient != nil {
		poolStats := RedisClusterClient.PoolStats()
		stats["cluster"] = map[string]interface{}{
			"hits":        poolStats.Hits,
			"misses":      poolStats.Misses,
			"timeouts":    poolStats.Timeouts,
			"total_conns": poolStats.TotalConns,
			"idle_conns":  poolStats.IdleConns,
			"stale_conns": poolStats.StaleConns,
		}
	}

	return stats
}

// Cleanup and shutdown

// Close closes Redis connections
func Close() error {
	var errors []string

	if RedisPubSub != nil {
		if err := RedisPubSub.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("pubsub close error: %v", err))
		}
	}

	if RedisClient != nil {
		if err := RedisClient.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("client close error: %v", err))
		}
	}

	if RedisClusterClient != nil {
		if err := RedisClusterClient.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("cluster client close error: %v", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("Redis close errors: %s", strings.Join(errors, ", "))
	}

	log.Println("Redis connections closed successfully")
	return nil
}

// Utility functions

// GenerateKey generates a prefixed cache key
func GenerateKey(prefix string, parts ...string) string {
	allParts := append([]string{prefix}, parts...)
	return strings.Join(allParts, ":")
}

// ParseKey parses a cache key into its parts
func ParseKey(key string) []string {
	return strings.Split(key, ":")
}

// ScanKeys scans for keys matching a pattern with cursor-based pagination
func ScanKeys(ctx context.Context, pattern string, count int64) ([]string, error) {
	client := GetRedisClient()
	if client == nil {
		return nil, fmt.Errorf("Redis client not initialized")
	}

	var allKeys []string
	var cursor uint64

	for {
		var keys []string
		var err error

		if RedisClusterClient != nil {
			// For cluster, we need to scan all masters
			err = RedisClusterClient.ForEachMaster(ctx, func(ctx context.Context, master *redis.Client) error {
				masterKeys, _, err := master.Scan(ctx, cursor, pattern, count).Result()
				if err != nil {
					return err
				}
				keys = append(keys, masterKeys...)
				return nil
			})
		} else {
			keys, cursor, err = RedisClient.Scan(ctx, cursor, pattern, count).Result()
		}

		if err != nil {
			return nil, err
		}

		allKeys = append(allKeys, keys...)

		if cursor == 0 {
			break
		}
	}

	return allKeys, nil
}

// FlushDatabase flushes the current database (development/testing only)
func FlushDatabase(ctx context.Context) error {
	client := GetRedisClient()
	if client == nil {
		return fmt.Errorf("Redis client not initialized")
	}

	// Only allow in development/test environments
	config := GetConfig()
	if config.IsProduction() {
		return fmt.Errorf("flush database not allowed in production")
	}

	if RedisClusterClient != nil {
		return RedisClusterClient.FlushDB(ctx).Err()
	}
	return RedisClient.FlushDB(ctx).Err()
}

// SetWithLock implements distributed locking with Redis
func SetWithLock(ctx context.Context, key, lockKey string, value interface{}, expiration, lockTimeout time.Duration) error {
	client := GetRedisClient()
	if client == nil {
		return fmt.Errorf("Redis client not initialized")
	}

	// Acquire lock
	lockValue := strconv.FormatInt(time.Now().UnixNano(), 10)
	acquired, err := client.SetNX(ctx, lockKey, lockValue, lockTimeout).Result()
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	if !acquired {
		return fmt.Errorf("failed to acquire lock: already locked")
	}

	// Defer lock release
	defer func() {
		// Use Lua script to safely release lock
		luaScript := `
			if redis.call("get", KEYS[1]) == ARGV[1] then
				return redis.call("del", KEYS[1])
			else
				return 0
			end
		`
		client.Eval(ctx, luaScript, []string{lockKey}, lockValue)
	}()

	// Set the actual value
	return Set(ctx, key, value, expiration)
}

// Batch operations

// BatchSet performs batch set operations
func BatchSet(ctx context.Context, operations map[string]interface{}, expiration time.Duration) error {
	client := GetRedisClient()
	if client == nil {
		return fmt.Errorf("Redis client not initialized")
	}

	pipe := client.Pipeline()

	for key, value := range operations {
		var data []byte
		var err error

		switch v := value.(type) {
		case string:
			data = []byte(v)
		case []byte:
			data = v
		default:
			data, err = json.Marshal(value)
			if err != nil {
				return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
			}
		}

		pipe.Set(ctx, key, data, expiration)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// BatchDelete performs batch delete operations
func BatchDelete(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	client := GetRedisClient()
	if client == nil {
		return fmt.Errorf("Redis client not initialized")
	}

	return client.Del(ctx, keys...).Err()
}

// Transaction support

// Transaction executes commands in a Redis transaction
func Transaction(ctx context.Context, fn func(*redis.Tx) error, keys ...string) error {
	client := GetRedisClient()
	if client == nil {
		return fmt.Errorf("Redis client not initialized")
	}

	// Redis cluster doesn't support transactions across different slots
	if RedisClusterClient != nil {
		return fmt.Errorf("transactions not supported in cluster mode")
	}

	return RedisClient.Watch(ctx, fn, keys...)
}
