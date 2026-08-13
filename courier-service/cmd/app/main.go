package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
)

type Courier struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Phone  string `json:"phone"`
	Status string `json:"status"`
}

func listCouriers(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := pool.Query(r.Context(), "SELECT id, name, phone, status FROM couriers")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var couriers []Courier
		for rows.Next() {
			var c Courier
			if err := rows.Scan(&c.ID, &c.Name, &c.Phone, &c.Status); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			couriers = append(couriers, c)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(couriers)
	}
}

func getCourier(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := mux.Vars(r)["id"]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		var courier Courier
		err = pool.QueryRow(r.Context(), "SELECT id, name, phone, status FROM couriers WHERE id=$1", id).
			Scan(&courier.ID, &courier.Name, &courier.Phone, &courier.Status)
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "courier not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(courier)
	}
}

func createCourier(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var courier Courier
		if err := json.NewDecoder(r.Body).Decode(&courier); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if courier.Name == "" || courier.Phone == "" || courier.Status == "" {
			http.Error(w, "missing required fields", http.StatusBadRequest)
			return
		}

		err := pool.QueryRow(r.Context(),
			"INSERT INTO couriers (name, phone, status) VALUES ($1,$2,$3) RETURNING id",
			courier.Name, courier.Phone, courier.Status).Scan(&courier.ID)

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			http.Error(w, "courier with this phone already exists", http.StatusConflict) // 409
			return
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(courier)
	}
}

func updateCourier(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			ID     int64   `json:"id"`
			Name   *string `json:"name"`
			Phone  *string `json:"phone"`
			Status *string `json:"status"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}

		if request.ID == 0 {
			http.Error(w, "id is required field", http.StatusBadRequest)
			return
		}

		var setClauses []string
		var args []any
		argIdx := 1
		if request.Name != nil {
			setClauses = append(setClauses, fmt.Sprintf("name=$%d", argIdx))
			args = append(args, *request.Name)
			argIdx++
		}
		if request.Phone != nil {
			setClauses = append(setClauses, fmt.Sprintf("phone=$%d", argIdx))
			args = append(args, *request.Phone)
			argIdx++
		}
		if request.Status != nil {
			setClauses = append(setClauses, fmt.Sprintf("status=$%d", argIdx))
			args = append(args, *request.Status)
			argIdx++
		}
		if len(setClauses) == 0 {
			http.Error(w, "no fields to update", http.StatusBadRequest)
			return
		}
		// updated_at обновляем всегда при любом изменении
		setClauses = append(setClauses, fmt.Sprintf("updated_at=now()"))

		args = append(args, request.ID) // последний аргумент — id для WHERE
		query := fmt.Sprintf(
			"UPDATE couriers SET %s WHERE id=$%d RETURNING id, name, phone, status",
			strings.Join(setClauses, ", "), argIdx,
		)

		var c Courier
		err := pool.QueryRow(r.Context(), query, args...).Scan(&c.ID, &c.Name, &c.Phone, &c.Status)
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "courier not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(c)
	}
}

func run(ctx context.Context) error {
	log.SetOutput(os.Stdout)

	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"),
	)

	port := 8080
	if envPort := os.Getenv("PORT"); envPort != "" {
		p, err := strconv.Atoi(envPort)
		if err != nil {
			log.Fatalf("Invalid port format: %v", err)
		}
		port = p
	}

	flagPort := flag.Int("port", port, "Port to listen on")
	flag.Parse()
	if *flagPort != 0 {
		port = *flagPort
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer pool.Close()

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

	r := mux.NewRouter()

	r.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{\"message\": \"pong\"}"))
	}).Methods("GET")

	r.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}).Methods("HEAD")

	r.HandleFunc("/couriers", listCouriers(pool)).Methods("GET")
	r.HandleFunc("/courier", createCourier(pool)).Methods("POST")
	r.HandleFunc("/courier/{id}", getCourier(pool)).Methods("GET")
	r.HandleFunc("/courier", updateCourier(pool)).Methods("PUT")

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down service-courier")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Error shutting down server: %v", err)
	}

	return nil
}

func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error:", err)
		cancel()
		os.Exit(1)
	}
}
