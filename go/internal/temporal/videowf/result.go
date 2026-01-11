package videowf

import "encoding/json"

type Result[T any] struct {
	Value T
	Error string
	Ready bool
}

func (r *Result[T]) Set(value T, err error) {
	r.Value = value
	if err != nil {
		r.Error = err.Error()
		if r.Error == "" {
			r.Error = "<empty error>"
		}
	} else {
		r.Error = ""
	}
	r.Ready = true
}

func (r *Result[T]) IsZero() bool {
	return !r.Ready
}

func (r *Result[T]) MarshalJSON() ([]byte, error) {
	if !r.Ready {
		panic("cannot marshal zero Result")
	}
	if r.Error != "" {
		type errorOnly struct {
			Error string `json:"error"`
		}
		return json.Marshal(&errorOnly{Error: r.Error})
	}
	type valueOnly struct {
		Value T `json:"value"`
	}
	return json.Marshal(&valueOnly{Value: r.Value})
}

func (r *Result[T]) UnmarshalJSON(data []byte) error {
	type alias struct {
		Value T     `json:"value"`
		Error string `json:"error"`
	}
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	r.Value = a.Value
	r.Error = a.Error
	r.Ready = true
	return nil
}

func transformResult[A any, B any](input Result[A], transform func(A) B) Result[B] {
	return Result[B]{
		Value: transform(input.Value),
		Error: input.Error,
		Ready: input.Ready,
	}
}
