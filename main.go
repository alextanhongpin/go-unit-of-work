package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

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

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()

		fmt.Println("do atomic", 1)
		if err := uowFactory().AtomicLock(context.Background(), 100, sleepAndWork); err != nil {
			panic(err)
		}
		fmt.Println("done atomic", 1)
	}()

	go func() {
		defer wg.Done()

		fmt.Println("do atomic", 2)
		if err := uowFactory().AtomicLock(context.Background(), 100, sleepAndWork); err != nil {
			panic(err)
		}
		fmt.Println("done atomic", 1)
	}()

	go func() {
		defer wg.Done()

		fmt.Println("do try atomic", 3)
		if err := uowFactory().AtomicTryLock(context.Background(), 100, sleepAndWork); err != nil {
			fmt.Println("do try atomic failed", err)
		} else {
			fmt.Println("done do try atomic", 3)
		}
	}()

	wg.Wait()
	fmt.Println("done")
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

func sleepAndWork(db *uow.UnitOfWork) error {
	fmt.Println("isTx?", db.IsTx())
	time.Sleep(5 * time.Second)

	ctx := context.Background()

	var n int
	if err := db.QueryRowContext(ctx, `select 1 + 1`).Scan(&n); err != nil {
		return fmt.Errorf("failed to query: %w", err)
	}
	fmt.Println("result:", n)

	return nil
}
