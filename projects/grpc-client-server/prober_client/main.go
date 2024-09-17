// Package main implements a client for Prober service.
package main

import (
	"context"
	"flag"
	"log"
	"time"

	pb "github.com/CodeYourFuture/immersive-go-course/grpc-client-server/prober"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

var (
	address         = flag.String("address", "localhost:50051", "the address to connect to")
	endpoint        = flag.String("endpoint", "http://google.com", "the endpoint to probe")
	numberOfProbes  = flag.Int("probes", 1, "the number of requests to make")
	timeoutDuration = flag.Duration("timeout", 1*time.Second, "the timeout duration")
)

func main() {
	// parse the flags
	flag.Parse()

	// create a connection the gRPC server (with insecure credentials)
	connection, err := grpc.Dial(*address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer connection.Close()

	// create a new client instance for Prober service using the connection above
	client := pb.NewProberClient(connection)

	// create a context with a timeout duration
	ctx, cancel := context.WithTimeout(context.Background(), *timeoutDuration)
	defer cancel()

	// call the DoProbes method on the Prober service (with the endpoint and number of probes) and receive a response
	response, err := client.DoProbes(ctx, &pb.ProbeRequest{Endpoint: *endpoint, NumberOfProbes: int32(*numberOfProbes)})
	if err != nil {
		// (!!!) the error returned by c.DoProbes() is not exactly of type context.DeadlineExceeded, but rather a gRPC error that wraps the deadline exceeded error.
		// To check if the error is a deadline exceeded error, you can use the status package from the google.golang.org/grpc/status module
		if status.Code(err) == codes.DeadlineExceeded {
			log.Fatalf("Probing exceeded the %v deadline", timeoutDuration)
		} else {
			log.Fatalf("Could not probe: %v", err)
		}
	}
	log.Printf("Average Response Time: %f ms (%d Requests)", response.GetAverageLatencyMsecs(), *numberOfProbes)
}
