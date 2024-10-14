package cache

import "time"

type data struct {
	value    []byte
	expireAt *time.Time
}

type Cache struct {
	data map[string]*data
}

func New() *Cache {
	c := &Cache{
		data: make(map[string]*data),
	}

	go func() {
		ticker := time.NewTicker(time.Hour)
		for t := range ticker.C {
			for key, data := range c.data {
				if data.expireAt == nil {
					continue
				}
				if t.After(*data.expireAt) {
					delete(c.data, key)
				}
			}
		}
	}()

	return c
}

func (c *Cache) Get(key string) ([]byte, bool) {
	data, exists := c.data[key]
	if data.expireAt != nil && data.expireAt.Before(time.Now()) {
		return nil, false
	}
	return data.value, exists
}

func (c *Cache) Put(key string, value []byte, expiration time.Duration) {
	data := &data{value: value}
	if expiration != 0 {
		t := time.Now().Add(expiration)
		data.expireAt = &t
	}
	c.data[key] = data
}

func (c *Cache) Delete(key string) {
	delete(c.data, key)
}
