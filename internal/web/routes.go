package web

import (
	"github.com/gin-gonic/gin"
)

func (app *application) routes() *gin.Engine {
	app.Mux.Use(app.Authenticate())
	api := app.Mux.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			v1.GET("/version", app.version)
			v1.GET("/stats", app.stats)
			v1.GET("/repository", app.repository)
			v1.POST("/search", app.search)
			v1.GET("/available/:category/:term", app.available)
			v1.GET("/convert", app.Convert)
			v1.GET("/contributors", app.Contributors)
			v1.POST("/submit", app.submit)

			/*
				user := v1.Group("/user")
				{
					user.POST("/login", app.Login)
					user.POST("/logout", app.Logout)
					user.POST("/register", app.Register)
					user.PUT("/activate", app.Activate)
				}

				v1.GET("/authtest", app.AuthTest)

				edit := v1.Group("/edit", app.RequireRoles([]string{"submitter"}))
				{
					edit.GET("/", app.listRequests)
					edit.POST("/submit", app.submit)
				}
				v1.POST("/bgc-registration", app.JWTAuthenticated([]models.Role{}), app.LegacyStoreSubmission)
				v1.POST("/bgc-detail-registration", app.JWTAuthenticated([]models.Role{}), app.LegacyStoreBgcDetailSubmission)
			*/
		}
	}

	redirect := app.Mux.Group("/go")
	{
		redirect.GET("/:accession", app.Redirect)
	}

	return app.Mux
}
