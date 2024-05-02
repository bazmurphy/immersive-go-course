package main

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
)

type TestKeyValue struct {
	key   string
	value int
}

func DisplayCacheContents(c *Cache[string, int]) {
	for cacheEntriesKey, cacheListElementPointer := range c.entries {
		cacheValue := cacheListElementPointer.Value.(*CacheEntry[string, int])
		fmt.Printf("(map) cacheEntriesKey: %v (list) cacheValue.value: %v\n", cacheEntriesKey, cacheValue.value)
	}
}

func PrintCacheStats(c *Cache[string, int]) {
	fmt.Println("----- Cache Stats ------")
	fmt.Printf("%-22s %v\n", "Successful Reads:", c.stats.successfulReads)
	fmt.Printf("%-22s %v\n", "Failed Reads:", c.stats.failedReads)
	fmt.Printf("%-22s %v\n", "Successful Writes:", c.stats.successfulWrites)
	fmt.Printf("%-22s %v\n", "Failed Writes:", c.stats.failedWrites)
	fmt.Printf("%-22s %v\n", "Successful Overwrites:", c.stats.successfulOverwrites)
	fmt.Printf("%-22s %v\n", "Successful Removes:", c.stats.successfulRemoves)
	fmt.Println("------------------------")
}

func TestCache(t *testing.T) {
	t.Run("try to get a key from an empty cache", func(t *testing.T) {
		cache := NewCache[string, int](3)

		_, ok := cache.Get("a")

		if ok != false {
			t.Errorf("found a key in the cache where there should have been none")
		}

		expectedFailedReads := 1

		if cache.stats.failedReads != expectedFailedReads {
			t.Errorf("failedReads: got %v | want %v", cache.stats.failedReads, expectedFailedReads)
		}
	})

	t.Run("add a single new key/value to an empty cache", func(t *testing.T) {
		cache := NewCache[string, int](3)

		insertKeyValue := TestKeyValue{"a", 100}

		ok := cache.Put(insertKeyValue.key, insertKeyValue.value)

		if ok != false {
			t.Errorf("found an existing value on the key in the cache, where there should have been no existing value")
		}

		cacheListElementPointer, ok := cache.entries[insertKeyValue.key]

		if !ok {
			t.Errorf("key %v should now exist in the cache but does not", insertKeyValue.key)
		}

		// (!) NEED TO LEARN >>> TYPE ASSERTION/CONVERSION
		// "It performs a type assertion on cacheListElementPointer.Value, checking if the value stored there is a pointer to a CacheEntry[string, int] struct."
		// "If the type assertion is successful, it retrieves the concrete value (the pointer to the CacheEntry struct) and assigns it to the cacheValue variable."
		cacheValue := cacheListElementPointer.Value.(*CacheEntry[string, int])

		if cacheValue.value != insertKeyValue.value {
			t.Errorf("cacheValue.value: got %v want %v", cacheValue.value, insertKeyValue.value)
		}

		expectedSuccessfulWrites := 1

		if cache.stats.successfulWrites != expectedSuccessfulWrites {
			t.Errorf("successfulWrites: got %v | want %v", cache.stats.successfulWrites, expectedSuccessfulWrites)
		}
	})

	t.Run("add and update a single new key/value to an empty cache", func(t *testing.T) {
		cache := NewCache[string, int](3)

		insertKeyValue := TestKeyValue{"a", 100}

		ok := cache.Put(insertKeyValue.key, insertKeyValue.value)

		if ok != false {
			t.Errorf("found an existing value on the key in the cache, where there should have been no existing value")
		}

		updatedValue := 200

		ok = cache.Put(insertKeyValue.key, updatedValue)

		if ok != true {
			t.Errorf("Put should return true to denote a successful updating of that key's value")
		}

		value, ok := cache.Get(insertKeyValue.key)

		if ok != true {
			t.Errorf("could not find key %v when there should be one", insertKeyValue.key)
		}

		if value != updatedValue {
			t.Errorf("got %v | want %v", value, updatedValue)
		}

		expectedSuccessfulWrites := 1
		expectedSuccessfulOverwrites := 1
		expectedSuccessfulReads := 1

		if cache.stats.successfulWrites != expectedSuccessfulWrites {
			t.Errorf("successfulWrites: got %v | want %v", cache.stats.successfulWrites, expectedSuccessfulWrites)
		}

		if cache.stats.successfulOverwrites != expectedSuccessfulOverwrites {
			t.Errorf("successfulOverwrites: got %v | want %v", cache.stats.successfulOverwrites, expectedSuccessfulOverwrites)
		}

		if cache.stats.successfulReads != expectedSuccessfulReads {
			t.Errorf("successfulReads: got %v | want %v", cache.stats.successfulReads, expectedSuccessfulReads)
		}
	})

	t.Run("add 3 new keys to an empty cache", func(t *testing.T) {
		cache := NewCache[string, int](3)

		insertKeyValues := []TestKeyValue{
			{key: "a", value: 1},
			{key: "b", value: 2},
			{key: "c", value: 3},
		}

		for _, insertKeyValue := range insertKeyValues {
			cache.Put(insertKeyValue.key, insertKeyValue.value)
		}

		for _, insertKeyValue := range insertKeyValues {
			cacheListElementPointer, ok := cache.entries[insertKeyValue.key]

			if !ok {
				t.Errorf("key %v was not found in the cache and should have been", insertKeyValue.key)
			}

			// (!) type assertion/conversion
			cacheValue := cacheListElementPointer.Value.(*CacheEntry[string, int])

			if cacheValue.value != insertKeyValue.value {
				t.Errorf("cacheValue.value: got %v want %v", cacheValue.value, insertKeyValue.value)
			}
		}

		expectedCacheListOrder := []int{3, 2, 1}

		cacheListElementPointer := cache.list.Front()

		for _, expectedCacheListValue := range expectedCacheListOrder {
			// (!) type assertion/conversion
			cacheValue := cacheListElementPointer.Value.(*CacheEntry[string, int])

			if cacheValue.value != expectedCacheListValue {
				t.Errorf("cache list element value: got %v | want %v", cacheValue.value, expectedCacheListValue)
			}

			cacheListElementPointer = cacheListElementPointer.Next()
		}

		expectedSuccessfulWrites := 3

		if cache.stats.successfulWrites != expectedSuccessfulWrites {
			t.Errorf("successfulWrites: got %v | want %v", cache.stats.successfulWrites, expectedSuccessfulWrites)
		}
	})

	t.Run("add 6 new keys to an empty cache, to check removes", func(t *testing.T) {
		cache := NewCache[string, int](3)

		insertKeyValues := []TestKeyValue{
			{key: "a", value: 1},
			{key: "b", value: 2},
			{key: "c", value: 3},
			{key: "d", value: 4},
			{key: "e", value: 5},
			{key: "f", value: 6},
		}

		for _, insertKeyValue := range insertKeyValues {
			cache.Put(insertKeyValue.key, insertKeyValue.value)
		}

		// note: loop from index 3 to index 5
		for _, insertKeyValue := range insertKeyValues[3:] {
			cacheListElementPointer, ok := cache.entries[insertKeyValue.key]

			if !ok {
				t.Errorf("key %v was not found in the cache and should have been", insertKeyValue.key)
			}

			// (!) type assertion/conversion
			cacheValue := cacheListElementPointer.Value.(*CacheEntry[string, int])

			if cacheValue.value != insertKeyValue.value {
				t.Errorf("cacheValue.value: got %v want %v", cacheValue.value, insertKeyValue.value)
			}
		}

		expectedCacheListOrder := []int{6, 5, 4}

		cacheListElementPointer := cache.list.Front()

		for _, expectedCacheListValue := range expectedCacheListOrder {
			// (!) type assertion/conversion
			cacheValue := cacheListElementPointer.Value.(*CacheEntry[string, int])

			if cacheValue.value != expectedCacheListValue {
				t.Errorf("cache list element value: got %v | want %v", cacheValue.value, expectedCacheListValue)
			}

			cacheListElementPointer = cacheListElementPointer.Next()
		}

		expectedSuccessfulWrites := 6
		expectedSuccessfulRemoves := 3

		if cache.stats.successfulWrites != expectedSuccessfulWrites {
			t.Errorf("successfulWrites: got %v | want %v", cache.stats.successfulWrites, expectedSuccessfulWrites)
		}

		if cache.stats.successfulRemoves != expectedSuccessfulRemoves {
			t.Errorf("successfulRemoves: got %v | want %v", cache.stats.successfulRemoves, expectedSuccessfulRemoves)
		}
	})
}

