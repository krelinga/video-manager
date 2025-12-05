package inbox

import (
	"context"

	"github.com/krelinga/video-manager-api/go/vmapi"
	"github.com/krelinga/video-manager/internal/lib/config"
)

type InboxService struct {
	Config *config.Config
}

func (s *InboxService) ListInboxDVDs(ctx context.Context, request vmapi.ListInboxDVDsRequestObject) (vmapi.ListInboxDVDsResponseObject, error) {
	return nil, nil // TODO: implement
}
