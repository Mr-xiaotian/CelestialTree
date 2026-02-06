package grpcapi

import (
	"github.com/Mr-xiaotian/CelestialTree/internal/memory"
	pb "github.com/Mr-xiaotian/CelestialTree/proto"
)

type Server struct {
	pb.UnimplementedCelestialTreeServiceServer
	store *memory.Store
}

func New(store *memory.Store) *Server {
	return &Server{store: store}
}
