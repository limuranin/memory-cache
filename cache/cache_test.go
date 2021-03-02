package cache

import (
	"context"
	"testing"
	"time"

	"memory-cache/config"

	"github.com/stretchr/testify/suite"
)

type CacheSuite struct {
	suite.Suite

	ctx    context.Context
	cancel context.CancelFunc
	cfg    *config.CacheCfg
	cache  *Cache

	key         string
	ttl         time.Duration
	stringValue string
	sliceValue  []interface{}
	mapValue    map[string]interface{}
}

func (s *CacheSuite) SetupSuite() {
	s.cfg = &config.CacheCfg{
		CleaningInterval: 1 * time.Hour,
	}

	s.key = "key"
	s.ttl = 1 * time.Hour
	s.stringValue = "value"
	s.sliceValue = []interface{}{"one", "two", "three"}
	s.mapValue = map[string]interface{}{
		"one":   "red",
		"two":   "green",
		"three": "black",
	}
}

func (s *CacheSuite) SetupTest() {
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.cache = NewCache(s.ctx, s.cfg)
}

func (s *CacheSuite) TearDownTest() {
	s.cancel()
}

func (s *CacheSuite) TestElementNotFound() {
	s.cache.Start()

	cacheValue, err := s.cache.Get(s.key)
	s.Require().Nil(cacheValue)
	s.Require().EqualError(err, ErrElementNotFound.Error())
}

func (s *CacheSuite) TestInvalidValueType() {
	s.cache.Start()

	value := 10
	err := s.cache.Set(s.key, value, s.ttl)
	s.Require().EqualError(err, ErrInvalidValueType.Error())
}

func (s *CacheSuite) TestStringValue() {
	s.cache.Start()
	s.Require().NoError(s.cache.Set(s.key, s.stringValue, s.ttl))

	cacheValue, err := s.cache.Get(s.key)
	s.Require().NoError(err)
	s.Require().Equal(s.stringValue, cacheValue)

	keys, err := s.cache.Keys()
	s.Require().NoError(err)
	s.Require().Len(keys, 1)
	s.Require().Contains(keys, s.key)

	err = s.cache.Remove(s.key)
	s.Require().NoError(err)

	keys, err = s.cache.Keys()
	s.Require().NoError(err)
	s.Require().Empty(keys)
}

func (s *CacheSuite) TestUpdateStringValue() {
	s.cache.Start()
	s.Require().NoError(s.cache.Set(s.key, s.stringValue, s.ttl))

	cacheValue, err := s.cache.Get(s.key)
	s.Require().NoError(err)
	s.Require().Equal(s.stringValue, cacheValue)

	newValue := "newValue"
	s.Require().NoError(s.cache.Set(s.key, newValue, s.ttl))

	cacheValue, err = s.cache.Get(s.key)
	s.Require().NoError(err)
	s.Require().Equal(newValue, cacheValue)
}

func (s *CacheSuite) TestUpdateValueType() {
	s.cache.Start()
	s.Require().NoError(s.cache.Set(s.key, s.stringValue, s.ttl))

	cacheValue, err := s.cache.Get(s.key)
	s.Require().NoError(err)
	s.Require().Equal(s.stringValue, cacheValue)

	s.Require().NoError(s.cache.Set(s.key, s.sliceValue, s.ttl))

	cacheValue, err = s.cache.Get(s.key)
	s.Require().NoError(err)
	s.Require().ElementsMatch(s.sliceValue, cacheValue)
}

func (s *CacheSuite) TestSliceValue() {
	s.cache.Start()
	s.Require().NoError(s.cache.Set(s.key, s.sliceValue, s.ttl))

	cacheValue, err := s.cache.Get(s.key)
	s.Require().NoError(err)
	s.Require().ElementsMatch(s.sliceValue, cacheValue)

	keys, err := s.cache.Keys()
	s.Require().NoError(err)
	s.Require().Len(keys, 1)
	s.Require().Contains(keys, s.key)

	err = s.cache.Remove(s.key)
	s.Require().NoError(err)

	keys, err = s.cache.Keys()
	s.Require().NoError(err)
	s.Require().Empty(keys)
}

func (s *CacheSuite) TestGetSliceElementByIndex() {
	s.cache.Start()
	s.Require().NoError(s.cache.Set(s.key, s.sliceValue, s.ttl))

	for i, v := range s.sliceValue {
		elemVal, err := s.cache.GetListElem(s.key, i)
		s.Require().NoError(err)
		s.Require().Equal(v, elemVal)
	}
}

func (s *CacheSuite) TestSliceIndexOutOfRange() {
	s.cache.Start()
	s.Require().NoError(s.cache.Set(s.key, s.sliceValue, s.ttl))

	val, err := s.cache.GetListElem(s.key, -1)
	s.Require().Nil(val)
	s.Require().EqualError(err, ErrIndexOutOfRange.Error())

	val, err = s.cache.GetListElem(s.key, 3)
	s.Require().Nil(val)
	s.Require().EqualError(err, ErrIndexOutOfRange.Error())
}

func (s *CacheSuite) TestExpectSliceValue() {
	s.cache.Start()
	s.Require().NoError(s.cache.Set(s.key, s.stringValue, s.ttl))

	sliceValue, err := s.cache.GetListElem(s.key, 1)
	s.Require().EqualError(err, ErrNotSliceValue.Error())
	s.Require().Nil(sliceValue)
}

