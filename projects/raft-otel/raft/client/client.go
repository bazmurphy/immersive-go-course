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
		for i, client := range clients {
			n := time.Now().Second()

			setResponse, err := setWithRetries(client, "current_second", fmt.Sprintf("%d", n), 3)
			if err != nil {
				log.Printf("CLIENT %d | SET Failure | Key: current_second | Response: %v | Error: %v\n", i+1, setResponse, err)
				continue
			}
			log.Printf("CLIENT %d | SET Success | Key: current_second | Response: %v\n", i+1, setResponse)

			time.Sleep(1 * time.Second) // allow consensus to happen

			getResponse, err := getWithRetries(client, "current_second", 3)
			if err != nil {
				log.Printf("CLIENT %d | GET Failure | Key: current_second | Response: %v | Error: %v\n", i+1, getResponse, err)
				continue
			}
			log.Printf("CLIENT %d | GET Success | Key: current_second | Response: %v\n", i+1, getResponse)

			// if we have a GET value, we can then make a CAS with it
			if getResponse.Value != "" {
				n = time.Now().Second()
				casResponse, err := casWithRetries(client, "current_second", getResponse.Value, fmt.Sprintf("%d", n), 3)
				if err != nil {
					log.Printf("CLIENT %d | CAS Failure | Key: current_second | Response: %v | Error: %v\n", i+1, casResponse, err)
					continue
				}
				log.Printf("CLIENT %d | CAS Success | Key: current_second | Response: %v\n", i+1, casResponse)

				time.Sleep(1 * time.Second) // allow consensus to happen
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func setWithRetries(client raft_proto.RaftKVServiceClient, keyname, value string, retries int) (*raft_proto.SetResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	setRequest := &raft_proto.SetRequest{
		Keyname: keyname,
		Value:   value,
	}

	setResponse, err := client.Set(ctx, setRequest)
	if err != nil {
		return nil, fmt.Errorf("SET FAILED: %v", err)
	}

	if setResponse.LeaderAddress != "" {
		if retries > 0 {
			newConn, err := grpc.Dial(fmt.Sprintf("%s:%d", setResponse.LeaderAddress, port), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				return nil, err
			}
			client = raft_proto.NewRaftKVServiceClient(newConn)
			return setWithRetries(client, keyname, value, retries-1)
		} else {
			return nil, fmt.Errorf("SET FAILED: out of retries")
		}
	}

	return setResponse, nil
}

func getWithRetries(client raft_proto.RaftKVServiceClient, keyname string, retries int) (*raft_proto.GetResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	getRequest := &raft_proto.GetRequest{
		Keyname: keyname,
	}

	getResponse, err := client.Get(ctx, getRequest)
	if err != nil {
		return nil, fmt.Errorf("GET FAILED: %v", err)
	}

	if getResponse.LeaderAddress != "" {
		if retries > 0 {
			newConn, err := grpc.Dial(fmt.Sprintf("%s:%d", getResponse.LeaderAddress, port), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				return nil, err
			}
			client = raft_proto.NewRaftKVServiceClient(newConn)
			return getWithRetries(client, keyname, retries-1)
		} else {
			return nil, fmt.Errorf("GET FAILED: out of retries")
		}
	}

	return getResponse, nil
}

func casWithRetries(client raft_proto.RaftKVServiceClient, keyname, expectedValue, newValue string, retries int) (*raft_proto.CasResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	casRequest := &raft_proto.CasRequest{
		Keyname:       keyname,
		ExpectedValue: expectedValue,
		NewValue:      newValue,
	}
	casResponse, err := client.Cas(ctx, casRequest)
	if err != nil {
		return nil, fmt.Errorf("CAS FAILED: %v", err)
	}

	if casResponse.LeaderAddress != "" {
		if retries > 0 {
			newConn, err := grpc.Dial(fmt.Sprintf("%s:%d", casResponse.LeaderAddress, port), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				return nil, err
			}
			client = raft_proto.NewRaftKVServiceClient(newConn)
			return casWithRetries(client, keyname, expectedValue, newValue, retries-1)
		} else {
			return nil, fmt.Errorf("CAS FAILED: out of retries")
		}
	}

	return casResponse, nil
}
