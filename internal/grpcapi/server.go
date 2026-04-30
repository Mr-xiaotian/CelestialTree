package grpcapi

import (
	"github.com/Mr-xiaotian/CelestialTree/internal/memory"
	pb "github.com/Mr-xiaotian/CelestialTree/proto"
)

// Server 是 gRPC 服务的实现，持有底层 Store 引用。
type Server struct {
	pb.UnimplementedCelestialTreeServiceServer
	store *memory.Store
}

// New 创建一个绑定到指定 Store 的 gRPC Server 实例。
func New(store *memory.Store) *Server {
	return &Server{store: store}
}
