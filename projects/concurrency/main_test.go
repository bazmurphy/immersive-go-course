package main

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
)

func TestCache(t *testing.T) {
	t.Run("test getting a key from empty cache", func(t *testing.T) {
		cache := NewCache[string, int](3)

		_, found := cache.Get("a")

		if found {
			t.Errorf("found a key in the cache when it should be empty")
		}
	})

	t.Run("test adding a key", func(t *testing.T) {
		cache := NewCache[string, int](3)

		cache.Put("a", 1)
		value, found := cache.Get("a")

		if found == false {
			t.Errorf("key 'a' should exist in the cache but does not")
		}
		if *value != 1 {
			t.Errorf("value: received %v expected %v", &value, 100)
		}
	})

	t.Run("test updating a key", func(t *testing.T) {
		cache := NewCache[string, int](3)

		cache.Put("a", 1)
		cache.Put("a", 2)
		value, found := cache.Get("a")

		if found == false {
			t.Errorf("key 'a' should exist in the cache but does not")
		}
		if *value != 2 {
			t.Errorf("value: received %v | expected %v", &value, 200)
		}
	})

	t.Run("test adding until capacity", func(t *testing.T) {
		cache := NewCache[string, int](3)

		cache.Put("a", 1)
		cache.Put("b", 2)
		cache.Put("c", 3)

		value, _ := cache.Get("a")
		if *value != 1 {
			t.Errorf("value: received %v | expected %v", &value, 1)
		}
		value, _ = cache.Get("b")
		if *value != 2 {
			t.Errorf("value: received %v | expected %v", &value, 2)
		}
		value, _ = cache.Get("c")
		if *value != 3 {
			t.Errorf("value: received %v | expected %v", &value, 3)
		}
	})

	t.Run("test remove on full capacity", func(t *testing.T) {
		cache := NewCache[string, int](3)

		cache.Put("a", 1)
		cache.Put("b", 2)
		cache.Put("c", 3)
		cache.Put("d", 4)
		cache.Put("e", 5)
		cache.Put("f", 6)

		_, found := cache.Get("a")
		if found != false {
			t.Errorf("key 'a' should not exist in the cache but does")
		}
		_, found = cache.Get("b")
		if found != false {
			t.Errorf("key 'b' should not exist in the cache but does")
		}
		_, found = cache.Get("c")
		if found != false {
			t.Errorf("key 'c' should not exist in the cache but does")
		}

		value, _ := cache.Get("d")
		if *value != 4 {
			t.Errorf("value: received %v | expected %v", &value, 4)
		}
		value, _ = cache.Get("e")
		if *value != 5 {
			t.Errorf("value: received %v | expected %v", &value, 5)
		}
		value, _ = cache.Get("f")
		if *value != 6 {
			t.Errorf("value: received %v | expected %v", &value, 6)
		}
	})

	t.Run("test remove in different order of read and write", func(t *testing.T) {
		cache := NewCache[string, int](3)

		cache.Put("a", 1)
		cache.Put("b", 2)
		cache.Put("c", 3)
		cache.Get("x")
		cache.Put("y", 9)
		cache.Get("z")
		cache.Put("d", 4)
		cache.Put("e", 5)
		cache.Put("f", 6)

		_, ok := cache.Get("a")
		if ok != false {
			t.Errorf("got %v | want %v", ok, 1)
		}
		_, ok = cache.Get("b")
		if ok != false {
			t.Errorf("got ok for Get 'b'")
		}
		_, ok = cache.Get("c")
		if ok != false {
			t.Errorf("got ok for Get 'c'")
		}

		value, _ := cache.Get("d")
		if *value != 4 {
			t.Errorf("got %v | want %v", value, 4)
		}
		value, _ = cache.Get("e")
		if *value != 5 {
			t.Errorf("got %v | want %v", value, 5)
		}
		value, _ = cache.Get("f")
		if *value != 6 {
			t.Errorf("got %v | want %v", value, 6)
		}
	})
}

func TestCacheGetConcurrency(t *testing.T) {
	// (method 3) using channels
	cache := NewCache[string, int](3)

	var putWriteCount int64

	for i := 1; i <= 3; i++ {
		dynamicKey := fmt.Sprintf("key-%d", i)
		dynamicValue := i

		ok := cache.Put(dynamicKey, dynamicValue)
		if ok == false {
			atomic.AddInt64(&putWriteCount, 1)
		}
	}

	fmt.Println("BAZ DEBUG | cache.entries", cache.entries["key-1"])

	// make a buffered getFoundChannel with capacity 100
	getFoundChannel := make(chan int, 100)

	var wg sync.WaitGroup

	var getFoundCount int64

	// spawn a goroutine to receive from the getFoundChannel
	go func() {
		for range getFoundChannel {
			atomic.AddInt64(&getFoundCount, 1)
		}
	}()

	var key1GetAttempts int64
	var key2GetAttempts int64
	var key3GetAttempts int64

	// an outer loop to spawn X number of goroutines
	for i := 1; i <= 100; i++ {
		wg.Add(1)

		// spawn a new goroutine
		go func() {
			defer wg.Done()

			randomKeyNumber := rand.Intn(3) + 1

			dynamicKey := fmt.Sprintf("key-%d", randomKeyNumber)

			switch dynamicKey {
			case "key-1":
				atomic.AddInt64(&key1GetAttempts, 1)
			case "key-2":
				atomic.AddInt64(&key2GetAttempts, 1)
			case "key-3":
				atomic.AddInt64(&key3GetAttempts, 1)
			}

			// an inner loop to run Get() X number of times
			for i := 1; i <= 100; i++ {
				_, ok := cache.Get(dynamicKey)
				if ok {
					getFoundChannel <- 1
				}
			}
		}()
	}

	wg.Wait()

	fmt.Printf("key1GetAttempts %d \nkey2GetAttempts %d\nkey3GetAttempts %d\n", key1GetAttempts, key1GetAttempts, key1GetAttempts)

	// close the getFoundChannel
	close(getFoundChannel)

	totalOperations := putWriteCount + getFoundCount

	fmt.Printf("putWrite %d\ngetFound %d\ntotalOperations %d\n", putWriteCount, getFoundCount, totalOperations)

	if putWriteCount != 3 {
		t.Errorf("put write operations: received %v | expected %v", putWriteCount, 3)
	}
	if getFoundCount != 10000 {
		t.Errorf("get found operations: received %v | expected %v", getFoundCount, 10000)
	}
	if totalOperations != 10003 {
		t.Errorf("total operations: received %v | expected %v", totalOperations, 10003)
	}
}

