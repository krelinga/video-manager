package policy

import (
	"path/filepath"

	"github.com/google/uuid"
)

type Paths struct {
	RootDirPath string
}

func (p Paths) DiscPath(discUUID uuid.UUID) string {
	return filepath.Join(p.RootDirPath, "discs", discUUID.String())
}

func (p Paths) InboxPath(basename string) string {
	return filepath.Join(p.RootDirPath, "inbox", basename)
}