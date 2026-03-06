package main

import (
	"context"
	"ecom/internal/env"
	"log"

	"github.com/jackc/pgx/v5"
)

func main() {
	cfg := config{
		addr: ":8080",
		db: dbConfig{
			dsn: env.GetString("GOOSE_DBSTRING", "host=xxxxxx user=postgres password=xxxxx dbname=xxxxx sslmode=disable"),
		},
	}
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, cfg.db.dsn)
	if err != nil {
		log.Println(err)
	}

	api := &application{
		config: cfg,
		db:     conn,
	}
	if err := api.run(api.mount()); err != nil {
		log.Fatal(err.Error())
	}

}
