package main

import (
	"sync"
)

// ---------- Generics Learning
// Cache is the name of the generic type
// The square brackets [] indicate that this is a generic type declaration
// K and V are type parameters which as as placeholders for the actual types that will be provided when using the Cache type
// comparable is a constrain on the K type parameter, which means K must be a type that can be compared using the == and != operators
// any is a constraint on the V type parameter which means that V can be any type including concrete types and interface types

type Cache[K comparable, V any] struct {
	// inside the Cache type you can use the type parameters K and V as if they were regular types
	// this allows you to write generic code that works with different types without duplicating code or using type assertions

	// a map to store the cache data in
	data map[K]V

	// a mutex to control the concurrent access to the data map
	mu sync.Mutex

	// need some way to remember what key was last accessed

	// if a new key is added when there are already 3
	// we need to know which existing key to delete

	// could use time(?)... no not good... too many operations and too fast

	// use some sort of data structure?
	// [0,1,2] and rearrange the indices based on last accessed(?)..
	// but that involves moving the items in the array around a lot

	// google LRU Cache data structures
	// "hash map" and "(doubly) linked list"
}

func NewCache[K comparable, V any](entryLimit int) Cache[K, V] {
	return Cache[K, V]{
		// make a map using the generics K and V
		// and use the entryLimit as the capacity of the map
		data: make(map[K]V, entryLimit),
	}
}

// Put adds the value to the cache,
// and returns a boolean to indicate whether a value already existed in the cache for that key.
// If there was previously a value, it replaces that value with this one.
// Any Put counts as a refresh in terms of LRU tracking.
func (c *Cache[K, V]) Put(key K, value V) bool {
	// use the mutex to lock and unlock access
	c.mu.Lock()
	defer c.mu.Unlock()

	// check if the key exists in the data map
	_, ok := c.data[key]

	//if the key didn't exist before
	if !ok {
		// add the key and its value
		c.data[key] = value

		// return false (to denote that the key didn't exist)
		return false
	}

	// if the key did exist before
	// overwrite the existing key's value with the value passed into Put()
	c.data[key] = value

	// return true (to denote that we overwrote the key's value)
	return true
}

// Get returns the value associated with the passed key,
// and a boolean to indicate whether a value was known or not.
// If not, nil is returned as the value.
// Any Get counts as a refresh in terms of LRU tracking.
func (c *Cache[K, V]) Get(key K) (*V, bool) {
	// use the mutex to lock and unlock access
	c.mu.Lock()
	defer c.mu.Unlock()

	// try to find the key in the data map
	value, ok := c.data[key]

	// if we can't find the key return a nil value and false (to denote an unsuccessful lookup)
	if !ok {
		// we have to make a nil value on the Generic Type V
		var nilValue V
		// before we can return it
		return &nilValue, false
	}

	// if we can find the key return it's value and true (to denote a successful lookup)
	return &value, true
}
