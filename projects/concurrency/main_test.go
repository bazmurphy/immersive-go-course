package main

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
)

func TestCache(t *testing.T) {
	t.Run("add a key to an empty cache", func(t *testing.T) {
		// instantiate a new cache with string keys and integer values
		testCache := NewCache[string, int](3)

		// try to read from the cache
		value, okOne := testCache.Get("baz")

		if okOne != false {
			t.Errorf("found a key in the cache where there should have been none")
		}

		// (!) deference the pointer to value
		if *value != 0 {
			t.Errorf("key's value should be 0 a nil type integer")
		}

		// try to add a key and value to the cache
		okTwo := testCache.Put("baz", 100)

		if okTwo != false {
			t.Errorf("found an existing value on the key in the cache where there should have been none")
		}

		// try to read from the cache
		value, okThree := testCache.Get("baz")

		if okThree != true {
			t.Errorf("could not find key when there should be one")
		}

		expected := 100

		// (!) deference the pointer to value
		if *value != expected {
			t.Errorf("got %v | want %v", value, expected)
		}
	})

	t.Run("add and update a key to an empty cache", func(t *testing.T) {
		// instantiate a new cache with string keys and integer values
		cache := NewCache[string, int](3)

		// try to read from the cache
		value, okOne := cache.Get("baz")

		if okOne != false {
			t.Errorf("found a key in the cache where there should have been none")
		}

		// (!) deference the pointer to value
		if *value != 0 {
			t.Errorf("key's value should be 0 a nil type integer")
		}

		// try to add a key and value to the cache
		okTwo := cache.Put("baz", 100)

		if okTwo != false {
			t.Errorf("")
		}

		// try to update the key's value
		okThree := cache.Put("baz", 200)

		if okThree != true {
			t.Errorf("Put should return true to denote a successful updating of that keys value")
		}

		// check the new value of the key
		value, okFour := cache.Get("baz")

		if okFour != true {
			t.Errorf("could not find key when there should be one")
		}

		expected := 200

		// (!) deference the pointer to value
		if *value != expected {
			t.Errorf("got %v | want %v", value, expected)
		}
	})

	t.Run("add 3 keys to an empty cache", func(t *testing.T) {
		// instantiate a new cache with string keys and integer values
		cache := NewCache[string, int](3)

		// add three key/values to the cache
		cache.Put("a", 10)
		cache.Put("b", 20)
		cache.Put("c", 30)

		expectedCacheData := map[string]int{
			"a": 10,
			"b": 20,
			"c": 30,
		}

		// error if the test cache data is not the same as the expected cache data
		if !reflect.DeepEqual(cache.data, expectedCacheData) {
			t.Errorf("got %v | want %v", cache.data, expectedCacheData)
		}
	})

	t.Run("add 4 keys to an empty cache", func(t *testing.T) {
		// instantiate a new cache with string keys and integer values
		cache := NewCache[string, int](3)

		// add three key/values to the cache
		cache.Put("a", 10)
		cache.Put("b", 20)
		cache.Put("c", 30)

		// try to add a fourth key/value to the cache
		ok := cache.Put("d", 30)

		expected := false

		// error if we are allowed to add a fourth/key value to the cache
		if ok != expected {
			t.Errorf("got %v | want %v", ok, expected)
		}
	})
}

func TestCacheWithConcurrency(t *testing.T) {
	// create a new cache
	cache := NewCache[string, int](3)

	// define a wait group
	var wg sync.WaitGroup

	// spawn 100 goroutines
	for i := 0; i < 100; i++ {
		// increment the wait group counter
		wg.Add(1)

		// create a dynamic key
		dynamicKey := fmt.Sprintf("key-%d", i)

		// spawn a new goroutine
		go func() {
			// decrement the wait group counter
			defer wg.Done()

			// in each goroutine run a Put() and a Get() on the cache 100 times
			for j := 0; j < 100; j++ {

				// create a dynamic value
				dynamicValue := i + j + 1

				// try to Put() to the cache
				cache.Put(dynamicKey, dynamicValue)

				// try to Get() from the cache
				value, ok := cache.Get(dynamicKey)

				// error if we can't find the key in the cache
				if !ok {
					t.Errorf("cache.Get() ok : got %v | want %v", ok, false)
				}

				// error if the values do not match
				if *value != dynamicValue {
					t.Errorf("cache.Get() value: got %v | want %v", *value, dynamicValue)
				}
			}
		}()
	}

	// wait until all the goroutines are finished
	wg.Wait()
}
