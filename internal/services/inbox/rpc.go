package inbox

import "github.com/krelinga/video-manager/internal/lib/config"

type InboxService struct {
	Config *config.Config
}

func (s *InboxService) List