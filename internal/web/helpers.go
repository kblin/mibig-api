package web

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	zap "go.uber.org/zap"
	"secondarymetabolites.org/mibig-api/internal/data"
)

func (app *application) clientError(c *gin.Context, status int) {
	c.JSON(status, gin.H{"message": http.StatusText(status)})
}

func (app *application) clientErrorWithMessage(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message, "message": http.StatusText(status)})
}

func (app *application) editConflict(c *gin.Context) {
	app.clientErrorWithMessage(c, http.StatusConflict, "unable to update due to an edit conflict, please try again")
}

func (app *application) serverError(c *gin.Context, err error) {
	app.logger.Errorw("server error", zap.Error(err))
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": http.StatusText(http.StatusInternalServerError)})
}

func (app *application) invalidAuthToken(c *gin.Context) {
	c.Writer.Header().Set("WWW-Authenticate", "Bearer")
	app.clientErrorWithMessage(c, http.StatusUnauthorized, "invalid or missing authentication token")
}

func (app *application) authenticationRequired(c *gin.Context) {
	app.clientErrorWithMessage(c, http.StatusUnauthorized, "you must be authenticated to access this resource")
}

func (app *application) inactiveAccount(c *gin.Context) {
	app.clientErrorWithMessage(c, http.StatusUnauthorized, "your account must be activated to access this resouce")
}

func (app *application) notPermitted(c *gin.Context) {
	message := "your account doesn't have the necessary permissions to access this resource"
	app.clientErrorWithMessage(c, http.StatusUnauthorized, message)
}

func (app *application) GetCurrentUser(c *gin.Context) *data.User {
	return c.MustGet("user").(*data.User)
}

func (app *application) background(fn func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				app.logger.Error(fmt.Errorf("%s", err))
			}
		}()

		fn()
	}()
}
