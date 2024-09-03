// Package value defines adapters for value types.
package value

// Ptr returns a pointer to its argument type containing v.
func Ptr[T any](v T) *T { return &v }

// At returns the value pointed to by p, or zero if p == nil.
func At[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

// AtDefault returns the value pointed to by p, or dflt if p == nil.
func AtDefault[T any](p *T, dflt T) T {
	if p == nil {
		return dflt
	}
	return *p
}

// Cond returns x if b is true, otherwise it returns y.
func Cond[T any](b bool, x, y T) T {
	if b {
		return x
	}
	return y
}

// AtMaybe returns Just(*p) if p != nil, or otherwise Absent().
func AtMaybe[T any](p *T) Maybe[T] {
	if p == nil {
		return Absent[T]()
	}
	return Just(*p)
}
