package cache

import (
	"time"

	"github.com/jellydator/ttlcache/v3"
)

type MembershipCache interface {
	Get(key string) bool
	Set(key string, value bool)
}

type MembershipTTL struct {
	cache *ttlcache.Cache[string, bool]
}

func NewMembershipTTL(ttl time.Duration) *MembershipTTL {
	opts := []ttlcache.Option[string, bool]{}
	if ttl > 0 {
		opts = append(opts, ttlcache.WithTTL[string, bool](ttl))
	} else {
		opts = append(opts, ttlcache.WithTTL[string, bool](DefaultTTL))
	}
	return &MembershipTTL{
		cache: ttlcache.New(opts...),
	}
}

func (c *MembershipTTL) Get(key string) bool {
	if item := c.cache.Get(key); item != nil {
		return item.Value()
	}
	return false
}

func (c *MembershipTTL) Set(key string, value bool) {
	c.cache.Set(key, value, ttlcache.DefaultTTL)
}

func (c *MembershipTTL) Start() {
	c.cache.Start()
}

func (c *MembershipTTL) Stop() {
	c.cache.Stop()
}
