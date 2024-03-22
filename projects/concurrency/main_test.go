package main

import (
	"reflect"
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
			t.Errorf("")
		}

		// try to update the key's value
		okThree := testCache.Put("baz", 200)

		if okThree != true {
			t.Errorf("Put should return true to denote a successful updating of that keys value")
		}

		// check the new value of the key
		value, okFour := testCache.Get("baz")

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
		testCache := NewCache[string, int](3)

		// add three key/values to the cache
		testCache.Put("a", 10)
		testCache.Put("b", 20)
		testCache.Put("c", 30)

		expectedCacheData := map[string]int{
			"a": 10,
			"b": 20,
			"c": 30,
		}

		// error if the test cache data is not the same as the expected cache data
		if !reflect.DeepEqual(testCache.data, expectedCacheData) {
			t.Errorf("got %v | want %v", testCache.data, expectedCacheData)
		}
	})

	t.Run("add 4 keys to an empty cache", func(t *testing.T) {
		// instantiate a new cache with string keys and integer values
		testCache := NewCache[string, int](3)

		// add three key/values to the cache
		testCache.Put("a", 10)
		testCache.Put("b", 20)
		testCache.Put("c", 30)

		// try to add a fourth key/value to the cache
		ok := testCache.Put("d", 30)

		expected := false

		// error if we are allowed to add a fourth/key value to the cache
		if ok != expected {
			t.Errorf("got %v | want %v", ok, expected)
		}
	})
}
