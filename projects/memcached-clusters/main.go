package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/bradfitz/gomemcache/memcache"
)

func main() {
	// setup the flags for the command line tool
	mcRouterAddress := flag.String("mcrouter", "localhost:11211", "the mcrouter address")
	memcachedAddresses := flag.String("memcacheds", "localhost:11212, localhost:11213, localhost:11214", "the list of memcached addresses")
	// parse the flags
	flag.Parse()

	// check the flags are working
	fmt.Println(*mcRouterAddress)
	fmt.Println(*memcachedAddresses)

	// make a mcrouter Client
	mcRouterClient := memcache.New(*mcRouterAddress)

	// make a key and value to test later
	myKey := "bazkey"
	myValue := "bazvalue"
	myMemcacheItem := &memcache.Item{Key: myKey, Value: []byte(myValue)}

	// set the key
	err := mcRouterClient.Set(myMemcacheItem)
	if err != nil {
		fmt.Println("error: failed to write the item into the cache")
		os.Exit(1)
	}
}
