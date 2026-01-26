package videoact

import (
	"context"
	"os"

	"github.com/google/uuid"
)

type DeleteDiscFileRequest struct {
	DiscUUID         uuid.UUID `json:"disc_uuid"`
	BasenameToDelete string    `json:"basename_to_delete"`
}

func (b *Basic) DeleteDiscFile(ctx context.Context, req *DeleteDiscFileRequest) error {
	path := b.Paths.FileInDiscPath(req.DiscUUID, req.BasenameToDelete)
	return os.Remove(path)
}
