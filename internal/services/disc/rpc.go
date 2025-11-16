package disc

import (
	"context"
	"net/http"
	"os"

	"buf.build/gen/go/krelinga/proto/connectrpc/go/krelinga/video_manager/disc/v1/discv1connect"
	discv1 "buf.build/gen/go/krelinga/proto/protocolbuffers/go/krelinga/video_manager/disc/v1"
	"connectrpc.com/connect"
	"github.com/krelinga/video-manager/internal/lib/config"
)

type DiscServiceHandler struct {
	Config *config.Config
}

func (s *DiscServiceHandler) ListInbox(ctx context.Context, req *connect.Request[discv1.ListInboxRequest]) (*connect.Response[discv1.ListInboxResponse], error) {
	entries, err := os.ReadDir(s.Config.DiscInboxDir)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}
	return connect.NewResponse(&discv1.ListInboxResponse{
		Directories: dirs,
	}), nil
}

func NewServiceHandler(handler discv1connect.DiscServiceHandler) (string, http.Handler) {
	return discv1connect.NewDiscServiceHandler(handler)
}
