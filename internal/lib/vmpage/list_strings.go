package vmpage

import (
	"errors"
	"fmt"
	"slices"
)

type Limit struct {
	Want    *uint32
	Default uint32
	Max     uint32
}

var ErrPanicBadLimit = errors.New("bad Limit")

func (l *Limit) Limit() uint32 {
	if l.Want != nil {
		if l.Max == 0 {
			panic(fmt.Errorf("%w: Max is zero", ErrPanicBadLimit))
		}
		return min(*l.Want, l.Max)
	}
	if l.Default == 0 {
		panic(fmt.Errorf("%w: Default is zero", ErrPanicBadLimit))
	}
	return l.Default
}

func ListFromStrings(strings []string, limit *Limit, pageToken *string) ([]string, *string, error) {
	limitValue := limit.Limit()
	lastSeenString, err := toLastSeenString(pageToken)
	if err != nil {
		return nil, nil, err
	}
	slices.Sort(strings)
	out := make([]string, 0, limitValue)
	for _, s := range strings {
		if s <= lastSeenString {
			continue
		}
		out = append(out, s)
		if uint32(len(out)) >= limitValue {
			break
		}
	}
	var nextPageToken *string
	if uint32(len(out)) == limitValue && len(out) > 0{
		token := fromLastSeenString(out[len(out)-1])
		nextPageToken = &token
	}
	return out, nextPageToken, nil
}
