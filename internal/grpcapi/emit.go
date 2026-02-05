package grpcapi

import (
	"context"
	"encoding/json"

	"celestialtree/internal/tree"
	pb "celestialtree/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

func (s *Server) Emit(ctx context.Context, req *pb.EmitRequest) (*pb.EmitResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "nil request")
	}

	// 把 google.protobuf.Struct 转成 JSON bytes，再塞给你现有的 store.Emit(req)
	var payload json.RawMessage
	if req.Payload != nil {
		b, err := protojson.Marshal(req.Payload)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid payload: %v", err)
		}
		// 注意：这里得到的是标准 JSON
		payload = json.RawMessage(b)
	}

	ev, err := s.store.Emit(tree.EmitRequest{
		Type:    req.Type,
		Message: req.Message,
		Payload: payload,
		Parents: req.Parents,
	})
	if err != nil {
		// 你 store.Emit 现在对参数问题会返回 fmt.Errorf(...)
		// 这里建议先粗分一下错误码（先跑通也可以统一 InvalidArgument）
		return nil, status.Errorf(codes.InvalidArgument, "emit failed: %v", err)
	}

	return &pb.EmitResponse{Id: ev.ID}, nil
}
