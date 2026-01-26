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

func (p Paths) FileInDiscPath(discUUID uuid.UUID, basename string) string {
	return filepath.Join(p.DiscPath(discUUID), basename)
}

func (p Paths) InboxPath(basename string) string {
	return filepath.Join(p.RootDirPath, "inbox", basename)
}

func (p Paths) DiscPreviewPath(discUUID uuid.UUID) string {
	return filepath.Join(p.RootDirPath, "disc_previews", p.DiscPath(discUUID))
}

func (p Paths) FileInDiscPreviewPath(discUUID uuid.UUID, basename string) string {
	return filepath.Join(p.DiscPreviewPath(discUUID), basename)
}
