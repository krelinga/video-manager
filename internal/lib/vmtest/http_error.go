package vmtest

import (
	"errors"
	"fmt"

	"github.com/krelinga/go-libs/deep"
	"github.com/krelinga/go-libs/match"

	"github.com/krelinga/video-manager/internal/lib/vmerr"
)

func HttpError(code int) match.Matcher {
	return match.Func(func(env deep.Env, vals match.Vals) match.Result {
		got := match.Want1[error](vals)
		var httpErr *vmerr.HttpError
		if !errors.As(got, &httpErr) {
			text := "expected vmerr.HttpError, got different error type"
			return match.NewResult(false, text)
		}
		if httpErr.StatusCode != code {
			text := "expected HTTP code %d, got %d"
			text = fmt.Sprintf(text, code, httpErr.StatusCode)
			return match.NewResult(false, text)
		}
		return match.NewResult(true, "")
	})
}
