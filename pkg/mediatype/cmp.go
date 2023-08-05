package mediatype

// Equals compares t1 to t2 and returns true if both media types should be considered equal.
// This method checks parameter equality as well.
// If you only want to check for equal types, use [EqualsType].
func Equals(t1 MediaType, t2 MediaType) bool {
	return t1.Equals(t2)
}

// EqualsType checks if t1 and t2 describe the same fundamental type.
// In contrast to [Equals] this method ignores any type parameters.
// Wildcards and their subtypes are NOT considered equal.
func EqualsType(t1 MediaType, t2 MediaType) bool {
	return t1.EqualsType(t2)
}
