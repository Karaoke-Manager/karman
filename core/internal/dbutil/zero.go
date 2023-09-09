package dbutil

func ZeroNil[T any](p *T) T {
	if p != nil {
		return *p
	}
	var zero T
	return zero
}
