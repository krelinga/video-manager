package vmpage

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"connectrpc.com/connect"
)

var ErrPanicTokenMarshall = errors.New("failed to marshall page token")

var ErrBadPageToken = connect.NewError(connect.CodeInvalidArgument, errors.New("failed to read page token"))

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
		panic(fmt.Errorf("%w: %w", ErrPanicTokenMarshall, err))
	}
	return base64.StdEncoding.EncodeToString(pageBytes)
}

func toLastSeenId(pageStr *string) (uint32, error) {
	if pageStr == nil {
		return 0, nil
	}
	pageBytes, err := base64.StdEncoding.DecodeString(*pageStr)
	if err != nil {
		return 0, fmt.Errorf("%w: could not decode base64 data: %w", ErrBadPageToken, err)
	}
	var page lastSeenIdPage
	if err := json.Unmarshal(pageBytes, &page); err != nil {
		return 0, fmt.Errorf("%w: could not decode json data: %w", ErrBadPageToken, err)
	}
	if page.MagicNumber != lastSeenIdPageMagicNumber {
		return 0, fmt.Errorf("%w: invalid magic number", ErrBadPageToken)
	}
	return page.LastSeenId, nil
}