package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	pb "github.com/CodeYourFuture/immersive-go-course/grpc-client-server/prober"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

// server is used to implement prober.ProberServer.
type server struct {
	pb.UnimplementedProberServer
}

func (s *server) DoProbes(ctx context.Context, in *pb.ProbeRequest) (*pb.ProbeReply, error) {
	// get the number of probes from the probe request
	numberOfProbes := in.GetNumberOfProbes()
	// fmt.Printf("DoProbes | numberOfProbes %v\n", numberOfProbes)

	// initialise a total time
	var totalTime time.Duration

	// support a number of repetitions and return average latency
	for i := 0; i < int(numberOfProbes); i++ {
		startTime := time.Now()
		// fmt.Printf("DoProbes | startTime %v\n", startTime)

		// make the request to the endpoint
		response, err := http.Get(in.GetEndpoint())

		// if the request errors
		if err != nil {
			return nil, fmt.Errorf("error: DoProbes request failed: %v", err)
		}

		// remember to close the response body
		defer response.Body.Close()

		// if the request status code is not OK
		if response.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("error: response was not OK: %v", err)
		}

		// get the elapsed time
		elapsedTime := time.Since(startTime)
		// fmt.Printf("DoProbes | elapsedTime %v\n", elapsedTime)

		// add the elapsed time to the total time
		totalTime += elapsedTime
		// fmt.Printf("DoProbes | totalTime %v\n", totalTime)
	}

	// calculate the average latency in milliseconds
	// (!) i am deliberately not using milliseconds here, so i can get more precision on the float, but why do i need to do this, this feels janky :/
	averageLatencyMsecs := float32(totalTime.Microseconds()) / float32(numberOfProbes) / 1000

	fmt.Printf("DoProbes | totalTime %v | numberOfProbes %d | averageLatencyMsecs %v\n", totalTime, numberOfProbes, averageLatencyMsecs)

	// return a probe reply
	return &pb.ProbeReply{AverageLatencyMsecs: averageLatencyMsecs}, nil
}

func main() {
	flag.Parse()
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterProberServer(grpcServer, &server{})
	log.Printf("server listening at %v", listener.Addr())
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
