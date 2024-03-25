package main

import (
	"fmt"
	"sync"
	"testing"
)

func TestCache(t *testing.T) {
	t.Run("try to get a key from an empty cache", func(t *testing.T) {
		cache := NewCache[string, int](3)

		_, ok := cache.Get("a")

		if ok != false {
			t.Errorf("found a key in the cache where there should have been none")
		}
	})

	t.Run("add a single key/value to an empty cache", func(t *testing.T) {
		cache := NewCache[string, int](3)

		insertKey := "a"
		insertValue := 100

		ok := cache.Put(insertKey, insertValue)

		if ok != false {
			t.Errorf("found an existing value on the key in the cache, where there should have been no existing value")
		}

		cacheListElementPointer, ok := cache.entries[insertKey]

		if !ok {
			t.Errorf("key %v should now exist in the cache but does not", insertKey)
		}

		// (!) NEED TO LEARN >>> TYPE ASSERTION/CONVERSION
		// "It performs a type assertion on cacheListElementPointer.Value, checking if the value stored there is a pointer to a CacheValue[string, int] struct."
		// "If the type assertion is successful, it retrieves the concrete value (the pointer to the CacheValue struct) and assigns it to the cacheValue variable."
		cacheValue := cacheListElementPointer.Value.(*CacheValue[string, int])

		if cacheValue.value != insertValue {
			t.Errorf("cacheValue.value: got %v want %v", cacheValue.value, insertValue)
		}
	})

	t.Run("add and update a single key/value to an empty cache", func(t *testing.T) {
		cache := NewCache[string, int](3)

		insertKey := "a"
		insertValue := 100

		ok := cache.Put(insertKey, insertValue)

		if ok != false {
			t.Errorf("found an existing value on the key in the cache, where there should have been no existing value")
		}

		updatedValue := 200

		ok = cache.Put(insertKey, updatedValue)

		if ok != true {
			t.Errorf("Put should return true to denote a successful updating of that key's value")
		}

		valuePointer, ok := cache.Get(insertKey)

		if ok != true {
			t.Errorf("could not find key %v when there should be one", insertKey)
		}

		// (!) deference the pointer to a value
		if *valuePointer != updatedValue {
			t.Errorf("got %v | want %v", *valuePointer, updatedValue)
		}
	})

	t.Run("add 3 keys to an empty cache", func(t *testing.T) {
		cache := NewCache[string, int](3)

		insertKeyValues := map[string]int{"a": 1, "b": 2, "c": 3}

		for insertKey, insertValue := range insertKeyValues {
			cache.Put(insertKey, insertValue)
		}

		for insertKey, insertValue := range insertKeyValues {
			cacheListElementPointer, ok := cache.entries[insertKey]

			if !ok {
				t.Errorf("key %v was not found in the cache and should have been", insertKey)
			}

			// (!) type assertion/conversion
			cacheValue := cacheListElementPointer.Value.(*CacheValue[string, int])

			if cacheValue.value != insertValue {
				t.Errorf("cacheValue.value: got %v want %v", cacheValue.value, insertValue)
			}
		}

		expectedCacheListOrder := []int{3, 2, 1}

		cacheListElementPointer := cache.list.Front()

		for _, expectedCacheListValue := range expectedCacheListOrder {
			// (!) type assertion/conversion
			cacheValue := cacheListElementPointer.Value.(*CacheValue[string, int])

			if cacheValue.value != expectedCacheListValue {
				t.Errorf("cache list element value: got %v | want %v", cacheValue.value, expectedCacheListValue)
			}

			cacheListElementPointer = cacheListElementPointer.Next()
		}
	})
}

// note: need to add another test here for 6 keys to an empty cache, to check the LRU order

// this test is failing, probably because of the way i am concurrently accessing the cache
func TestCacheConcurrency(t *testing.T) {
	cache := NewCache[string, int](3)

	// define a wait group
	var wg sync.WaitGroup

	// a loop to spawn X number of goroutines
	for i := 0; i < 10; i++ {
		// increment the wait group counter
		wg.Add(1)

		// create a dynamic key
		dynamicKey := fmt.Sprintf("key-%d", i)

		// spawn a new goroutine
		go func() {
			// decrement the wait group counter
			defer wg.Done()

			// an inner loop to run a Put() and a Get() in each goroutine
			for j := 0; j < 10; j++ {

				// create a dynamic value
				dynamicValue := i + j + 1

				// try to Put() to the cache
				cache.Put(dynamicKey, dynamicValue)

				// ----- DEBUG
				fmt.Printf("PUT | dynamicKey: %v | dynamicValue: %v\n", dynamicKey, dynamicValue)

				// try to Get() from the cache
				valuePointer, ok := cache.Get(dynamicKey)

				// ----- DEBUG
				fmt.Printf("GET | dynamicKey: %v | valuePointer: %v | *valuePointer: %v | ok: %v\n", dynamicKey, valuePointer, *valuePointer, ok)

				// error if we can't find the key in the cache
				if !ok {
					t.Errorf("cache.Get() ok: got %v | want %v", ok, false)
				}

				// error if the values do not match
				// (!) deference the pointer to a value
				if *valuePointer != dynamicValue {
					t.Errorf("cache.Get() value does not match: got %v | want %v", *valuePointer, dynamicValue)
				}
			}
		}()
	}

	// wait until all the goroutines are finished
	wg.Wait()

	// check the final state of the cache

	// for cacheEntriesKey, cacheListElementPointer := range cache.entries {
	// 	cacheValue := cacheListElementPointer.Value.(*CacheValue[string, int])
	// 	fmt.Printf("(map) cacheEntriesKey: %v (list) cacheValue.value: %v\n", cacheEntriesKey, cacheValue.value)
	// }
}
