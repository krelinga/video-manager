package pgxcatalog

type patchOptField[T any] struct {
	value   *T
	changed bool
}

func (pf *patchOptField[T]) Set(value T) {
	pf.value = &value
	pf.changed = true
}

func (pf *patchOptField[T]) Clear() {
	pf.value = nil
	pf.changed = true
}

func (pf *patchOptField[T]) Changed() bool {
	return pf.changed
}

type patchReqField[T any] struct {
	value *T
}

func (pf *patchReqField[T]) Set(value T) {
	pf.value = &value
}

func (pf *patchReqField[T]) Changed() bool {
	return pf.value != nil
}

func (pf *patchReqField[T]) Get() T {
	return *pf.value
}

func (pf *patchOptField[T]) Get() *T {
	return pf.value
}
