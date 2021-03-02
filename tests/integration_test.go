package tests

import (
	"context"
	"testing"
	"time"

	"memory-cache/cache"
	"memory-cache/client"
	"memory-cache/config"
	"memory-cache/logger"
	"memory-cache/server"

	"github.com/stretchr/testify/suite"
)

const ShutdownServerTimeout = 10 * time.Second

type IntegrationSuite struct {
	suite.Suite

	srv          *server.Server
	cacheStorage *cache.Cache
	cacheCancel  context.CancelFunc

	cacher server.Cacher

	key         string
	ttl         time.Duration
	stringValue string
	sliceValue  []interface{}
	mapValue    map[string]interface{}
}

func (s *IntegrationSuite) SetupSuite() {
	cfg, err := config.Init()
	s.Require().NoError(err)

	s.Require().NoError(logger.Init())

	logger.Info("Start suite setup")

	logger.Infof("Start cache with cleaning interval: %v", cfg.Cache.CleaningInterval)
	var cacheCtx context.Context
	cacheCtx, s.cacheCancel = context.WithCancel(context.Background())
	cacheStorage := cache.NewCache(cacheCtx, cfg.Cache)
	cacheStorage.Start()

	logger.Infof("Start server listen address: %v", cfg.Server.ListenAddress)
	s.srv = server.NewServer(cfg.Server, cacheStorage)
	s.Require().NoError(s.srv.Start())

	s.cacher = client.NewClient(cfg.Server.ListenAddress)

	s.key = "key"
	s.ttl = 1 * time.Hour
	s.stringValue = "value"
	s.sliceValue = []interface{}{"one", "two", "three"}
	s.mapValue = map[string]interface{}{
		"one":   "red",
		"two":   "green",
		"three": "black",
	}

	logger.Info("Finish suite setup")
}

func (s *IntegrationSuite) TearDownSuite() {
	logger.Info("Start suite tear down")

	logger.Info("Stopping cache")
	s.cacheCancel()

	logger.Infof("Shutting down the server, wait gracefully shutdown for %v", ShutdownServerTimeout)
	shutdownServerCtx, shutdownServerCancelFunc := context.WithTimeout(context.Background(), ShutdownServerTimeout)
	defer shutdownServerCancelFunc()

	s.Require().NoError(s.srv.Shutdown(shutdownServerCtx))
	logger.Info("Server shutdown gracefully")

	logger.Info("Finish suite tear down")
}

func (s *IntegrationSuite) TestStringValue() {
	s.Require().NoError(s.cacher.Set(s.key, s.stringValue, s.ttl))

	cacheValue, err := s.cacher.Get(s.key)
	s.Require().NoError(err)
	s.Require().Equal(s.stringValue, cacheValue)

	keys, err := s.cacher.Keys()
	s.Require().NoError(err)
	s.Require().Len(keys, 1)
	s.Require().Contains(keys, s.key)

	err = s.cacher.Remove(s.key)
	s.Require().NoError(err)

	keys, err = s.cacher.Keys()
	s.Require().NoError(err)
	s.Require().Empty(keys)
}

func (s *IntegrationSuite) TestSliceValue() {
	s.Require().NoError(s.cacher.Set(s.key, s.sliceValue, s.ttl))

	cacheValue, err := s.cacher.Get(s.key)
	s.Require().NoError(err)
	s.Require().ElementsMatch(s.sliceValue, cacheValue)

	keys, err := s.cacher.Keys()
	s.Require().NoError(err)
	s.Require().Len(keys, 1)
	s.Require().Contains(keys, s.key)

	err = s.cacher.Remove(s.key)
	s.Require().NoError(err)

	keys, err = s.cacher.Keys()
	s.Require().NoError(err)
	s.Require().Empty(keys)
}

func (s *IntegrationSuite) TestMapValue() {
	s.Require().NoError(s.cacher.Set(s.key, s.mapValue, s.ttl))

	cacheValue, err := s.cacher.Get(s.key)
	s.Require().NoError(err)

	cacheMap, ok := cacheValue.(map[string]interface{})
	s.Require().True(ok)
	s.Require().Equal(len(s.mapValue), len(cacheMap))
	for k, v := range s.mapValue {
		sv, ok := cacheMap[k]
		s.Require().True(ok)
		s.Require().Equal(v, sv)
	}

	keys, err := s.cacher.Keys()
	s.Require().NoError(err)
	s.Require().Len(keys, 1)
	s.Require().Contains(keys, s.key)

	err = s.cacher.Remove(s.key)
	s.Require().NoError(err)

	keys, err = s.cacher.Keys()
	s.Require().NoError(err)
	s.Require().Empty(keys)
}

func (s *IntegrationSuite) TestGetSliceElementByIndex() {
	s.Require().NoError(s.cacher.Set(s.key, s.sliceValue, s.ttl))

	for i, v := range s.sliceValue {
		elemVal, err := s.cacher.GetListElem(s.key, i)
		s.Require().NoError(err)
		s.Require().Equal(v, elemVal)
	}
}

func (s *IntegrationSuite) TestMapElementKeyValue() {
	s.Require().NoError(s.cacher.Set(s.key, s.mapValue, s.ttl))

	for k, v := range s.mapValue {
		elemKeyValue, err := s.cacher.GetMapElemValue(s.key, k)
		s.Require().NoError(err)
		s.Require().Equal(v, elemKeyValue)
	}
}

func (s *IntegrationSuite) TestElementNotFound() {
	cacheValue, err := s.cacher.Get(s.key)
	s.Require().Nil(cacheValue)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), cache.ErrElementNotFound.Error())
}

func (s *IntegrationSuite) TestSliceIndexOutOfRange() {
	s.Require().NoError(s.cacher.Set(s.key, s.sliceValue, s.ttl))

	val, err := s.cacher.GetListElem(s.key, 3)
	s.Require().Nil(val)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), cache.ErrIndexOutOfRange.Error())
}

func (s *IntegrationSuite) TestMapElementNotFound() {
	s.Require().NoError(s.cacher.Set(s.key, s.mapValue, s.ttl))

	elemKeyValue, err := s.cacher.GetMapElemValue(s.key, "someKey")
	s.Require().Nil(elemKeyValue)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), cache.ErrMapElementNotFound.Error())
}

func (s *IntegrationSuite) TestExpectSliceValue() {
	s.Require().NoError(s.cacher.Set(s.key, s.stringValue, s.ttl))

	sliceValue, err := s.cacher.GetListElem(s.key, 1)
	s.Require().Nil(sliceValue)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), cache.ErrNotSliceValue.Error())
}

func (s *IntegrationSuite) TestExpectMapValue() {
	s.Require().NoError(s.cacher.Set(s.key, s.stringValue, s.ttl))

	elemKeyValue, err := s.cacher.GetMapElemValue(s.key, "someKey")
	s.Require().Nil(elemKeyValue)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), cache.ErrNotMapValue.Error())
}

func TestIntegration(t *testing.T) {
	suite.Run(t, new(IntegrationSuite))
}
