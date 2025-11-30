package vmtest

import (
	"fmt"

	"connectrpc.com/connect"
	"github.com/krelinga/go-libs/deep"
	"github.com/krelinga/go-libs/match"
)

func ConnectCode(code connect.Code) match.Matcher {
	return match.Func(func(env deep.Env, vals match.Vals) match.Result {
		got := match.Want1[error](vals)
		gotCode := connect.CodeOf(got)
		if gotCode != code {
			text := fmt.Sprintf("expected connect code %v, got %v", code, gotCode)
			return match.NewResult(false, text)
		}
		return match.NewResult(true, "")
	})
}
