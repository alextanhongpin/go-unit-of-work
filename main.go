package main

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/alextanhongpin/uow"
	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("connecting to postgres")
	db, err := sql.Open("postgres", "postgres://john:123456@127.0.0.1:5432/test?sslmode=disable")
	if err != nil {
		panic(err)
	}
	if err := db.Ping(); err != nil {
		panic(err)
	}
	fmt.Println("connected to pg")

	uowFactory := func() *uow.UnitOfWork {
		return uow.New(db)
	}

	ctx := context.Background()
	if err := uowFactory().AtomicFn(ctx, doWork); err != nil {
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
