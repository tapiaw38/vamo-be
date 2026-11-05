package auth

import "context"

type contextKey string

const userIDContextKey contextKey = "vamo.user_id"
const rolesContextKey contextKey = "vamo.roles"

func contextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDContextKey, userID)
}
func UserID(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(userIDContextKey).(string)
	return value, ok
}

func contextWithRoles(ctx context.Context, roles []string) context.Context {
	return context.WithValue(ctx, rolesContextKey, roles)
}

func Roles(ctx context.Context) []string {
	value, _ := ctx.Value(rolesContextKey).([]string)
	return value
}

func HasRole(ctx context.Context, role string) bool {
	for _, candidate := range Roles(ctx) {
		if candidate == role {
			return true
		}
	}
	return false
}
