// Package main implements a client for Prober service.
package main

import (
	"context"
	"flag"
	"log"

	pb "github.com/CodeYourFuture/immersive-go-course/grpc-client-server/prober"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	address        = flag.String("address", "localhost:50051", "the address to connect to")
	endpoint       = flag.String("endpoint", "http://google.com", "the endpoint to probe")
	numberOfProbes = flag.Int("probes", 1, "the number of requests to make")
)

func main() {
	flag.Parse()

	// Set up a connection to the server.
	conn, err := grpc.Dial(*address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewProberClient(conn)

	// Contact the server and print out its response.
	ctx := context.Background() // TODO: add a timeout

	r, err := c.DoProbes(ctx, &pb.ProbeRequest{Endpoint: *endpoint, NumberOfProbes: int32(*numberOfProbes)})
	if err != nil {
		log.Fatalf("could not probe: %v", err)
	}
	log.Printf("Average Response Time: %f", r.GetAverageLatencyMsecs())
}
