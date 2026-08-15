package main

import (
	"context"
	"courier-service/internal/handler"
	"courier-service/internal/repository"
	"courier-service/internal/usecase"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
	flag "github.com/spf13/pflag"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	dbPool := initDBPool()
	defer dbPool.Close()

	profileRepository := repository.NewProfileRepository(dbPool)

	healthUseCase := usecase.NewHealthUseCase(profileRepository)
	health := handler.NewHealthController(healthUseCase)

	profileUseCase := usecase.NewProfileUseCase(profileRepository)
	profile := handler.NewProfileController(profileUseCase)

	server := &http.Server{
		Addr:    ":" + resolvePort(),
		Handler: initRouter(profile, health),
	}

	go gracefulShutdown(server)

	log.Printf("Starting server on %s\n", server.Addr)
	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Error starting server: %v", err)
	}
}

func resolvePort() string {
	port := os.Getenv("PORT")

	var flagPort = flag.String("port", "", "Укажите порт")
	flag.Parse()

	if flagPort != nil && *flagPort != "" {
		port = *flagPort
	}

	if port == "" {
		port = "8080"
	}

	return port
}

func gracefulShutdown(server *http.Server) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	log.Println("Shutting down service-courier")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Error shutting down server: %v", err)
	}

	log.Println("Server gracefully stopped")
}

func initRouter(p *handler.ProfileController, h *handler.HealthController) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/ping", h.Ping)
	r.Head("/healthcheck", h.HealthCheck)

	r.Route("/couriers", func(r chi.Router) {
		r.Get("/", p.GetAll)
		r.Get("/{id}", p.GetOne)
		r.Post("/", p.Create)
		r.Put("/", p.Update)
		r.Delete("/{id}", p.Delete)
	})

	return r
}

func initDBPool() *pgxpool.Pool {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"),
	)

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Error opening db for migrations: %v", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("Error setting goose dialect: %v", err)
	}
	if err := goose.Up(db, "migrations"); err != nil {
		log.Fatalf("Error applying migrations: %v", err)
	}

	return pool
}
