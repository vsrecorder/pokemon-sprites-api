package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/vsrecorder/pokemon-sprites-api/internal/controller"
	"github.com/vsrecorder/pokemon-sprites-api/internal/infrastructure"
	"github.com/vsrecorder/pokemon-sprites-api/internal/infrastructure/postgres"
	"github.com/vsrecorder/pokemon-sprites-api/internal/usecase"
	"gorm.io/gorm"
)

const (
	relativePath = "/api/v1beta"
)

type APIServer struct {
	httpServer *http.Server
	db         *gorm.DB
}

func NewAPIServer(addr string, handler http.Handler, db *gorm.DB) *APIServer {
	return &APIServer{
		httpServer: &http.Server{
			Addr:    addr,
			Handler: handler,
		},
		db: db,
	}
}

func (s *APIServer) Start(ctx context.Context) error {
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("server listen error: %v", err)
		}
	}()
	log.Printf("server started on %s", s.httpServer.Addr)

	<-ctx.Done()
	return s.Shutdown()
}

func (s *APIServer) Shutdown() error {
	log.Println("shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Printf("server forced to shutdown: %v", err)
		return err
	}

	log.Println("cleanup: closing DB connection...")

	if sqlDB, err := s.db.DB(); err != nil {
		log.Printf("db close error: %v", err)
	} else {
		if err := sqlDB.Close(); err != nil {
			log.Printf("db close error: %v", err)
		}
	}

	log.Printf("db closed")
	log.Println("server exited cleanly")

	return nil
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("failed to load .env file: %v", err)
	}

	dbHostname := os.Getenv("DB_HOSTNAME")
	dbPort := os.Getenv("DB_PORT")
	userName := os.Getenv("DB_USER_NAME")
	userPassword := os.Getenv("DB_USER_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	db, err := postgres.NewDB(dbHostname, dbPort, userName, userPassword, dbName)
	if err != nil {
		log.Fatalf("failed to connect database: %v\n", err)
	}

	r := gin.Default()
	r.SetTrustedProxies(nil)
	r.Use(cors.New(cors.Config{
		AllowHeaders: []string{
			"Authorization",
			"Content-Type",
		},
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},
		AllowOrigins: []string{
			"http://localhost:3000",
			"https://vsrecorder.mobi",
			"https://beta.vsrecorder.mobi",
			"https://local.vsrecorder.mobi",
		},
		AllowCredentials: false,
		MaxAge:           1 * time.Hour,
	}))

	controller.NewPokemonSprite(
		r,
		infrastructure.NewPokemonSprite(db),
		usecase.NewPokemonSprite(
			infrastructure.NewPokemonSprite(db),
		),
	).RegisterRoute(relativePath)

	{
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		server := NewAPIServer(":8998", r, db)
		if err := server.Start(ctx); err != nil {
			log.Fatalf("failed to run server: %v", err)
		}
	}
}
