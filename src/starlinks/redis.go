package starlinks

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis"
)

const (
	DEFAULT_KEY_BASE = "starlink_cache"

	// map_set, id, count_set
	QUERY_SCRIPT = `
        local result = redis.call('hget', KEYS[1], KEYS[2])
        if result ~= nil then
            redis.call('zincrby', KEYS[3], 1, KEYS[2])
        else
            result = ''
        end
        return result
    `

	QUERY_BATCH_SCRIPT = `
        local results={}
        for i=3,#KEYS-1,1 do
            results[i] = redis.call('hget', KEYS[1], KEYS[i])
            if results[i] != nil then
                redis.call('zincrby', KEYS[2], 1, KEYS[i])
            end
        end
        return results
    `

	CLEAN_CACHE_SCRIPT = `
        local sub = ARGS[1] - redis.call('zcard', KEYS[1])
        if sub < 0 then
            local cleaned = redis.call('zpopmin', KEYS[1], -sub)
            local keys={'hdel', KEYS[2]}
            for i=1, #cleaned, 2 do 
                keys[i + 2] = clean[i]
            end
            redis.call(unpack(keys))
            return -sub
        end
        return 0
    `

	// map_set, count_set, id, link
	ADD_LINK_SCRIPT = `
        redis.call('zadd', KEYS[2], 0, KEYS[3])
        redis.call('hset', KEYS[1], KEYS[3], KEYS[4])
        return 1
    `

	// map_set, count_set, id1, link1, id2, link2, ....
	ADD_LINKS_SCRIPT = `
        redis.call('hmset', KEYS[1], unpack(KEYS, 3))
        local keys={}
        for i = 3, #KEYS, 2 do 
            keys[i - 1] = KEYS[i]
            keys[i - 2] = 0
        end
        redis.call('zadd', KEYS[2], unpack(keys))
        return #KEYS
    `

	// map_set, count_set, id1, id2, ...
	REMOVE_LINKS_SCRIPT = `
        redis.call('hdel', KEYS[1], unpack(keys, 3))
        redis.call('zdel', KEYS[2], unpack(keys, 3))
    `
)

type RedisLinkCache struct {
	client       *redis.Client
	max_instance uint
	key_base     string
}

func NewRedisLinkCache(dsn string) (CacheStorage, error) {
	domain, path, _ := parseNetPath(dsn)
	if domain != "tcp" {
		return nil, errors.New("Redis should be connected via TCP.")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     path,
		DB:       0,
		Password: "",
	})

	return &RedisLinkCache{
		client:       client,
		max_instance: 2000,
		key_base:     DEFAULT_KEY_BASE,
	}, nil
}

func (cache *RedisLinkCache) SetMaxCacheInstance(max uint) error {
	cache.max_instance = max
	return nil
}

//func (cache *RedisLinkCache) SetCacheKeyBase(base string) error {
//}

func (cache *RedisLinkCache) QueryLink(id LinkID) (string, error) {
	cmd := cache.client.Eval(QUERY_SCRIPT, []string{cache.key_base + "map", id.ToString(), cache.key_base + "cnt"})
	result, err := cmd.Result()
	if err != nil {
		return "", err
	}
	link, ok := result.(string)
	if !ok {
		return "", errors.New("Cached link is not string.")
	}
	return link, nil
}

func (cache *RedisLinkCache) QueryLinks(ids []LinkID) ([]string, error) {
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = fmt.Sprintf("%v", id)
	}
	cmd := cache.client.Eval(QUERY_BATCH_SCRIPT, append([]string{cache.key_base + "map", cache.key_base + "cnt"}, keys...))
	result, err := cmd.Result()
	if err != nil {
		return nil, err
	}
	links, ok := result.([]string)
	if !ok {
		return nil, errors.New("Redis return unexpected types.")
	}
	return links, nil
}

func (cache *RedisLinkCache) CacheClean() (int, error) {
	cmd := cache.client.Eval(CLEAN_CACHE_SCRIPT, []string{cache.key_base + "cnt", cache.key_base + "map"}, cache.max_instance)
	result, err := cmd.Result()
	if err != nil {
		return 0, err
	}
	cnt, ok := result.(int)
	if !ok {
		return 0, errors.New("Redis return unexpected types.")
	}
	return cnt, nil
}

func (cache *RedisLinkCache) AddLink(id LinkID, url string) error {
	cmd := cache.client.Eval(ADD_LINK_SCRIPT, []string{cache.key_base + "map", cache.key_base + "cnt", id.ToString(), url})
	_, err := cmd.Result()
	return err
}

func (cache *RedisLinkCache) AddLinks(url_map map[LinkID]string) error {
	args := make([]string, len(url_map))
	i := 0
	for id, link := range url_map {
		args[i] = fmt.Sprintf("%v", id)
		args[i+1] = link
		i += 2
	}
	cmd := cache.client.Eval(ADD_LINKS_SCRIPT, append([]string{cache.key_base + "map", cache.key_base + "cnt"}, args...))
	_, err := cmd.Result()
	return err
}

func (cache *RedisLinkCache) RemoveLink(id LinkID) error {
	return cache.RemoveLinks([]LinkID{id})
}

func (cache *RedisLinkCache) RemoveLinks(ids []LinkID) error {
	args := make([]string, len(ids))
	for i, id := range ids {
		args[i] = fmt.Sprintf("%v", id)
	}
	cmd := cache.client.Eval(REMOVE_LINKS_SCRIPT, append([]string{cache.key_base + "map", cache.key_base + "cnt"}, args...))
	_, err := cmd.Result()
	return err
}
