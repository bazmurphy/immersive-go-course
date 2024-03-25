package main

import (
	"container/list"
	"sync"
)

type Cache[K comparable, V any] struct {
	capacity int
	entries  map[K]*list.Element
	list     *list.List
	mu       sync.Mutex
}

type CacheValue[K comparable, V any] struct {
	// i don't like that i am duplicating the key here... but it is for easier lookup later
	key   K
	value V
}

func NewCache[K comparable, V any](entryLimit int) Cache[K, V] {
	return Cache[K, V]{
		capacity: entryLimit,
		entries:  make(map[K]*list.Element, entryLimit),
		list:     list.New(),
	}
}

func (c *Cache[K, V]) Put(key K, value V) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	// check if there is a key in the entries map to update
	listElement, ok := c.entries[key]

	if ok {
		// update the list element's value
		listElement.Value = &CacheValue[K, V]{key, value}
		// move the list element to the front of the list
		c.list.MoveToFront(listElement)
		return true
	}

	if len(c.entries) == c.capacity {
		// get the last list element
		lastListElement := c.list.Back()
		// get it's value (a struct with key and value)
		lastListElementValue := lastListElement.Value.(*CacheValue[K, V])
		// delete that key from the entries map
		delete(c.entries, lastListElementValue.key)
		// delete that list element from the list
		c.list.Remove(lastListElement)
	}

	// if we reach here there is no key in the entries map

	// create a new list element and put at the front of the list
	listElement = c.list.PushFront(&CacheValue[K, V]{key, value})

	// make a new key/value pair in the entries map whose value points to the new list element
	c.entries[key] = listElement

	return false
}

func (c *Cache[K, V]) Get(key K) (*V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// check if there is a key in the entries map to get the value of
	listElement, ok := c.entries[key]

	if !ok {
		var nilValue V
		return &nilValue, false
	}

	// if we reach here there is a key in the entries map

	// move the list.Element the front of the list
	c.list.MoveToFront(listElement)

	// get the value from the list element
	cacheValue := listElement.Value.(*CacheValue[K, V])

	return &cacheValue.value, true
}
