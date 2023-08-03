package contextkey

// ContextKey is a value for use with context.WithValue. It's used as
// a pointer so it fits in an any without allocation.
type ContextKey struct {
	Name string
}

func New(name string) *ContextKey {
	return &ContextKey{Name: name}
}

func (k *ContextKey) String() string { return "net/http context value " + k.Name }
