package cache

import (
	"strconv"

	"github.com/pkg/errors"

	"go.uber.org/zap"
	"max.ks1230/project-base/internal/logger"

	"github.com/bradfitz/gomemcache/memcache"
)

var defaultBase = 10

type MemcacheClient struct {
	client *memcache.Client
}

type config interface {
	Hosts() []string
}

func NewMemcache(config config) (*MemcacheClient, error) {
	logger.Info("memcached hosts", zap.Strings("hosts", config.Hosts()))
	mc := memcache.New(config.Hosts()...)
	return &MemcacheClient{mc}, mc.Ping()
}

func formatKey(userID int64, option string) string {
	return strconv.FormatInt(userID, defaultBase) + ":" + option
}

func (mc *MemcacheClient) CacheReport(userID int64, option string, report string) error {
	logger.Info("cache report", zap.Int64("userID", userID), zap.String("option", option))
	return mc.client.Set(&memcache.Item{
		Key:   formatKey(userID, option),
		Value: []byte(report)},
	)
}

func (mc *MemcacheClient) GetReport(userID int64, option string) (string, error) {
	logger.Info("get report from cache", zap.Int64("userID", userID), zap.String("option", option))
	item, err := mc.client.Get(formatKey(userID, option))
	if err != nil {
		return "", err
	}
	return string(item.Value), nil
}

func (mc *MemcacheClient) InvalidateCache(userID int64, options []string) error {
	logger.Info("invalidate cache", zap.Int64("userID", userID))

	for _, opt := range options {
		err := mc.client.Delete(formatKey(userID, opt))
		if err != nil && !errors.Is(err, memcache.ErrCacheMiss) {
			return err
		}
	}
	return nil
}
