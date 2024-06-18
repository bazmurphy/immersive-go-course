// Server container for a Raft Consensus Module. Exposes Raft to the network
// and enables RPCs between Raft peers.
//
// Based on code by Eli Bendersky [https://eli.thegreenplace.net], modified somewhat by Laura Nolan.
// This code is in the public domain.
package raft

import (
	"context"
	"fmt"
	"log"
	"net"
	"raft/raft_proto"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// Server wraps a raft.ConsensusModule along with a rpc.Server that exposes its
// methods as gRPC endpoints. It also manages the peers of the Raft server. The
// main goal of this type is to simplify the code of raft.Server for
// presentation purposes. raft.ConsensusModule has a *Server to do its peer
// communication and doesn't have to worry about the specifics of running an
// RPC server.
type Server struct {
	// mutex for controlling concurrent access
	mu sync.Mutex

	// unique identifier for the server within the raft cluster
	serverId string
	// ip address of the server
	ip string

	// pointer to the consensus module instance (that implements the core raft consensus algorithm)
	cm *ConsensusModule
	// implementation of the Storage interface that provides persistent storage for the raft log and server state
	// used by the consensus module to persist its internal state: current term, voted for, and raft log
	// (!) this is an internal component of raft and is not directly accessed by clients
	storage Storage

	// listens for incoming network connections
	listener net.Listener
	// port number on which the server listens for connections
	listenPort int
	// pointer to the grpc instance which serves the grpc endpoints for the raft and kv services
	grpcServer *grpc.Server

	// a channel used to communicate committed log entries from the consensus module to the finite state machine (fsm) (key-value store)
	commitChan chan CommitEntry
	// a map that stores grpc client connections to other raft servers (nodes/peers) in the cluster
	peerClients map[string]raft_proto.RaftServiceClient

	// a channel that indicates when the server is ready to start serving requests
	ready <-chan interface{}
	// a channel used to signal the server to shut down gracefully
	quit chan interface{}

	raft_proto.UnimplementedRaftServiceServer
	raft_proto.UnimplementedRaftKVServiceServer

	// a pointer to the KV instance which represents the finite state machine (key-value store) that is managed by raft
	// (!) this is the key-value store that the Clients interact with
	fsm *KV
}

// represents a key-value store and has a map vals to store key-value pairs
type KV struct {
	vals map[string]string
}

// constructor that creates a new instance of KV with an initialized vals map
func NewKV() *KV {
	return &KV{vals: make(map[string]string)}
}

// constructor that creates a new instance of Server
// using the provided parameters: serverId, ip, storage, ready, commitChan, listenPort
// then initializes various fields of the Server struct
// returns the new instance of Server
func NewServer(serverId string, ip string, storage Storage, ready <-chan interface{}, commitChan chan CommitEntry, listenPort int) *Server {
	s := new(Server)
	s.serverId = serverId
	s.ip = ip
	s.peerClients = make(map[string]raft_proto.RaftServiceClient)
	s.storage = storage
	s.ready = ready
	s.commitChan = commitChan
	s.quit = make(chan interface{})
	s.listenPort = listenPort
	return s
}

// the Serve method (of Server):
// - creates a new consensus module
// - listens on the specified IP and TCP port
// - creates a new gRPC server
// - registers the Raft and KV services with the gRPC server
// - starts the gRPC server and serves on that TPC port (above)
// - then runs the FSM (if provided) in a separate goroutine to read commits from the commitChan

// If fsm is set then commitChan is read by the FSM, otherwise commitChan can be read by tests
// Bit icky. Oh well.
func (s *Server) Serve(fsm *KV) {
	s.mu.Lock()
	s.cm = NewConsensusModule(s.serverId, s, s.storage, s.ready, s.commitChan)
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.ip, s.listenPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s.listener = lis

	gs := grpc.NewServer()
	s.grpcServer = gs

	raft_proto.RegisterRaftServiceServer(gs, s)
	raft_proto.RegisterRaftKVServiceServer(gs, s)

	go gs.Serve(s.listener)
	if fsm != nil {
		s.fsm = fsm
		// TODO should close this goroutine on shutdown
		go s.fsm.readCommits(s.commitChan)
	}

	log.Printf("[%v] listening at %s", s.serverId, s.listener.Addr())

	s.mu.Unlock()
}

// closes all the client connections to peers for this server
func (s *Server) DisconnectAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id := range s.peerClients {
		if s.peerClients[id] != nil {
			s.peerClients[id] = nil
		}
	}
}

// closes the server and waits for it to shut down properly
func (s *Server) Shutdown() {
	s.grpcServer.GracefulStop()
	s.cm.Stop()
	close(s.quit)
	s.listener.Close()
}

// returns the network address on which the server is listening
func (s *Server) GetListenAddr() net.Addr {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.listener.Addr()
}

