package web

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"secondarymetabolites.org/mibig-api/internal/data"
)

const AUTH_TOKEN_DURATION = 24 * time.Hour
const AUTH_COOKIE_NAME = "authentication_token"

type loginData struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (app *application) Login(c *gin.Context) {
	login := loginData{}
	err := c.BindJSON(&login)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	user, err := app.Models.Submitters.Authenticate(login.Email, login.Password)
	if err != nil {
		if errors.Is(err, data.ErrInvalidCredentials) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		} else {
			app.logger.Error(err.Error())
			app.serverError(c, err)
			return
		}
	}

	token, err := app.Models.Tokens.New(user.Id, AUTH_TOKEN_DURATION, data.ScopeAuthentication)
	if err != nil {
		app.serverError(c, err)
	}

	c.SetCookie(AUTH_COOKIE_NAME, token.Plaintext, int(AUTH_TOKEN_DURATION/1000), "/", viper.GetString("server.name"), false, true)
	c.JSON(http.StatusOK, gin.H{"authentication_token": token})
}

func (app *application) Logout(c *gin.Context) {
	c.SetCookie(AUTH_COOKIE_NAME, "", -1, "/", viper.GetString("server.name"), false, true)
	c.AbortWithStatus(http.StatusNoContent)
}

func (app *application) AuthTest(c *gin.Context) {
	user := app.GetCurrentUser(c)
	c.String(http.StatusOK, "Hello %s!", user.CallName)
}

func (app *application) Register(c *gin.Context) {
	var input struct {
		Name        string `json:"name"`
		Email       string `json:"email"`
		Password    string `json:"password"`
		CallName    string `json:"call_name"`
		Institution string `json:"institution"`
		Public      bool   `json:"public"`
		GdprConsent bool   `json:"gdpr_consent"`
	}
	err := c.BindJSON(&input)
	if err != nil {
		app.clientError(c, http.StatusBadRequest)
		return
	}

	user := &data.User{
		Email:       input.Email,
		Name:        input.Name,
		CallName:    input.CallName,
		Institution: input.Institution,
		Public:      input.Public,
		GDPRConsent: input.GdprConsent,
		Active:      false,
	}

	err = app.Models.Submitters.Insert(user, input.Password)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			app.clientErrorWithMessage(c, http.StatusBadRequest, "Email address already in use.")
		default:
			app.serverError(c, err)
		}
		return
	}

	user, err = app.Models.Submitters.Get(input.Email, false)
	if err != nil {
		app.serverError(c, err)
		return
	}

	token, err := app.Models.Tokens.New(user.Id, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		app.serverError(c, err)
		return
	}

	app.background(func() {
		data := map[string]interface{}{
			"activationToken": token.Plaintext,
			"baseUrl":         viper.GetString("ui.base"),
		}

		err = app.Mail.SendFromTemplate(user.Email, "user_welcome.tmpl", data)
		if err != nil {
			app.logger.Errorw("failed to send email", "error", err.Error())
		}
	})

	c.JSON(http.StatusCreated, gin.H{"user_id": user.Id})
}

func (app *application) Activate(c *gin.Context) {
	var input struct {
		TokenPlaintext string `json:"token"`
	}

	err := c.BindJSON(&input)
	if err != nil {
		app.clientError(c, http.StatusBadRequest)
		return
	}

	// TODO: Add a validation check here

	user, err := app.Models.Submitters.GetForToken(data.ScopeActivation, input.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.clientErrorWithMessage(c, http.StatusBadRequest, "invalid or expired activation token")
		default:
			app.serverError(c, err)
		}
		return
	}

	user.Active = true

	err = app.Models.Submitters.Update(user, "")
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflict(c)
		default:
			app.serverError(c, err)
		}
		return
	}

	err = app.Models.Tokens.DeleteAllForUser(user.Id, data.ScopeActivation)
	if err != nil {
		app.serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"user_id": user.Id, "active": user.Active})
}
