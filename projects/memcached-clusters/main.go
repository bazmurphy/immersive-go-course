package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

func parseServerFlags() (string, []string) {
	// initialise these to use flag.StringVar (not flag.String) to avoid having to pass pointers around
	var mcRouterServer string
	var memcachedServers string

	// setup the flags for the command line tool
	flag.StringVar(&mcRouterServer, "mcrouter", "", "the mcrouter server address")
	flag.StringVar(&memcachedServers, "memcacheds", "", "the list of memcached server addresses")

	// parse the flags
	flag.Parse()

	// check the flags
	if mcRouterServer == "" {
		fmt.Println("error: mcrouter server address was not provided, please provide one with --mcrouter=X")
		os.Exit(1)
	}

	if memcachedServers == "" {
		fmt.Println("error: memcached server addresses were not provided, please provide them with --memcacheds=X")
		os.Exit(1)
	}

	// breakup the string into a slice of individual memcached server addresses
	memcachedServersSlice := strings.Split(memcachedServers, ",")

	return mcRouterServer, memcachedServersSlice
}

func guessAndPrintTopology(memcachedServersWithKeyCountInt int, totalMemcachedServers int) {
	// return a topology guess based on how many memcached servers had the key
	switch memcachedServersWithKeyCountInt {
	case totalMemcachedServers:
		fmt.Printf("🟣 Replicated topology: %d/%d memcached servers had the key\n", memcachedServersWithKeyCountInt, totalMemcachedServers)
	case 1:
		fmt.Printf("🟣 Sharded topology: %d/%d memcached servers had the key\n", memcachedServersWithKeyCountInt, totalMemcachedServers)
	default:
		fmt.Printf("🟣 Undetermined topology: %d/%d memcached servers had the key\n", memcachedServersWithKeyCountInt, totalMemcachedServers)
	}
}

func checkKeyAcrossMemcachedServers(memcachedServersSlice []string, key string) int {
	// initialise a count
	var memcachedServersWithKeyCount int32

	// initialise a wait group
	var wg sync.WaitGroup

	// loop over the memcached servers
	for _, memcachedServer := range memcachedServersSlice {
		// increment the wait group
		wg.Add(1)

		// (!) A note on "yellow squigglies"
		// In Go Versions PRE 1.22
		// you needed to create a new variable for each iteration of the loop like:
		// server := memcachedServer
		// and then pass it to the go routine as an argument like:
		// gofunc(server string){...}(server)
		// so it has it's own copy with the correct value (not modified)
		// however this is NOT necessary as we are using Go 1.22+

		// spawn a go routine per memcached server to run these gets concurrently
		go func() {
			// decrement the wait group
			defer wg.Done()

			// make a new client for each memcached server
			memcachedClient := memcache.New(memcachedServer)

			// attempt to get the key from that specific memcached server
			item, err := memcachedClient.Get(key)
			if err != nil {
				// does this fall under the category of ignoring the error silently(?) and/or not handling it(?) what is the way to do this(?)
				fmt.Printf("🔴 memcached server %s | key: %s NOT FOUND\n", memcachedServer, key)
			} else {
				fmt.Printf("🟢 memcached server %s | key: %s FOUND | value: %vs\n", memcachedServer, key, string(item.Value))
				// if we find the key increment the count
				atomic.AddInt32(&memcachedServersWithKeyCount, 1)
			}
		}()
	}

	// wait for all (sub)goroutines to finish
	wg.Wait()

	// convert the int32 (necessary for atomic operations) to an int (janky!!)
	memcachedServersWithKeyCountInt := int(memcachedServersWithKeyCount)

	return memcachedServersWithKeyCountInt
}

func main() {
	start := time.Now()

	mcRouterServer, memcachedServersSlice := parseServerFlags()

	// make a mcrouter client
	mcRouterClient := memcache.New(mcRouterServer)

	// ping all instances
	err := mcRouterClient.Ping()
	if err != nil {
		fmt.Printf("error: mcrouter failed to ping all memcached servers: %v\n", err)
	}
	fmt.Printf("🟢 mcrouter PING | successfully pinged all memcached servers\n")

	// make a key and value to test later
	testKey := "testkey"
	testValue := "testvalue"

	// attempt to set the key into the cache
	err = mcRouterClient.Set(&memcache.Item{Key: testKey, Value: []byte(testValue)})
	if err != nil {
		fmt.Printf("error: failed to SET the testItem into the cache: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("🟢 mcrouter SET | key: %v | value: %v\n", testKey, testValue)

	// attempt to get the key from the cache
	testItem, err := mcRouterClient.Get(testKey)
	if err != nil {
		fmt.Printf("error: failed to GET the testItem from the cache: %v\n", err)
		os.Exit(1)
	}
	// check the get operation
	fmt.Printf("🟢 mcrouter GET | key: %v | value: %v\n", testItem.Key, string(testItem.Value))

	// ========== BATTLE SCAR LEFT HERE DELIBERATELY ==========
	// ---------- DEBUG NIGHTMARE BEGINS HERE ----------
	//
	// ERROR:
	// memcache: unexpected line in get response: "SERVER_ERROR unexpected result mc_res_unknown (0) for get\r\n"
	//
	// CHANGING LINE 389 in go/pkg/mod/github.com/bradfitz/gomemecache/memcache/memcache.go
	// CHANGE "gets" to "get"
	// BEFORE:
	// if _, err := fmt.Fprintf(rw, "gets %s\r\n", strings.Join(keys, " ")); err != nil {
	// AFTER:
	// if _, err := fmt.Fprintf(rw, "get %s\r\n", strings.Join(keys, " ")); err != nil {
	//
	// ---------- DEBUG NIGHTMARE ENDS HERE ----------

	// attempt to check the key's existence on each individual memcached server
	memcachedServersWithKeyCountInt := checkKeyAcrossMemcachedServers(memcachedServersSlice, testKey)

	finish := time.Now()
	duration := finish.Sub(start)
	fmt.Printf("✅ topology scan completed in %v\n", duration)

	// establish how many memcached servers were initially provided
	totalMemcachedServers := len(memcachedServersSlice)

	// guess and print the topology of the mcrouter setup
	guessAndPrintTopology(memcachedServersWithKeyCountInt, totalMemcachedServers)
}

// replicated :

// $ go run . --mcrouter=localhost:11211 --memcacheds=localhost:11212,localhost:11213,localhost:11214
// 🟢 mcrouter PING | successfully pinged all memcached servers
// 🟢 mcrouter SET | key: testkey | value: testvalue
// 🟢 mcrouter GET | key: testkey | value: testvalue
// 🟢 memcached server localhost:11212 | key: testkey FOUND | value: testvalues
// 🟢 memcached server localhost:11213 | key: testkey FOUND | value: testvalues
// 🟢 memcached server localhost:11214 | key: testkey FOUND | value: testvalues
// ✅ topology scan completed in 9.9336ms
// 🟣 Replicated topology: 3/3 memcached servers had the key

// sharded :

// $ go run . --mcrouter=localhost:11211 --memcacheds=localhost:11212,localhost:11213,localhost:11214
// 🟢 mcrouter PING | successfully pinged all memcached servers
// 🟢 mcrouter SET | key: testkey | value: testvalue
// 🟢 mcrouter GET | key: testkey | value: testvalue
// 🟢 memcached server localhost:11212 | key: testkey FOUND | value: testvalues
// 🔴 memcached server localhost:11214 | key: testkey NOT FOUND
// 🔴 memcached server localhost:11213 | key: testkey NOT FOUND
// ✅ topology scan completed in 9.9734ms
// 🟣 Sharded topology: 1/3 memcached servers had the key