func (s *CacheSuite) TestMapValue() {
	s.cache.Start()
	s.Require().NoError(s.cache.Set(s.key, s.mapValue, s.ttl))

	cacheValue, err := s.cache.Get(s.key)
	s.Require().NoError(err)

	cacheMap, ok := cacheValue.(map[string]interface{})
	s.Require().True(ok)
	s.Require().Equal(len(s.mapValue), len(cacheMap))
	for k, v := range s.mapValue {
		sv, ok := cacheMap[k]
		s.Require().True(ok)
		s.Require().Equal(v, sv)
	}

	keys, err := s.cache.Keys()
	s.Require().NoError(err)
	s.Require().Len(keys, 1)
	s.Require().Contains(keys, s.key)

	err = s.cache.Remove(s.key)
	s.Require().NoError(err)

	keys, err = s.cache.Keys()
	s.Require().NoError(err)
	s.Require().Empty(keys)
}

func (s *CacheSuite) TestMapElementKeyValue() {
	s.cache.Start()
	s.Require().NoError(s.cache.Set(s.key, s.mapValue, s.ttl))

	for k, v := range s.mapValue {
		elemKeyValue, err := s.cache.GetMapElemValue(s.key, k)
		s.Require().NoError(err)
		s.Require().Equal(v, elemKeyValue)
	}
}

func (s *CacheSuite) TestMapElementNotFound() {
	s.cache.Start()
	s.Require().NoError(s.cache.Set(s.key, s.mapValue, s.ttl))

	elemKeyValue, err := s.cache.GetMapElemValue(s.key, "someKey")
	s.Require().EqualError(err, ErrMapElementNotFound.Error())
	s.Require().Nil(elemKeyValue)
}

func (s *CacheSuite) TestExpectMapValue() {
	s.cache.Start()
	s.Require().NoError(s.cache.Set(s.key, s.stringValue, s.ttl))

	elemKeyValue, err := s.cache.GetMapElemValue(s.key, "someKey")
	s.Require().EqualError(err, ErrNotMapValue.Error())
	s.Require().Nil(elemKeyValue)
}

func (s *CacheSuite) TestCleanByTtl() {
	cleaningInterval := 100 * time.Millisecond
	ttl := 50 * time.Millisecond

	s.cache.cfg.CleaningInterval = cleaningInterval
	s.cache.Start()

	s.Require().NoError(s.cache.Set(s.key, s.stringValue, ttl))

	cacheValue, err := s.cache.Get(s.key)
	s.Require().NoError(err)
	s.Require().Equal(s.stringValue, cacheValue)

	<-time.After(cleaningInterval + 10*time.Millisecond)
	cacheValue, err = s.cache.Get(s.key)
	s.Require().Nil(cacheValue)
	s.Require().EqualError(err, ErrElementNotFound.Error())
}

func (s *CacheSuite) TestGetExpiredElement() {
	cleaningInterval := 100 * time.Millisecond
	ttl := 50 * time.Millisecond

	s.cache.cfg.CleaningInterval = cleaningInterval
	s.cache.Start()

	s.Require().NoError(s.cache.Set(s.key, s.stringValue, ttl))

	cacheValue, err := s.cache.Get(s.key)
	s.Require().NoError(err)
	s.Require().Equal(s.stringValue, cacheValue)

	<-time.After(cleaningInterval - 10*time.Millisecond)
	cacheValue, err = s.cache.Get(s.key)
	s.Require().Nil(cacheValue)
	s.Require().EqualError(err, ErrElementExpired.Error())
}

func (s *CacheSuite) TestZeroTtl() {
	s.cache.Start()

	s.Require().NoError(s.cache.Set(s.key, s.stringValue, 0))

	cacheValue, err := s.cache.Get(s.key)
	cacheValue, err = s.cache.Get(s.key)
	s.Require().Nil(cacheValue)
	s.Require().EqualError(err, ErrElementExpired.Error())
}

func (s *CacheSuite) TestUpdateTtl() {
	cleaningInterval := 100 * time.Millisecond
	s.cache.cfg.CleaningInterval = cleaningInterval
	s.cache.Start()

	ttl := 50 * time.Millisecond
	s.Require().NoError(s.cache.Set(s.key, s.stringValue, ttl))
	cacheValue, err := s.cache.Get(s.key)
	s.Require().NoError(err)
	s.Require().Equal(s.stringValue, cacheValue)

	newTtl := 150 * time.Millisecond
	s.Require().NoError(s.cache.Set(s.key, s.stringValue, newTtl))

	<-time.After(cleaningInterval + 10*time.Millisecond)
	cacheValue, err = s.cache.Get(s.key)
	s.Require().NoError(err)
	s.Require().Equal(s.stringValue, cacheValue)

	<-time.After(cleaningInterval + 10*time.Millisecond)
	cacheValue, err = s.cache.Get(s.key)
	s.Require().Nil(cacheValue)
	s.Require().EqualError(err, ErrElementNotFound.Error())
}

func TestCache(t *testing.T) {
	suite.Run(t, new(CacheSuite))
}
