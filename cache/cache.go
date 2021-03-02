package cache

import (
	"context"
	"errors"
	"sync"
	"time"

	"memory-cache/config"
)

var (
	ErrInvalidValueType   = errors.New("invalid value type")
	ErrElementNotFound    = errors.New("element not found in cache")
	ErrElementExpired     = errors.New("element has already expired and will be removed on cache cleaning")
	ErrNotSliceValue      = errors.New("value is not a slice")
	ErrNotMapValue        = errors.New("value is not a map")
	ErrIndexOutOfRange    = errors.New("slice index out of range")
	ErrMapElementNotFound = errors.New("element not found in map")
)

type item struct {
	value          interface{}
	expirationTime time.Time
}

type Cache struct {
	cfg *config.CacheCfg
	ctx context.Context
	sync.RWMutex
	data map[string]*item
}

func NewCache(ctx context.Context, cfg *config.CacheCfg) *Cache {
	return &Cache{
		cfg:     cfg,
		ctx:     ctx,
		RWMutex: sync.RWMutex{},
		data:    make(map[string]*item),
	}
}

func (c *Cache) Start() {
	go func() {
		ticker := time.NewTicker(c.cfg.CleaningInterval)
		defer ticker.Stop()

		done := c.ctx.Done()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				c.deleteExpired()
			}
		}
	}()
}

func (c *Cache) deleteExpired() {
	c.Lock()
	defer c.Unlock()

	for key, item := range c.data {
		if item.expirationTime.Before(time.Now()) {
			delete(c.data, key)
		}
	}
}

func (c *Cache) Set(key string, value interface{}, ttl time.Duration) error {
	if err := checkValueType(value); err != nil {
		return err
	}

	item := &item{
		value:          value,
		expirationTime: time.Now().Add(ttl),
	}

	c.Lock()
	defer c.Unlock()

	c.data[key] = item
	return nil
}

func (c *Cache) Get(key string) (interface{}, error) {
	c.RLock()
	defer c.RUnlock()

	return c.unsafeGet(key)
}

func (c *Cache) GetListElem(key string, index int) (interface{}, error) {
	c.RLock()
	defer c.RUnlock()

	itemValue, err := c.unsafeGet(key)
	if err != nil {
		return nil, err
	}

	itemValueAsSlice, ok := itemValue.([]interface{})
	if !ok {
		return nil, ErrNotSliceValue
	}

	if (index < 0) || (len(itemValueAsSlice) <= index) {
		return nil, ErrIndexOutOfRange
	}

	return itemValueAsSlice[index], nil
}

func (c *Cache) GetMapElemValue(key string, mapKey string) (interface{}, error) {
	c.RLock()
	defer c.RUnlock()

	itemValue, err := c.unsafeGet(key)
	if err != nil {
		return nil, err
	}

	itemValueAsMap, ok := itemValue.(map[string]interface{})
	if !ok {
		return nil, ErrNotMapValue
	}

	mapKeyVal, ok := itemValueAsMap[mapKey]
	if !ok {
		return nil, ErrMapElementNotFound
	}

	return mapKeyVal, nil
}

func (c *Cache) Remove(key string) error {
	c.Lock()
	defer c.Unlock()

	delete(c.data, key)

	return nil
}

func (c *Cache) Keys() ([]string, error) {
	c.RLock()
	defer c.RUnlock()

	keys := make([]string, 0, len(c.data))
	for k := range c.data {
		keys = append(keys, k)
	}

	return keys, nil
}

func (c *Cache) unsafeGet(key string) (interface{}, error) {
	item, ok := c.data[key]
	if !ok {
		return nil, ErrElementNotFound
	}

	if item.expirationTime.Before(time.Now()) {
		return nil, ErrElementExpired
	}

	return item.value, nil
}

func checkValueType(value interface{}) error {
	switch value.(type) {
	case string:
		return nil
	case []interface{}:
		return nil
	case map[string]interface{}:
		return nil
	default:
		return ErrInvalidValueType
	}
}
