package main

import (
	"net/http"
	"time"

	"ecom/cmd/app/auth"
	repo "ecom/internal/adapters/postgresql/sqlc"

	"github.com/jackc/pgx/v5"
)

type application struct {
	config config
	db     *pgx.Conn
}

type config struct {
	addr string
	db   dbConfig
}

type dbConfig struct {
	dsn string
}

func (app *application) mount() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("I'm active"))
	})

	userService := auth.NewService(repo.New(app.db))
	userHandler := auth.NewHandler(userService)
	mux.HandleFunc("/signup", userHandler.Signup)
	mux.HandleFunc("/login", userHandler.Login)
	mux.HandleFunc("/refresh", userHandler.RefreshToken)
	mux.HandleFunc("/verify-user", userHandler.VerifyUser)
	mux.HandleFunc("/send-reset", userHandler.SendResetTokenToEmail)

	return mux
}

func (app *application) run(h http.Handler) error {
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      h,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}
	return srv.ListenAndServe()
}
