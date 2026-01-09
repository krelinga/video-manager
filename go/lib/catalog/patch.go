package catalog

// A patch operation that can be applied to values.
type ValPatcher[T any] interface {
	ValPatch(*T)
}

// A patch operation that can be applied to pointers.
// If the pointer is nil but the patch needs to set a value, then
// the patch will allocate a new value.
type PtrPatcher[T any] interface {
	PtrPatch(**T)
}

// A patch operation that can be applied to both values and pointers.
type Patcher[T any] interface {
	ValPatcher[T]
	PtrPatcher[T]
}

type patchClearPtr[T any] struct{}

func (p patchClearPtr[T]) PtrPatch(ptr **T) {
	*ptr = nil
}

// Clear returns a patch operation that sets a pointer to nil.
func Clear[T any]() PtrPatcher[T] {
	return patchClearPtr[T]{}
}

type patchSetVal[T any] struct {
	val T
}

func (p *patchSetVal[T]) ValPatch(ptr *T) {
	*ptr = p.val
}

func (p *patchSetVal[T]) PtrPatch(ptr **T) {
	*ptr = &p.val
}

// Set returns a patch operation that sets a value or pointer to the given value.
func Set[T any](val T) Patcher[T] {
	return &patchSetVal[T]{val: val}
}
