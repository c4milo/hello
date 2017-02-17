package hello

import (
	"testing"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/hooklift/assert"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func TestSayHi(t *testing.T) {
	s := &service{
		counter: &counter{
			m: make(map[string]uint64),
		},
	}

	tests := []struct {
		desc     string
		name     string
		res      *SayHiResponse
		expCount uint64
		err      error
	}{
		{
			"it defaults to greeting a strange when no name is provided",
			"", &SayHiResponse{Greeting: "Hello, strange!"}, 1, nil,
		},
		{
			"it greets using the name provided",
			"camilo", &SayHiResponse{Greeting: "Hello, camilo!"}, 1, nil,
		},
		{
			"it greets using the name provided",
			"camilo", &SayHiResponse{Greeting: "Hello, camilo!"}, 2, nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			req := &SayHiRequest{Name: tt.name}

			res, err := s.SayHi(context.Background(), req)
			if err != nil {
				assert.Equals(t, tt.err, err)
			}
			assert.Equals(t, tt.res, res)

			s.counter.RLock()
			defer s.counter.RUnlock()
			assert.Equals(t, tt.expCount, s.counter.m[req.Name])
		})
	}
}

func TestCounts(t *testing.T) {
	s := &service{
		counter: &counter{m: make(map[string]uint64)},
	}

	tests := []struct {
		desc    string
		counter map[string]uint64
		err     error
	}{
		{
			"it returns an error when no counts are found",
			map[string]uint64{}, grpc.Errorf(codes.NotFound, "there is no visits recorded at this moment"),
		},
		{
			"it returns counts successfully",
			map[string]uint64{"alice": 3, "camilo": 2}, nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			s.counter.m = tt.counter

			res, err := s.Counts(context.Background(), &empty.Empty{})
			if err != nil {
				assert.Equals(t, tt.err, err)
				assert.Equals(t, (*CountsResponse)(nil), res)
				return
			}

			assert.Cond(t, res != nil, "we were expecting a non-nil response")
			assert.Equals(t, len(tt.counter), len(res.Counts))

			for _, c := range res.Counts {
				assert.Equals(t, tt.counter[c.Name], c.Count)
			}
		})
	}

}

func TestDeleteCounts(t *testing.T) {
	s := &service{
		counter: &counter{
			m: map[string]uint64{
				"alice":  4,
				"camilo": 2,
			},
		},
	}

	_, err := s.DeleteCounts(context.Background(), &empty.Empty{})
	assert.Ok(t, err)
	assert.Equals(t, len(s.counter.m), 0)
}
