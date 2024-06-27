package web

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"secondarymetabolites.org/mibig-api/internal/data"
	"secondarymetabolites.org/mibig-api/internal/utils"
)

func (app *application) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Let caches know that the response will vary depending on the Authorization header
		c.Writer.Header().Add("Vary", "Authorization")

		var user *data.User

		token, err := getToken(c)
		if err != nil {
			switch {
			// No credentials means we default to the anonymous user
			case errors.Is(err, data.ErrNoCredentails):
				c.Set("user", data.AnonymousUser)
				c.Next()
			case errors.Is(err, data.ErrInvalidCredentials):
				c.AbortWithStatus(http.StatusUnauthorized)
			default:
				app.serverError(c, err)
			}
			return
		}

		user, err = app.Models.Submitters.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthToken(c)
			default:
				app.serverError(c, err)
			}
			return
		}
		c.Set("user", user)

		c.Next()

	}
}

const HEADER_PREFIX string = "Bearer "

func getToken(c *gin.Context) (string, error) {

	authCookie, err := c.Cookie(AUTH_COOKIE_NAME)
	if err == nil {
		return authCookie, nil
	}

	header := c.GetHeader("Authorization")
	if header == "" {
		return "", data.ErrNoCredentails
	}
	if !strings.HasPrefix(header, HEADER_PREFIX) {
		return "", data.ErrInvalidCredentials
	}

	header = strings.TrimPrefix(header, HEADER_PREFIX)

	return header, nil
}

func (app *application) RequireAuthenticatedUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := app.GetCurrentUser(c)

		if user.IsAnonymous() {
			app.authenticationRequired(c)
			return
		}

		c.Next()
	}
}

func (app *application) RequireActivatedUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := app.GetCurrentUser(c)

		if user.IsAnonymous() {
			app.authenticationRequired(c)
			return
		}

		if !user.Active {
			app.inactiveAccount(c)
			return
		}

		c.Next()
	}
}

func (app *application) RequireRoles(requiredRoles []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := app.GetCurrentUser(c)

		if user.IsAnonymous() {
			app.authenticationRequired(c)
			return
		}

		if !user.Active {
			app.inactiveAccount(c)
			return
		}

		validRoles := utils.IntersectString(requiredRoles, data.RolesToStrings(user.Roles))
		if len(validRoles) == 0 {
			app.notPermitted(c)
			return
		}
	}
}
