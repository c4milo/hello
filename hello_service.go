package hello

import (
	"fmt"
	"sync"

	"github.com/c4milo/handlers/grpcutil"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type counter struct {
	// maps are not safe for concurrent access in Go, so we need to synchronize access with a mutex.
	sync.RWMutex
	m map[string]uint64
}

type service struct {
	counter *counter
}

func (s *service) SayHi(ctx context.Context, r *SayHiRequest) (*SayHiResponse, error) {
	if r.Name == "" {
		r.Name = "strange"
	}

	// If there is any concern for lock contention during high traffic, we can increase the
	// counter in a goroutine instead, since sending the greeting message does not rely on it.
	s.counter.Lock()
	defer s.counter.Unlock()

	name := r.Name
	s.counter.m[name]++

	return &SayHiResponse{
		Greeting: fmt.Sprintf("Hello, %s!", name),
	}, nil
}

func (s *service) Counts(ctx context.Context, r *empty.Empty) (*CountsResponse, error) {
	s.counter.RLock()
	defer s.counter.RUnlock()

	if len(s.counter.m) == 0 {
		// return nil, grpc.Errorf(codes.NotFound, "no visits recorded at this moment")
		err := new(Error)
		err.Status = uint32(codes.NotFound)
		err.Message = "Visits counter is zero"
		s := status.New(codes.NotFound, "no visits recorded")
		d, _ := s.WithDetails(err)
		return nil, d.Err()
	}

	res := new(CountsResponse)
	for k, v := range s.counter.m {
		res.Counts = append(res.Counts, &Count{
			Name:  k,
			Count: v,
		})
	}
	return res, nil
}

func (s *service) DeleteCounts(ctx context.Context, r *empty.Empty) (*empty.Empty, error) {
	// We lock the map just in case there are in-flight requests we don't want to lose.
	s.counter.Lock()
	defer s.counter.Unlock()

	s.counter = &counter{
		m: make(map[string]uint64),
	}

	return &empty.Empty{}, nil
}

// RegisterService registers service with a given GRPC server.
func RegisterService(binding grpcutil.ServiceBinding) error {
	// Creates a new service instance and injects gRPC clients for dependent services
	s := &service{
		counter: &counter{
			m: make(map[string]uint64),
		},
		// Any other service used within as well as test mocks can also be injected here if needed.
	}

	// Registers GRPC service.
	RegisterHelloServer(binding.GRPCServer, s)

	// Registers HTTP endpoint in GRPC Gateway Muxer. Enabling OpenAPI.
	return RegisterHelloHandler(context.Background(), binding.GRPCGatewayMuxer, binding.GRPCGatewayClient)
}
