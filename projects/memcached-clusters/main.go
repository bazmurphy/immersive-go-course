package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/bradfitz/gomemcache/memcache"
)

func main() {
	// setup the flags for the command line tool
	mcRouterServerAddress := flag.String("mcrouter", "localhost:11211", "the mcrouter server address")
	memcachedServerAddresses := flag.String("memcacheds", "localhost:11212, localhost:11213, localhost:11214", "the list of memcached server addresses")

	// // parse the flags
	flag.Parse()

	// check the flags are working
	// fmt.Println("mcRouterServerAddress:", *mcRouterServerAddress)
	// fmt.Println("memcachedServerAddresses:", *memcachedServerAddresses)

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
	item, err := mcRouterClient.Get(myKey)
	if err != nil {
		// fmt.Printf("error: failed to read the key from the cache: %v\n", err)
		fmt.Println("mcRouterClient.Get err", err)
		os.Exit(1)
	}

	// ERROR:
	// memcache: unexpected line in get response: "SERVER_ERROR unexpected result mc_res_unknown (0) for get\r\n"

	// CHANGING LINE 389 in go/pkg/mod/github.com/bradfitz/gomemecache/memcache/memcache.go
	// CHANGE "gets" to "get"
	// BEFORE:
	// if _, err := fmt.Fprintf(rw, "gets %s\r\n", strings.Join(keys, " ")); err != nil {
	// AFTER:
	// if _, err := fmt.Fprintf(rw, "get %s\r\n", strings.Join(keys, " ")); err != nil {

	// check the get operation
	fmt.Printf("GET | key: %v value: %v\n", item.Key, string(item.Value))

	// breakup the string into individual memcached server addresses
	memcachedServers := strings.Split(*memcachedServerAddresses, ", ")

	// initialise a count
	memcachedServersWithKey := 0

	// loop over the memcached servers
	for _, memcachedServer := range memcachedServers {
		// make a new client for each memcached server
		memcachedClient := memcache.New(memcachedServer)

		// attempt to get the key from that specific memcached server
		item, err := memcachedClient.Get(myKey)
		if err != nil {
			fmt.Printf("key: %s NOT FOUND on memcached server %s\n", myKey, memcachedServer)
			// TOOD: this is silently ignoring the error yes, come back and fix this
			continue
		} else {
			fmt.Printf("key: %s FOUND with value: %v on memcached server %s \n", myKey, string(item.Value), memcachedServer)
			// if we find the key increment the count
			memcachedServersWithKey++
		}
	}

	totalMemcachedServers := len(memcachedServers)

	switch memcachedServersWithKey {
	case totalMemcachedServers:
		fmt.Printf("Replicated Topology: %d/%d memcached servers had the key\n", memcachedServersWithKey, totalMemcachedServers)
	case 1:
		fmt.Printf("Sharded Topology: %d/%d memcached server had the key", memcachedServersWithKey, totalMemcachedServers)
	default:
		fmt.Printf("Undetermined Topology: %d/%d memcached servers had the key", memcachedServersWithKey, totalMemcachedServers)
	}
}