func TestCachePutConcurrency(t *testing.T) {
	cache := NewCache[string, int](3)

	// define a wait group
	var wg sync.WaitGroup

	// an outer loop to spawn X number of goroutines
	for i := 1; i <= 100; i++ {
		// increment the wait group counter
		wg.Add(1)
		// create a dynamic key
		dynamicKey := fmt.Sprintf("key-%d", i)

		// spawn a new goroutine
		go func() {
			// decrement the wait group counter
			defer wg.Done()
			// an inner loop to run Put() X number of times
			for j := 1; j <= 100; j++ {
				// generate a dynamic value
				dynamicValue := i + j
				// try to Put() to the cache
				cache.Put(dynamicKey, dynamicValue)
			}
		}()
	}
	// wait until all the goroutines are finished
	wg.Wait()

	expectedSuccessfulWritesAndOverwrites := 10000

	if cache.stats.successfulWrites+cache.stats.successfulOverwrites != expectedSuccessfulWritesAndOverwrites {
		t.Errorf("successfulWrites+successfulOverwrites: got %v | want %v", (cache.stats.successfulWrites + cache.stats.successfulOverwrites), expectedSuccessfulWritesAndOverwrites)
	}

	// check the final state of the cache
	for cacheEntriesKey, cacheListElementPointer := range cache.entries {
		cacheValue := cacheListElementPointer.Value.(*CacheEntry[string, int])
		fmt.Printf("(map) cacheEntriesKey: %v (list) cacheValue.value: %v\n", cacheEntriesKey, cacheValue.value)
	}

	DisplayCacheContents(&cache)
	PrintCacheStats(&cache)
}

