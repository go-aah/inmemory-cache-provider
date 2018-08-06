// Copyright (c) Jeevanandam M. (https://github.com/jeevatkm)
// aahframework.org/cache/inmemory source code and usage is governed by a MIT style
// license that can be found in the LICENSE file.

package inmemory

import (
	"fmt"
	"testing"
	"time"

	"aahframework.org/essentials.v0"

	"aahframework.org/aah.v0/cache"
	"aahframework.org/test.v0/assert"
)

func TestInmemoryCacheAddAndGet(t *testing.T) {
	c := createTestCache(t, "addgetcache")
	for i := 0; i < 20; i++ {
		c.Put(fmt.Sprintf("key_%v", i), i, 3*time.Second)
	}

	for i := 5; i < 10; i++ {
		v := c.Get(fmt.Sprintf("key_%v", i))
		assert.Equal(t, i, v)
	}
	assert.Equal(t, "addgetcache", c.Name())
}

func TestInmemoryCacheAddAndDelete(t *testing.T) {
	c := createTestCache(t, "adddeletecache")
	for i := 0; i < 20; i++ {
		key := fmt.Sprintf("key_%v", i)
		c.Put(key, i, 3*time.Second)
		c.Delete(key)
	}
	assert.Equal(t, "adddeletecache", c.Name())
}

func TestInmemoryCacheGetOrAddAndExists(t *testing.T) {
	c := createTestCache(t, "addexistscache")
	for i := 0; i < 20; i++ {
		key := fmt.Sprintf("key_%v", i)
		c.GetOrPut(key, i, 3*time.Second)
		assert.True(t, c.Exists(key))
		c.GetOrPut(key, i, 3*time.Second)
		c.Put(key, i, 3*time.Second)
	}

	for i := 20; i < 30; i++ {
		assert.False(t, c.Exists(fmt.Sprintf("key_%v", i)))
	}
	assert.Equal(t, "addexistscache", c.Name())
}

func TestInmemoryCacheConcurrency(t *testing.T) {
	c := createTestCache(t, "testconcurrency")

	for i := 0; i < 10; i++ {
		c.Put(fmt.Sprintf("key_%v", i), i, 3*time.Second)
	}

	for i := 5; i < 10; i++ {
		v := c.Get(fmt.Sprintf("key_%v", i))
		assert.Equal(t, i, v)
	}

	go func() {
		for i := 11; i < 20; i++ {
			key := fmt.Sprintf("key_%v", i)
			c.GetOrPut(key, i, 1*time.Second)
			assert.True(t, c.Exists(key))
		}
	}()

	go func() {
		for i := 5; i < 9; i++ {
			key := fmt.Sprintf("key_%v", i)
			c.Delete(key)
			assert.Nil(t, c.Get(key))
		}
	}()

	time.Sleep(5 * time.Second)
}

func TestInmemoryMultipleCache(t *testing.T) {
	mgr := createCacheMgr()

	names := []string{"testcache1", "testcache2", "testcache3"}
	for _, name := range names {
		err := mgr.CreateCache(&cache.Config{Name: name, ProviderName: "inmemory"})
		assert.FailNowOnError(t, err, "unable to create cache")

		c := mgr.Cache(name)
		assert.NotNil(t, c)
		assert.Equal(t, name, c.Name())

		for i := 0; i < 20; i++ {
			c.Put(fmt.Sprintf("key_%v", i), i, 3*time.Second)
		}

		for i := 5; i < 10; i++ {
			v := c.Get(fmt.Sprintf("key_%v", i))
			assert.Equal(t, i, v)
		}
		c.Flush()
	}

	assert.Equal(t, 3, len(mgr.CacheNames()))
	assert.True(t, ess.IsSliceContainsString(mgr.CacheNames(), "testcache2"))
}

func TestInmemorySlideEvictionMode(t *testing.T) {
	mgr := createCacheMgr()
	err := mgr.CreateCache(&cache.Config{
		Name:          "testslidemode",
		ProviderName:  "inmemory",
		EvictionMode:  cache.EvictionModeSlide,
		SweepInterval: 2 * time.Second,
	})
	assert.FailNowOnError(t, err, "unable to create cache")

	c := mgr.Cache("testslidemode")
	assert.NotNil(t, c)

	for i := 0; i < 10; i++ {
		c.Put(fmt.Sprintf("key_%v", i), i, 3*time.Second)
	}

	for i := 5; i < 10; i++ {
		v := c.Get(fmt.Sprintf("key_%v", i))
		assert.Equal(t, i, v)
	}

	go func() {
		for i := 11; i < 20; i++ {
			key := fmt.Sprintf("key_%v", i)
			c.GetOrPut(key, i, 1*time.Second)
			assert.True(t, c.Exists(key))
		}
		for {
			for i := 11; i < 20; i++ {
				c.Get(fmt.Sprintf("key_%v", i))
			}
		}
	}()

	go func() {
		for i := 5; i < 9; i++ {
			key := fmt.Sprintf("key_%v", i)
			c.Delete(key)
			assert.Nil(t, c.Get(key))
		}
	}()

	time.Sleep(5 * time.Second)

	for i := 11; i < 20; i++ {
		v := c.Get(fmt.Sprintf("key_%v", i))
		assert.Equal(t, i, v)
	}
}

func createCacheMgr() *cache.Manager {
	mgr := cache.NewManager()
	mgr.AddProvider("inmemory", new(Provider))
	mgr.InitProviders(nil, nil)
	return mgr
}

func createTestCache(t *testing.T, name string) cache.Cache {
	mgr := createCacheMgr()

	err := mgr.CreateCache(&cache.Config{Name: name, ProviderName: "inmemory", SweepInterval: 2 * time.Second})
	assert.FailNowOnError(t, err, "unable to create cache")

	c := mgr.Cache(name)
	assert.NotNil(t, c)
	return c
}

func BenchmarkInmemoryCacheGet(b *testing.B) {
	b.StopTimer()
	mgr := createCacheMgr()
	_ = mgr.CreateCache(&cache.Config{
		Name:          "testslidemode",
		ProviderName:  "inmemory",
		EvictionMode:  cache.EvictionModeSlide,
		SweepInterval: 2 * time.Second,
	})

	c := mgr.Cache("testslidemode")
	for i := 0; i < 100000; i++ {
		_ = c.Put(fmt.Sprintf("key_%v", i), i, 3*time.Second)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		c.Get("key_50000")
	}
}
