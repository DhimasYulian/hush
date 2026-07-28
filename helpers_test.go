package hush

// intPtr returns a pointer to v.
func intPtr(v int) *int {
	return &v
}

// boolPtr returns a pointer to v.
func boolPtr(v bool) *bool {
	return &v
}
