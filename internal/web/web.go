package web

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	zap "go.uber.org/zap"

	"secondarymetabolites.org/mibig-api/internal/mailer"
	"secondarymetabolites.org/mibig-api/internal/models"
)

type application struct {
	logger         *zap.SugaredLogger
	Models         models.Models
	Mail           mailer.Mailer
	Mux            *gin.Engine
	RepositoryPath string
}

func Run(debug bool) {

	if !debug {
		// set Gin to release mode
		gin.SetMode(gin.ReleaseMode)
	}

	logger := setupLogging(debug)
	defer logger.Sync()

	db, err := initDb(viper.GetString("database.uri"))
	if err != nil {
		logger.Fatalf(err.Error())
	}

	mailConfig := mailer.MailConfig{
		Host:     viper.GetString("mail.host"),
		Port:     viper.GetInt("mail.port"),
		Username: viper.GetString("mail.username"),
		Password: viper.GetString("mail.password"),
		Sender:   viper.GetString("mail.sender"),
	}

	mailSender := mailer.New(&mailConfig)
	mux := setupMux(debug, logger.Desugar())

	repositoryPath, err := filepath.Abs(viper.GetString("server.repository"))
	if err != nil {
		logger.Fatalf(err.Error())
	}
	logger.Infow("using repository path", "path", repositoryPath)

	app := &application{
		logger:         logger,
		Models:         models.NewModels(db),
		Mail:           mailSender,
		Mux:            mux,
		RepositoryPath: repositoryPath,
	}

	mux = app.routes()

	address := fmt.Sprintf("%s:%d", viper.GetString("server.address"), viper.GetInt("server.port"))

	srv := &http.Server{
		Addr:         address,
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	shutdownError := make(chan error)

	// Gracefully shut down
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		s := <-quit
		logger.Infow("caught signal, shutting down", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		shutdownError <- srv.Shutdown(ctx)
	}()

	logger.Infow("starting server",
		"address", address,
	)
	err = srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		logger.Fatalf(err.Error())
	}

	err = <-shutdownError
	if err != nil {
		logger.Fatalf(err.Error())
	}

	logger.Infow("stopped server", "address", address)

}

func setupMux(debug bool, logger *zap.Logger) *gin.Engine {
	var mux *gin.Engine
	if !debug {
		// In production mode, use zap Logger middleware
		mux = gin.New()
		mux.Use(ginzap.Ginzap(logger, time.RFC3339, true))
		mux.Use(ginzap.RecoveryWithZap(logger, true))
	} else {
		// otherwise use the default Gin logging, which is prettier
		mux = gin.Default()
	}
	return mux
}

func setupLogging(debug bool) *zap.SugaredLogger {
	logger, err := zap.NewProduction()
	if debug {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		log.Fatalf("Failed to set up logging: %s", err.Error())
	}
	return logger.Sugar()
}

func initDb(dbUri string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbUri)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
