package main

import (
	"context"
	"database/sql"
	"log"

	"github.com/alextanhongpin/go-unit-of-work/app"
	"github.com/alextanhongpin/uow"
	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", "postgres://john:123456@127.0.0.1:5432/test?sslmode=disable")
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	u := uow.New(db)
	repo := app.NewPostgresUserRepository(u)
	if err := repo.Migrate(context.Background()); err != nil {
		log.Fatalf("failed to run migration: %v", err)
	}

	uc := app.NewUserUseCase(u, repo)

	email := "john.doe@mail.com"
	deviceID := "device-123"

	err = uc.Register(context.Background(), email, deviceID)
	if err != nil {
		log.Fatalf("failed to register user: %v", err)
	}
	log.Println("successfully registered")
}
