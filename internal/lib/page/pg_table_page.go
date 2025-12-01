package page

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"connectrpc.com/connect"
)

type lastSeenIdPage struct {
	MagicNumber uint32 `json:"magic_number"`
	LastSeenId  uint32 `json:"last_seen_id"`
}

const lastSeenIdPageMagicNumber uint32 = 3734103234

func fromLastSeenId(lastSeenId uint32) string {
	page := &lastSeenIdPage{
		MagicNumber: lastSeenIdPageMagicNumber,
		LastSeenId:  lastSeenId,
	}
	pageBytes, err := json.Marshal(page)
	if err != nil {
		panic(fmt.Errorf("%w: %w", ErrMarshal, err))
	}
	return base64.StdEncoding.EncodeToString(pageBytes)
}

func toLastSeenId(pageStr *string) (uint32, error) {
	if pageStr == nil {
		return 0, nil
	}
	pageBytes, err := base64.StdEncoding.DecodeString(*pageStr)
	if err != nil {
		return 0, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("could not decode base64 data: %w", err))
	}
	var page lastSeenIdPage
	if err := json.Unmarshal(pageBytes, &page); err != nil {
		return 0, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("could not decode json data: %w", err))
	}
	if page.MagicNumber != lastSeenIdPageMagicNumber {
		return 0, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid magic number"))
	}
	return page.LastSeenId, nil
}