// establishes a connection to a peer (identified by peerId and peerAddr)
// - creates a new gRPC client for the peer if it doesn't exist
// - adds the peer ID to the Consensus Module
func (s *Server) ConnectToPeer(peerId string, peerAddr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.peerClients[peerId] == nil {
		conn, err := grpc.Dial(peerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return err
		}
		client := raft_proto.NewRaftServiceClient(conn)
		s.peerClients[peerId] = client
	}
	s.cm.AddPeerID(peerId)
	return nil
}

// disconnects this server from the peer (identified by peerId)
func (s *Server) DisconnectPeer(peerId string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.peerClients[peerId] != nil {
		s.peerClients[peerId] = nil
		return nil
	}
	return nil
}

// sends a RequestVote RPC to a peer (identified by id)
// - gets the peer client by id
// - constructs a RequestVoteRequest message (using the args)
// - sends the message to the peer using the peer client
// - populates the RequestVoteReply with the response
func (s *Server) CallRequestVote(id string, args RequestVoteArgs, reply *RequestVoteReply) error {
	s.mu.Lock()
	peer := s.peerClients[id]
	s.mu.Unlock()

	// If this is called after shutdown (where client.Close is called), it will
	// return an error.
	if peer == nil {
		return fmt.Errorf("call client %s after it's closed", id)
	} else {
		req := raft_proto.RequestVoteRequest{
			Term:         int64(args.Term),
			CandidateId:  args.CandidateId,
			LastLogIndex: int64(args.LastLogIndex),
			LastLogTerm:  int64(args.LastLogTerm),
		}
		resp, err := peer.RequestVote(context.TODO(), &req)
		if err != nil {
			return err
		}

		reply.Term = int(resp.GetTerm())
		reply.VoteGranted = resp.GetVoteGranted()
	}
	return nil
}

// the gRPC handler for the RequestVote RPC
// - receives a RequestVoteRequest
// - converts it to RequestVoteArgs
// - calls the RequestVote method of the consensus module
// - returns a RequestVoteResponse
func (s *Server) RequestVote(ctx context.Context, req *raft_proto.RequestVoteRequest) (*raft_proto.RequestVoteResponse, error) {
	fmt.Printf("[%s] received RequestVote %+v\n", s.serverId, req)

	rva := RequestVoteArgs{
		Term:         int(req.GetTerm()),
		CandidateId:  req.GetCandidateId(),
		LastLogIndex: int(req.GetLastLogIndex()),
		LastLogTerm:  int(req.GetLastLogTerm()),
	}

	rvr := RequestVoteReply{}
	err := s.cm.RequestVote(rva, &rvr)
	if err != nil {
		return nil, err
	}

	resp := raft_proto.RequestVoteResponse{
		Term:        int64(rvr.Term),
		VoteGranted: rvr.VoteGranted,
	}
	return &resp, nil
}

// sends an AppendEntries RPC to a peer (identified by id)
// - gets the peer client by id
// - constructs a AppendEntriesRequest message (using the args)
// - sends the message to the peer using the peer client
// - populates the AppendEntriesReply with the response
func (s *Server) CallAppendEntries(id string, args AppendEntriesArgs, reply *AppendEntriesReply) error {
	s.mu.Lock()
	peer := s.peerClients[id]
	s.mu.Unlock()

	// If this is called after shutdown (where client.Close is called), it will
	// return an error.
	if peer == nil {
		return fmt.Errorf("call client %s after it's closed", id)
	} else {
		req := raft_proto.AppendEntriesRequest{
			Term:         int64(args.Term),
			Leader:       args.LeaderId,
			PrevLogIndex: int64(args.PrevLogIndex),
			PrevLogTerm:  int64(args.PrevLogTerm),
			LeaderCommit: int64(args.LeaderCommit),
			Entries:      make([]*raft_proto.LogEntry, 0),
		}

		for _, e := range args.Entries {
			en := raft_proto.LogEntry{
				Term:    int64(e.Term),
				Command: &raft_proto.Command{Command: e.Command.Command, Args: e.Command.Args},
			}
			req.Entries = append(req.Entries, &en)
		}

		resp, err := peer.AppendEntries(context.TODO(), &req)
		if err != nil {
			return err
		}

		reply.ConflictIndex = int(resp.GetConflictIndex())
		reply.ConflictTerm = int(resp.GetConflictTerm())
		reply.Success = resp.GetSuccess()
		reply.Term = int(resp.GetTerm())
	}
	return nil
}

