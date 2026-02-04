package grpcapi

import (
	"celestialtree/internal/memory"
	pb "celestialtree/proto"
)

type Server struct {
	pb.UnimplementedCelestialTreeServiceServer
	store *memory.Store
}

func New(store *memory.Store) *Server {
	return &Server{store: store}
}
