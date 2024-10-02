package middleware

import (
	"context"
	goErrors "errors"
	"github.com/kermesse-backend/internal/types"
	"github.com/kermesse-backend/internal/users"
	"github.com/kermesse-backend/pkg/errors"
	"github.com/kermesse-backend/pkg/jwt"
	"net/http"
	"os"
	"strings"
)

func IsAuth(handlerFunc errors.ErrorHandler, usersRepository users.UsersRepository, roles ...string) errors.ErrorHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		token := r.Header.Get("Authorization")
		if token == "" {
			return errors.CustomError{
				Key: errors.Unauthorized,
				Err: goErrors.New("token is required"),
			}
		}

		tokenParts := strings.Split(token, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			return errors.CustomError{
				Key: errors.Unauthorized,
				Err: goErrors.New("invalid token"),
			}
		}

		userId, err := jwt.GetTokenUserId(tokenParts[1], os.Getenv("JWT_SECRET"))
		if err != nil {
			return errors.CustomError{
				Key: errors.Unauthorized,
				Err: goErrors.New("invalid token"),
			}
		}

		user, err := usersRepository.GetUserById(userId)
		if err != nil {
			return err
		}

		if len(roles) > 0 {
			roleAllowed := false
			for _, role := range roles {
				if user.Role == role {
					roleAllowed = true
					break
				}
			}
			if !roleAllowed {
				return errors.CustomError{
					Key: errors.Forbidden,
					Err: goErrors.New("user does not have the required role"),
				}
			}
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, types.UserIDSessionKey, user.Id)
		ctx = context.WithValue(ctx, types.UserRoleSessionKey, user.Role)
		r = r.WithContext(ctx)

		return handlerFunc(w, r)
	}
}
