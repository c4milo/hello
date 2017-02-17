package hello

import (
	"fmt"
	"sync"

	"github.com/c4milo/handlers/grpcutil"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type counter struct {
	// maps are not safe for concurrent access in Go, so we need to synchronize them with a mutex.
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
	// counter in a goroutine instead since sending the greeting message does not rely on it.
	s.counter.Lock()
	defer s.counter.Unlock()
	name := r.Name

	s.counter.m[name]++

	return &SayHiResponse{
		Greeting: fmt.Sprintf("Hello, %s!", r.Name),
	}, nil
}

func (s *service) Health(ctx context.Context, r *empty.Empty) (*HealthResponse, error) {
	return nil, grpc.Errorf(codes.Unimplemented, "nothing to see here")
}

func (s *service) Counts(ctx context.Context, r *empty.Empty) (*CountsResponse, error) {
	s.counter.RLock()
	defer s.counter.RUnlock()

	if len(s.counter.m) == 0 {
		return nil, grpc.Errorf(codes.NotFound, "there is no visits recorded at this moment")
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
	// We lock the map just in case there is an in-flight request that we don't want to lose.
	s.counter.Lock()
	defer s.counter.Unlock()

	// previous map allocated memory is freed by Go's garbage collector.
	s.counter = &counter{
		m: make(map[string]uint64),
	}

	return &empty.Empty{}, nil
}

// RegisterService registers service with a given GRPC server.
func RegisterService(binding grpcutil.ServiceBinding) error {
	// Creates a new service instance and injects gRPC clients for dependent services
	service := &service{
		counter: &counter{
			m: make(map[string]uint64),
		},
		// gRPC services used within this service's logic or test mocks can also
		// be injected here if needed.
	}

	// Registers GRPC service.
	RegisterHelloServer(binding.GRPCServer, service)

	// Registers HTTP endpoint in GRPC Gateway Muxer. Enabling OpenAPI.
	return RegisterHelloHandler(context.Background(), binding.GRPCGatewayMuxer, binding.GRPCGatewayClient)
}
