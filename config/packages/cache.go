package packages

import (
	"encoding/json"
	"github.com/patrickmn/go-cache"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
)

var Cache *cache.Cache
var cacheMutex sync.Mutex

const (
	CACHE_EXPIRATION       = 24 * time.Hour
	CLEAN_UP_INTERVEL_TIME = 1 * time.Hour
	FILE                   = "cache.json"
	FILE_CACHE_KEY         = "isFileCache"
	FILE_CACHE_EXPIRE_TIME = 6 * time.Hour
)

type CachedData struct {
	Data interface{}
}

type CacheData struct {
	Items map[string]interface{} `json:"items"`
}

func CacheDeclare(storeCache bool) *cache.Cache {

	if Cache != nil {
		if _, found := Cache.Get(FILE_CACHE_KEY); !found && !storeCache {
			_ = LoadFromFileAndCache()
		}
		return Cache
	}

	if Cache == nil {
		Cache = cache.New(CACHE_EXPIRATION, CLEAN_UP_INTERVEL_TIME)
	}
	if _, found := Cache.Get(FILE_CACHE_KEY); !found && !storeCache {
		_ = LoadFromFileAndCache()
	}
	return Cache
}

func SaveCacheToFile() error {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	cacheItems := Cache.Items()
	cacheData := CacheData{
		Items: make(map[string]interface{}, len(cacheItems)),
	}

	for key, item := range cacheItems {
		cacheData.Items[key] = item.Object
	}

	cacheJSON, err := json.Marshal(cacheData)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(FILE, cacheJSON, 0644)
	return err
}

func LoadFromFileAndCache() error {

	if _, err := os.Stat(FILE); err == nil {
		cacheData, err := ioutil.ReadFile(FILE)
		if err == nil {

			var cacheItems CacheData
			if err := json.Unmarshal(cacheData, &cacheItems); err != nil {
				log.Println("Error loading cache data:", err)
			}
			for key, item := range cacheItems.Items {
				Cache.Set(key, item, cache.DefaultExpiration)
			}
			Cache.Set(FILE_CACHE_KEY, true, FILE_CACHE_EXPIRE_TIME)
		}
	}
	return nil
}

func RemoveCacheFile() error {
	if _, fileErr := os.Stat(FILE); fileErr == nil {
		if err := os.Remove(FILE); err != nil {
			return err
		}
	}
	return nil
}
