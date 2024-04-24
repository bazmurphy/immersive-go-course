package main

import (
	"flag"
	"fmt"

	"github.com/bradfitz/gomemcache/memcache"
)

func main() {
	// setup the flags for the command line tool
	mcRouterServerAddress := flag.String("mcrouter", "localhost:11211", "the mcrouter server address")
	memcachedServerAddresses := flag.String("memcacheds", "localhost:11212, localhost:11213, localhost:11214", "the list of memcached server addresses")
	// parse the flags
	flag.Parse()

	// check the flags are working
	fmt.Println("mcRouterServerAddress:", *mcRouterServerAddress)
	fmt.Println("memcachedServerAddresses:", *memcachedServerAddresses)

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
	err = mcRouterClient.Set(&memcache.Item{Key: myKey, Value: []byte(myValue)})
	if err != nil {
		fmt.Printf("error: failed to write the item into the cache: %v\n", err)
		// os.Exit(1)
	}

	// get the key
	item, err := mcRouterClient.Get(myKey)
	if err != nil {
		fmt.Printf("error: failed to read the key from the cache: %v\n", err)
		// os.Exit(1)
	}
	fmt.Println("item", item)
}
