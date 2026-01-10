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

// Optional is a wrapper type that indicates whether a value is present or not.
// It is always in one of two states:
// - Nil: indicates that the value is not present.
// - Non-nil: indicates that the value is present, and holds the value.
// The zero value of Optional is the Nil state.
type Optional[T any] struct {
	val *T
}

// IsNil returns true if the Optional is in the Nil state.
func (o Optional[T]) IsNil() bool {
	return o.val == nil
}

// Get returns the value and a boolean indicating whether the value is present.
// If the value is Nil, the zero value of T is returned.
func (o Optional[T]) Get() (T, bool) {
	if o.val == nil {
		var zero T
		return zero, false
	}
	return *o.val, true
}

// Must returns the value, panicking if the value is Nil.
func (o Optional[T]) Must() T {
	if o.val == nil {
		panic("Optional: value is Nil")
	}
	return *o.val
}

// Ptr returns a pointer to the value, or nil if the value is Nil.
// The returned pointer is a copy of the internal pointer, so modifying
// the returned pointer does not affect the Optional (except for reference types).
func (o Optional[T]) Ptr() *T {
	if o.val == nil {
		return nil
	}
	newVal := new(T)
	*newVal = *o.val
	return newVal
}
