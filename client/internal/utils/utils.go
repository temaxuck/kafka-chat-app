package utils

func NilOrString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func StrPtr(s string) *string {
	return &s
}