func TestCacheGetConcurrency(t *testing.T) {
	cache := NewCache[string, int](3)

	// PUT 3 key/values to the cache (to then GET from)
	for i := 1; i <= 3; i++ {
		dynamicKey := fmt.Sprintf("key-%d", i)
		dynamicValue := i * 10
		cache.Put(dynamicKey, dynamicValue)
	}

	// define a wait group
	var wg sync.WaitGroup

	// an outer loop to spawn X number of goroutines
	for i := 1; i <= 100; i++ {
		// increment the wait group counter
		wg.Add(1)

		// spawn a new goroutine
		go func() {
			// decrement the wait group counter
			defer wg.Done()

			// generate a random number between 1-3
			randomKeyNumber := rand.Intn(3) + 1

			// create a key
			dynamicKey := fmt.Sprintf("key-%d", randomKeyNumber)

			// an inner loop to run Get() X number of times
			for i := 1; i <= 100; i++ {
				// try to Get() from the cache
				cache.Get(dynamicKey)
			}
		}()
	}

	// wait until all the goroutines are finished
	wg.Wait()

	expectedSuccessfulWrites := 3
	expectedSuccessfulReads := 10000

	if cache.stats.successfulWrites != expectedSuccessfulWrites {
		t.Errorf("successfulWrites: got %v | want %v", cache.stats.successfulWrites, expectedSuccessfulWrites)
	}

	if cache.stats.successfulReads != expectedSuccessfulReads {
		t.Errorf("successfulRemoves: got %v | want %v", cache.stats.successfulReads, expectedSuccessfulReads)
	}

	DisplayCacheContents(&cache)
	PrintCacheStats(&cache)
}
