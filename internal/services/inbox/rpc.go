package inbox

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/krelinga/video-manager-api/go/vmapi"
	"github.com/krelinga/video-manager/internal/lib/config"
	"github.com/krelinga/video-manager/internal/lib/vmerr"
	"github.com/krelinga/video-manager/internal/lib/vmpage"
)

type InboxService struct {
	Config *config.Config
}

func (s *InboxService) ListInboxDVDs(ctx context.Context, request vmapi.ListInboxDVDsRequestObject) (vmapi.ListInboxDVDsResponseObject, error) {
	files, err := os.ReadDir(s.Config.DiscInboxDir)
	if err != nil {
		return nil, vmerr.InternalError(fmt.Errorf("could not read DVD inbox dir: %w", err))
	}
	var dirs []string
	for _, d := range files {
		if !d.IsDir() {
			continue
		}
		fullPath := filepath.Join(s.Config.DiscInboxDir, d.Name())
		dirs = append(dirs, fullPath)
	}
	limit := &vmpage.Limit{
		Want:    request.Params.PageSize,
		Default: 50,
		Max:     100,
	}
	dirs, token, err := vmpage.ListFromStrings(dirs, limit, request.Params.PageToken)
	if err != nil {
		return nil, err
	}
	resp := vmapi.ListInboxDVDs200JSONResponse{
		Paths:     dirs,
		NextPageToken: token,
	}

	return resp, nil
}
