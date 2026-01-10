package catalog

// Patcher is an interface for types that can apply a patch to a value of type T.
type Patcher[T any] interface {
	Patch(*T)
}

// Opt is a generic optional value type.
// If the Opt is nil, it represents the absence of a value.
// If the Opt is non-nil, it contains a function that returns a value of type T.
type Opt[T any] func() T

// Get retrieves the value contained in the Opt.
// It panics if the Opt is nil.
// This is syntactic sugar, allowing users to write o.Get() instead of o().
func (o Opt[T]) Get() T {
	if o == nil {
		panic("attempted to get value from nil Opt")
	}
	return o()
}

// Ptr retrieves a pointer to the value contained in the Opt.
// It returns nil if the Opt is nil.
// The returned pointer is independent of the Opt, meaning changes to the Opt do not affect the pointer and vice versa.
func (o Opt[T]) Ptr() *T {
	if o == nil {
		return nil
	}
	val := o()
	return &val
}

// NewOpt creates a new Opt containing the given value.
func NewOpt[T any](val T) Opt[T] {
	return func() T {
		return val
	}
}

// NilOpt creates a nil Opt, representing the absence of a value.
// It is also possible to use 'nil' directly where an Opt is expected.
func NilOpt[T any]() Opt[T] {
	return nil
}

// NewOptPtr creates a new Opt from a pointer to a value.
// If the pointer is nil, it returns a nil Opt.
// If the pointer is non-nil, it returns an Opt containing the value pointed to.
// The returned Opt is independent of the input pointer, meaning changes to the pointer do not affect the Opt and vice versa.
func NewOptPtr[T any](val *T) Opt[T] {
	if val == nil {
		return NilOpt[T]()
	}
	return NewOpt(*val)
}

// Patch is a function that applies modifications to a value of type T.
// A Patch can be nil, in which case it does nothing.
// Patch implements the Patcher interface.
type Patch[T any] func(*T)

// Patch applies the patch to the given pointer to T.
// If the Patch is nil, it does nothing.
func (p Patch[T]) Patch(ptr *T) {
	if p == nil {
		return
	}
	p(ptr)
}

// NewPatch creates a new Patch that, when called, updates the patched value to be equal to the given value.
func NewPatch[T any](val T) Patch[T] {
	return func(ptr *T) {
		*ptr = val
	}
}