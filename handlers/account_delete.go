package handlers

import (
	"context"

	"github.com/gin-gonic/gin"

	"villum/account"
)

// The account package owns the transaction; these aliases let the handlers in
// this package compare errors without importing the package everywhere.
var (
	errLastAdmin       = account.ErrLastAdmin
	errAccountNotFound = account.ErrNotFound
)

func deleteAccount(ctx context.Context, userID int64) error {
	return account.Delete(ctx, userID)
}

// RegisterAccountAPIRoutes registers the JSON endpoints the Flutter client uses
// to manage its own account. The router group already supplies AuthRequired,
// APITokenRequired and CSRFRequired.
func RegisterAccountAPIRoutes(r *gin.RouterGroup) {
	r.DELETE("/me", HandleDeleteMe)
}