// the gRPC handler for the AppendEntries RPC
// - receives an AppendEntriesRequest
// - converts the request to AppendEntriesArgs
// - appends all request Entries to the AppendEntriesArgs Entries log
// - calls the AppendEntries method of the Consensus Module (passing it the AppendEntriesArgs)
// - constructs and returns an AppendEntriesResponse
func (s *Server) AppendEntries(ctx context.Context, req *raft_proto.AppendEntriesRequest) (*raft_proto.AppendEntriesResponse, error) {
	aea := AppendEntriesArgs{
		Term:         int(req.GetTerm()),
		LeaderId:     req.GetLeader(),
		PrevLogIndex: int(req.GetPrevLogIndex()),
		PrevLogTerm:  int(req.GetPrevLogTerm()),
		Entries:      make([]LogEntry, 0),
		LeaderCommit: int(req.GetLeaderCommit()),
	}
	for _, e := range req.Entries {
		en := LogEntry{
			Term:    int(e.Term),
			Command: CommandImpl{Command: e.Command.Command, Args: e.Command.Args},
		}
		aea.Entries = append(aea.Entries, en)
	}

	aer := AppendEntriesReply{}
	err := s.cm.AppendEntries(aea, &aer)
	if err != nil {
		return nil, err
	}

	resp := raft_proto.AppendEntriesResponse{
		Term:          int64(aer.Term),
		Success:       aer.Success,
		ConflictIndex: int64(aer.ConflictIndex),
		ConflictTerm:  int64(aer.ConflictTerm),
	}

	return &resp, nil
}

// the gRPC handler for the Set RPC
// - receives a SetRequest
// - constructs a CommandImpl with the "set" command and the provided key and value
// - submits the command to the consensus module
// - if there is an error, returns a gRPC error status code
// - returns a SetResponse (empty denoting success?)
func (s *Server) Set(ctx context.Context, req *raft_proto.SetRequest) (*raft_proto.SetResponse, error) {
	// TODO proxy to leader if not leader
	if s.cm.state != Leader {
		return &raft_proto.SetResponse{LeaderAddress: s.currentLeaderAddress()}, nil
	}

	cmd := CommandImpl{
		Command: "set",
		Args:    []string{req.Keyname, req.Value},
	}

	res := s.cm.Submit(cmd)
	if !res {
		return nil, status.Error(codes.Unavailable, "not the leader")
	}

	return &raft_proto.SetResponse{}, nil
}

// the gRPC handler for the Get RPC
// - receives a GetRequest
// - checks if the server is the leader, if not returns an empty GetResponse response, and gRPC error status code
// - retrieves the value associated with the provided key from the key-value store (fsm)
// - returns a GetResponse
func (s *Server) Get(ctx context.Context, req *raft_proto.GetRequest) (*raft_proto.GetResponse, error) {
	// TODO allow gets from non-leader if the query specified
	if s.cm.state != Leader {
		return &raft_proto.GetResponse{LeaderAddress: s.currentLeaderAddress()}, nil
	}

	res := s.fsm.get(req.Keyname)
	return &raft_proto.GetResponse{Value: res}, nil
}

// the gRPC handler for the Cas RPC
// - receives a CasRequest
// - constructs a CommandImpl with the "cas" command and the provided key, expected value, and new value
// - submits the command to the consensus module
// - if there is an error, returns a gRPC error status code
// - returns a CasResponse
func (s *Server) Cas(ctx context.Context, req *raft_proto.CasRequest) (*raft_proto.CasResponse, error) {
	if s.cm.state != Leader {
		return &raft_proto.CasResponse{LeaderAddress: s.currentLeaderAddress()}, nil
	}

	cmd := CommandImpl{
		Command: "cas",
		Args:    []string{req.Keyname, req.ExpectedValue, req.NewValue},
	}

	res := s.cm.Submit(cmd)
	if !res {
		return nil, status.Error(codes.Unavailable, "not the leader")
	}

	return &raft_proto.CasResponse{}, nil
}

func (s *Server) currentLeaderAddress() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cm.leaderId
}

// reads commits from the commitChan
// applies the "set" or "cas" command to the key-value store
func (kv *KV) readCommits(ch chan CommitEntry) {
	for {
		entry := <-ch
		if entry.Command.Command == "set" {
			if len(entry.Command.Args) != 2 {
				log.Printf("Can't parse this set command %+v", entry.Command)
			}
			kn := entry.Command.Args[0]
			val := entry.Command.Args[1]
			kv.set(kn, val)
		} else if entry.Command.Command == "cas" {
			if len(entry.Command.Args) != 3 {
				log.Printf("Can't parse this cas command %+v", entry.Command)
			}
			kn := entry.Command.Args[0]
			expectedValue := entry.Command.Args[1]
			newValue := entry.Command.Args[2]
			kv.cas(kn, expectedValue, newValue)
		}
	}
}

// sets a key value pair in the KV vals map
func (kv *KV) set(k string, v string) {
	kv.vals[k] = v
}

// retrieves the value associated with the provided key from the KV vals map
func (kv *KV) get(k string) string {
	return kv.vals[k]
}

// checks if the key value pair is the expected value, if yes then updates the key in the KV vals map with a new value
func (kv *KV) cas(k string, expectedValue string, newValue string) {
	if kv.vals[k] == expectedValue {
		kv.vals[k] = newValue
	}
}
