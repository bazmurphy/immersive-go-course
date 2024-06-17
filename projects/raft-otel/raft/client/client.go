package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"raft/raft_proto"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const port = 7600

func main() {
	addr := flag.String("dns", "raft", "dns address for raft cluster")

	if addr == nil || *addr == "" {
		fmt.Printf("Must supply dns address of cluster\n")
		os.Exit(1)
	}

	time.Sleep(time.Second * 5) // wait for raft servers to come up

	ips, err := net.LookupIP(*addr)
	if err != nil {
		fmt.Printf("Could not get IPs: %v\n", err)
		os.Exit(1)
	}

	clients := make([]raft_proto.RaftKVServiceClient, 0)

	for _, ip := range ips {
		fmt.Printf("Connecting to %s\n", ip.String())
		conn, err := grpc.Dial(fmt.Sprintf("%s:%d", ip.String(), port), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("%v", err)
		}
		client := raft_proto.NewRaftKVServiceClient(conn)
		clients = append(clients, client)
	}

	for {
		for i, c := range clients {
			n := time.Now().Second()
			setResponse, err := c.Set(context.TODO(), &raft_proto.SetRequest{
				Keyname: "current_second",
				Value:   fmt.Sprintf("%d", n),
			})
			if err != nil {
				log.Printf("CLIENT %d | SET Failure | Key: current_second | Error: %v\n", i+1, err)
			} else {
				log.Printf("CLIENT %d | SET Success | Key: current_second | Response: %v\n", i+1, setResponse)

				time.Sleep(1 * time.Second) // allow consensus to happen
			}

			getResponse, err := c.Get(context.TODO(), &raft_proto.GetRequest{
				Keyname: "current_second"},
			)
			if err != nil {
				log.Printf("CLIENT %d | GET Failure | Key: current_second | Error: %v\n", i+1, err)
			} else {
				log.Printf("CLIENT %d | GET Success | Key: current_second | Response: %v\n", i+1, getResponse)
			}

			// if the GET was successful, then we have an expectedValue to try a CAS with
			if getResponse != nil {
				n = time.Now().Second()

				casResponse, err := c.Cas(context.TODO(), &raft_proto.CasRequest{
					Keyname:       "current_second",
					ExpectedValue: getResponse.Value,
					NewValue:      fmt.Sprintf("%d", n)},
				)
				if err != nil {
					log.Printf("CLIENT %d | CAS Failure | Key: current_second | Error: %v \n", i+1, err)
					continue
				}
				log.Printf("CLIENT %d | CAS Success | Key: current_second | Response: %v\n", i+1, casResponse)

				time.Sleep(1 * time.Second) // allow consensus to happen
			}
		}
		time.Sleep(5 * time.Second)
	}
}
