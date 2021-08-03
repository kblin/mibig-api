package web

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	zap "go.uber.org/zap"
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

func (app *application) notFound(c *gin.Context) {
	app.clientError(c, http.StatusNotFound)
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
