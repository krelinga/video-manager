package videowf

import "encoding/json"

type Result[T any] struct {
	Value T
	Error error
	Ready bool
}

func (r *Result[T]) Set(value T, err error) {
	r.Value = value
	r.Error = err
	r.Ready = true
}

func (r *Result[T]) IsZero() bool {
	return !r.Ready
}

func (r *Result[T]) MarshalJSON() ([]byte, error) {
	if !r.Ready {
		panic("cannot marshal zero Result")
	}
	if r.Error != nil {
		errorString := r.Error.Error()
		if errorString == "" {
			errorString = "<empty error>"
		}
		type errorOnly struct {
			Error string `json:"error"`
		}
		return json.Marshal(&errorOnly{Error: errorString})
	}
	type valueOnly struct {
		Value T `json:"value"`
	}
	return json.Marshal(&valueOnly{Value: r.Value})
}

func (r *Result[T]) UnmarshalJSON(data []byte) error {
	type alias Result[T]
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*r = Result[T](a)
	r.Ready = true
	return nil
}
