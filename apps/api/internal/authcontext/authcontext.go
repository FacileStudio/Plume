package authcontext

import "context"

// Identity holds the authenticated user's ID and email, carried through a
// request context.
type Identity struct {
	UserID string
	Email  string
}

type contextKey struct{}

// WithIdentity returns a copy of parentContext carrying the given Identity.
func WithIdentity(parentContext context.Context, identity Identity) context.Context {
	return context.WithValue(parentContext, contextKey{}, identity)
}

// IdentityFromContext extracts the Identity stored in parentContext, if any.
func IdentityFromContext(parentContext context.Context) (Identity, bool) {
	identity, ok := parentContext.Value(contextKey{}).(Identity)
	return identity, ok
}
