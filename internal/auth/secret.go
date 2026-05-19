package auth

type JWTSecret struct {
	value string
}

func NewJWTSecret(s string) JWTSecret {
	return JWTSecret{value: s}
}

func (s JWTSecret) String() string { return "[REDACTED]" }
func (s JWTSecret) Bytes() []byte  { return []byte(s.value) }
func (s JWTSecret) IsZero() bool   { return s.value == "" }
