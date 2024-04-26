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

func main() {
	start := time.Now()

	// setup the flags for the command line tool
	mcRouterServerAddress := flag.String("mcrouter", "", "the mcrouter server address")
	memcachedServerAddresses := flag.String("memcacheds", "", "the list of memcached server addresses")

	// // parse the flags
	flag.Parse()

	// check the flags
	if *mcRouterServerAddress == "" {
		fmt.Println("error: mcrouter server address was not provided, please provide one with --mcrouter=X")
		os.Exit(1)
	}

	if *memcachedServerAddresses == "" {
		fmt.Println("error: memcached server addresses were not provided, please provide them with --memcacheds=X")
		os.Exit(1)
	}

	// make a mcrouter client
	mcRouterClient := memcache.New(*mcRouterServerAddress)

	// ping all instances
	err := mcRouterClient.Ping()
	if err != nil {
		fmt.Printf("error: failed to ping all instances: %v\n", err)
	}

	// make a key and value to test later
	myKey := "mykey"
	myValue := "myvalue"

	// set the key
	err = mcRouterClient.Set(&memcache.Item{Key: myKey, Value: []byte(myValue), CasID: 1})
	if err != nil {
		// fmt.Printf("error: failed to write the item into the cache: %v\n", err)
		fmt.Println("mcRouterClient.Set err", err)
		os.Exit(1)
	}

	// get the key
	_, err = mcRouterClient.Get(myKey)
	if err != nil {
		// fmt.Printf("error: failed to read the key from the cache: %v\n", err)
		fmt.Println("mcRouterClient.Get err", err)
		os.Exit(1)
	}

	// BATTLE SCAR LEFT IN HERE DELIBERATELY
	//
	// --------- DEBUG NIGHTMARE BEGINS HERE ---------
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
	// check the get operation
	// fmt.Printf("mcrouter client GET | key: %v value: %v\n", item.Key, string(item.Value))
	// ---------DEBUG NIGHTMARE ON HOLD HERE ---------

	// breakup the string into individual memcached server addresses
	memcachedServers := strings.Split(*memcachedServerAddresses, ",")

	// initialise a count
	var memcachedServersWithKeyCount int32

	// initialise a wait group
	var wg sync.WaitGroup

	// loop over the memcached servers
	for _, memcachedServer := range memcachedServers {
		// increment the wait group
		wg.Add(1)

		// (!) In Go Versions pre 1.22
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
			item, err := memcachedClient.Get(myKey)
			if err != nil {
				// does this fall under the category of ignoring the error silently(?) and/or not handling it(?)
				fmt.Printf("🔴 memcached server %s | key: %s NOT FOUND\n", memcachedServer, myKey)
			} else {
				fmt.Printf("🟢 memcached server %s | key: %s FOUND with value: %vs\n", memcachedServer, myKey, string(item.Value))
				// if we find the key increment the count
				atomic.AddInt32(&memcachedServersWithKeyCount, 1)
			}
		}()
	}

	// wait for all (sub)goroutines to finish
	wg.Wait()

	// establish how many memcached servers were initially provided
	totalMemcachedServers := len(memcachedServers)

	// convert the int32 (necessary for atomic operations) to an int (janky!!)
	memcachedServersWithKeyCountInt := int(memcachedServersWithKeyCount)

	finish := time.Now()
	duration := finish.Sub(start)
	fmt.Printf("✅ topology scan completed in %v\n", duration)

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

// replicated

// $ go run . --mcrouter=localhost:11211 --memcacheds=localhost:11212,localhost:11213,localhost:11214
// 🟢 memcached server localhost:11213 | key: mykey FOUND with value: myvalues
// 🟢 memcached server localhost:11214 | key: mykey FOUND with value: myvalues
// 🟢 memcached server localhost:11212 | key: mykey FOUND with value: myvalues
// ✅ topology scan completed in 10.0357ms
// 🟣 Replicated topology: 3/3 memcached servers had the key

// sharded

// $ go run . --mcrouter=localhost:11211 --memcacheds=localhost:11212,localhost:11213,localhost:11214
// 🔴 memcached server localhost:11214 | key: mykey NOT FOUND
// 🟢 memcached server localhost:11212 | key: mykey FOUND with value: myvalues
// 🔴 memcached server localhost:11213 | key: mykey NOT FOUND
// ✅ topology scan completed in 10.5302ms
// 🟣 Sharded topology: 1/3 memcached servers had the key
