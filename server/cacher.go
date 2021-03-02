package server

import "time"

type Cacher interface {
	Set(key string, value interface{}, ttl time.Duration) error
	Get(key string) (interface{}, error)
	GetListElem(key string, index int) (interface{}, error)
	GetMapElemValue(key string, mapKey string) (interface{}, error)
	Remove(key string) error
	Keys() ([]string, error)
}
