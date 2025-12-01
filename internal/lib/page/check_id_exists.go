package page

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
)

func CheckIdExists(ctx context.Context, qr QueryRower, table string, id uint32) error {
	const query = "SELECT COUNT(*) FROM $1 WHERE id = $2"
	var count int
	err := qr.QueryRow(ctx, query, table, id).Scan(&count)
	if err != nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to check if id exists in %q: %w", table, err))
	}
	if count == 0 {
		return connect.NewError(connect.CodeNotFound, fmt.Errorf("id %d does not exist", id))
	}
	return nil
}