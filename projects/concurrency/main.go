package main

import (
	"container/list"
	"sync"
)

type Cache[K comparable, V any] struct {
	// the maximum number of entries in the cache
	capacity int

	mu sync.Mutex
	// entries stores the key and value (a pointer to a list element)
	entries map[K]*list.Element
	// a list used for eviction policy
	list *list.List
}

type CacheEntry[K comparable, V any] struct {
	key   K
	value V
}

func NewCache[K comparable, V any](capacity int) Cache[K, V] {
	return Cache[K, V]{
		capacity: capacity,
		entries:  make(map[K]*list.Element, capacity),
		list:     list.New(),
	}
}

func (c *Cache[K, V]) Put(key K, value V) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	// check if there is a key in the `entries` map to update
	listElement, ok := c.entries[key]

	// if there does exist:
	if ok {
		// update the list element's value
		listElement.Value = &CacheEntry[K, V]{key, value}
		// move the list element to the front of the list
		c.list.MoveToFront(listElement)

		return true
	}

	// if the capacity is reached:
	if len(c.entries) == c.capacity {
		// get the last `list` element
		lastListElement := c.list.Back()
		// get it's value (a struct with key and value)
		lastListElementValue := lastListElement.Value.(*CacheEntry[K, V])
		// delete that key from the `entries` map
		delete(c.entries, lastListElementValue.key)
		// delete that list element from the `list`
		c.list.Remove(lastListElement)
	}

	// if there is no key:

	// create a new list element and insert it at the front of the `list`
	listElement = c.list.PushFront(&CacheEntry[K, V]{key, value})
	// make a new key/value pair in the `entries` map whose value points to the new list element
	c.entries[key] = listElement

	return false
}

func (c *Cache[K, V]) Get(key K) (*V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// check if there is a key in the `entries` map to get the value of
	listElement, ok := c.entries[key]

	// if the key doesn't exist:
	if !ok {
		// var zeroValue V
		// return &zeroValue, false
		return nil, false
	}

	// if key does exist:

	// move the list.Element the front of the `list`
	c.list.MoveToFront(listElement)

	// get the value from the list element
	cacheValue := listElement.Value.(*CacheEntry[K, V])

	return &cacheValue.value, true
}
