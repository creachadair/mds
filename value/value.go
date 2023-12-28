// Package value defines adapters for value types.
package value

// Ptr returns a pointer to its argument type containing v.
func Ptr[T any](v T) *T { return &v }
