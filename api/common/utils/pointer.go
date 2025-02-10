package utils

// PtrTo returns a pointer to the given value
func PtrTo[T any](v T) *T {
	return &v
}

// ValueOr returns the value of the pointer or the default value if the pointer is nil
func ValueOr[T any](ptr *T, defaultValue T) T {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}