func TestCachePutAndGetConcurrency(t *testing.T) {
	// (method 2) using atomics
	cache := NewCache[string, int](3)

	var putWriteCount int64
	var putUpdateCount int64
	var getFoundCount int64
	var getNotFoundCount int64

	var wg sync.WaitGroup

	for i := 1; i <= 100; i++ {
		wg.Add(2)

		dynamicKey := fmt.Sprintf("key-%d", i)

		go func() {
			defer wg.Done()

			for j := 1; j <= 100; j++ {
				dynamicValue := i + j

				ok := cache.Put(dynamicKey, dynamicValue)
				if !ok {
					atomic.AddInt64(&putWriteCount, 1)
				} else {
					atomic.AddInt64(&putUpdateCount, 1)
				}
			}
		}()

		go func() {
			defer wg.Done()

			for j := 1; j <= 100; j++ {
				_, ok := cache.Get(dynamicKey)
				if !ok {
					atomic.AddInt64(&getNotFoundCount, 1)
				} else {
					atomic.AddInt64(&getFoundCount, 1)
				}
			}
		}()
	}

	wg.Wait()

	totalOperations := putWriteCount + putUpdateCount + getFoundCount + getNotFoundCount

	// fmt.Printf("putWriteCount %d\nputUpdateCount %d\ngetFoundCount %d\ngetNotFoundCount %d\ntotalOperations %d\n", putWriteCount, putUpdateCount, getFoundCount, getNotFoundCount, totalOperations)

	if totalOperations != 20000 {
		t.Errorf("total operations: received %v | expected %v", totalOperations, 20000)
	}
}

// --------------------
// Notes, from session with Laura:

// func TestCacheGetConcurrency(t *testing.T) {
// 	cache := NewCache[string, int](3)
// 	successfulWrites := 0

// 	// var lock sync.Mutex
// 	// successfulReads := 0
// 	// var sr atomic.Uint64

// 	successfulReadsChannel := make(chan int, 100*100)

// 	// PUT 3 key/values to the cache (to then GET from)
// 	for i := 1; i <= 3; i++ {
// 		dynamicKey := fmt.Sprintf("key-%d", i)
// 		dynamicValue := i * 10

// 		ok := cache.Put(dynamicKey, dynamicValue)
// 		if ok {
// 			successfulWrites++
// 		}
// 	}

// 	// define a wait group
// 	var wg sync.WaitGroup

// 	// initialise a successful reads variable
// 	var successfulReads int64

// 	// create a separate goroutine to receive from the successfulReadsChannel
// 	go func() {
// 		for range successfulReadsChannel {
// 			// and increment successfulReads
// 			atomic.AddInt64(&successfulReads, 1)
// 		}
// 	}()

// 	// an outer loop to spawn X number of goroutines
// 	for i := 1; i <= 100; i++ {
// 		// increment the wait group counter
// 		wg.Add(1)

// 		// spawn a new goroutine
// 		go func() {
// 			// decrement the wait group counter
// 			defer wg.Done()

// 			// generate a random number between 1-3
// 			randomKeyNumber := rand.Intn(3) + 1

// 			// create a key
// 			dynamicKey := fmt.Sprintf("key-%d", randomKeyNumber)

// 			// an inner loop to run Get() X number of times
// 			for i := 1; i <= 100; i++ {
// 				// try to Get() from the cache
// 				_, ok := cache.Get(dynamicKey)
// 				if ok {
// 					// 1 - use a mutex lock/unlock
// 					// 2 - use atomics
// 					// 3 - use channels
// 					// lock.Lock()
// 					// successfulReads++
// 					// lock.Unlock()
// 					// sr.Add(1)
// 					successfulReadsChannel <- 1
// 				}
// 			}
// 		}()
// 	}

// 	// wait until all the goroutines are finished
// 	wg.Wait()

// 	// close the channel to signal that no more values will be sent
// 	close(successfulReadsChannel)

// 	fmt.Printf("successfulReads: %d\n", successfulReads)
// }
