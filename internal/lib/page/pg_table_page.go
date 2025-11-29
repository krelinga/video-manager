package page

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type lastSeenIdPage struct {
	MagicNumber uint32 `json:"magic_number"`
	LastSeenId  uint32 `json:"last_seen_id"`
}

const lastSeenIdPageMagicNumber uint32 = 3734103234

func FromLastSeenId(lastSeenId uint32) string {
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

func ToLastSeenId(pageStr string) (uint32, error) {
	if pageStr == "" {
		return 0, nil
	}
	pageBytes, err := base64.StdEncoding.DecodeString(pageStr)
	if err != nil {
		return 0, fmt.Errorf("%w: could not decode base64 data: %w", ErrUnmarshal, err)
	}
	var page lastSeenIdPage
	if err := json.Unmarshal(pageBytes, &page); err != nil {
		return 0, fmt.Errorf("%w: could not decode json data: %w", ErrUnmarshal, err)
	}
	if page.MagicNumber != lastSeenIdPageMagicNumber {
		return 0, fmt.Errorf("%w: invalid magic number", ErrUnmarshal)
	}
	return page.LastSeenId, nil
}
