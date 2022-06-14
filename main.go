package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/alextanhongpin/go-unit-of-work/pkg/uow"
	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", "postgres://john:123456@127.0.0.1:5432/test?sslmode=disable")
	if err != nil {
		panic(err)
	}
	if err := db.Ping(); err != nil {
		panic(err)
	}

	uowFactory := func() *uow.UnitOfWork {
		return uow.New(db)
	}

	tx, err := uowFactory().Atomic()
	if err != nil {
		panic(err)
	}

	if err := doWork(tx); err != nil {
		panic(err)
	}

	if err := uowFactory().AtomicFn(doWork2); err != nil {
		panic(err)
	}
}

func doWork(db *uow.UnitOfWork) error {
	fmt.Println("isTx?", db.IsTx())

	defer func() {
		if err := db.Rollback(); err != nil {
			log.Fatalf("failed to rollback: %s", err)
		}
	}()

	ctx := context.Background()

	var n int
	if err := db.QueryRowContext(ctx, `select 1 + 1`).Scan(&n); err != nil {
		return fmt.Errorf("failed to query: %w", err)
	}
	fmt.Println("result:", n)

	if err := db.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	return nil
}

func doWork2(db *uow.UnitOfWork) error {
	fmt.Println("isTx?", db.IsTx())

	ctx := context.Background()

	var n int
	if err := db.QueryRowContext(ctx, `select 1 + 1`).Scan(&n); err != nil {
		return fmt.Errorf("failed to query: %w", err)
	}
	fmt.Println("result:", n)

	return nil
}
