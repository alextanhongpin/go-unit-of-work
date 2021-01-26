package main

import (
	"context"
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

// Need to know if it is a transaction or not.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type key string

var uowkey key = "uow"

func main() {
	db, err := sql.Open("postgres", "postgres://john:123456@127.0.0.1:5432/test?sslmode=disable")
	if err != nil {
		panic(err)
	}
	if err := db.Ping(); err != nil {
		panic(err)
	}

	uow := NewUnitOfWork(db) // The usecase layer needs to have a factory for UoW.
	ctx := context.Background()

	// Explicit wrapping of transaction context.
	err = uow.Wrap(ctx, func(ctx context.Context) error {
		if err := Repo(ctx); err != nil {
			return err
		}

		return Repo(ctx)
	})
	if err != nil {
		panic(err)
	}

	// Piping transaction, only if we do not care of the payload.
	uow.Pipe(ctx, Repo, Repo)

	// Perform single query without transaction.
	uow.Do(ctx, Repo)

	// Obtain db context without transaction.
	ctx = uow.Context(ctx)
	if err := Repo(ctx); err != nil {
		panic(err)
	}
}

func Repo(ctx context.Context) error {
	db := uowValue(ctx)
	var i int
	if err := db.QueryRowContext(ctx, "SELECT 1 + 1").Scan(&i); err != nil {
		return err
	}
	log.Printf("sum=%d tx=%t\n", i, db.IsTransaction)
	return nil
}

type UnitOfWork struct {
	db *sql.DB
}

func NewUnitOfWork(db *sql.DB) *UnitOfWork {
	return &UnitOfWork{db: db}
}

func (u *UnitOfWork) Wrap(ctx context.Context, fn func(context.Context) error) error {
	return Transact(u.db, func(tx *sql.Tx) error {
		ctx = context.WithValue(ctx, uowkey, &UnitOfWorkTx{DBTX: tx, IsTransaction: true})
		return fn(ctx)
	})
}

func (u *UnitOfWork) Do(ctx context.Context, fn func(context.Context) error) error {
	return fn(u.Context(ctx))
}

func (u *UnitOfWork) Context(ctx context.Context) context.Context {
	return context.WithValue(ctx, uowkey, &UnitOfWorkTx{DBTX: u.db})
}

func (u *UnitOfWork) Pipe(ctx context.Context, fns ...func(context.Context) error) error {
	return Transact(u.db, func(tx *sql.Tx) error {
		ctx = context.WithValue(ctx, uowkey, &UnitOfWorkTx{DBTX: tx, IsTransaction: true})
		for _, fn := range fns {
			if err := fn(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}

type UnitOfWorkTx struct {
	DBTX
	IsTransaction bool
}

func uowValue(ctx context.Context) *UnitOfWorkTx {
	if uow, ok := ctx.Value(uowkey).(*UnitOfWorkTx); ok {
		return uow
	}
	panic("no db context")
}

func Transact(db *sql.DB, txFn func(*sql.Tx) error) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			tx.Rollback() // err is non-nil; don't change it
		} else {
			err = tx.Commit() // err is nil; if Commit returns error update err
		}
	}()
	err = txFn(tx)
	return err
}
