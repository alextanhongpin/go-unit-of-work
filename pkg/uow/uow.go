package uow

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
)

var (
	ErrNestedTransaction = errors.New("uow: cannot nest transaction")
	ErrUnknownDBType     = errors.New("uow: unknown db type")
)

type IDB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type UnitOfWork struct {
	IDB
	once sync.Once
	isTx bool
}

func New(db IDB) *UnitOfWork {
	switch db.(type) {
	case *sql.DB:
		return &UnitOfWork{
			IDB: db,
		}
	case *sql.Tx:
		return &UnitOfWork{
			IDB:  db,
			isTx: true,
		}
	default:
		panic(fmt.Errorf("%w: %+v", ErrUnknownDBType, db))
	}
}

func (uow *UnitOfWork) Atomic() (*UnitOfWork, error) {
	if uow.isTx {
		return nil, ErrNestedTransaction
	}

	db, ok := uow.IDB.(*sql.DB)
	if !ok {
		return nil, fmt.Errorf("%w: %+v", ErrUnknownDBType, db)
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	return New(tx), nil
}

func (uow *UnitOfWork) AtomicTx(ctx context.Context, opts *sql.TxOptions) (*UnitOfWork, error) {
	if uow.isTx {
		return nil, ErrNestedTransaction
	}

	db, ok := uow.IDB.(*sql.DB)
	if !ok {
		return nil, fmt.Errorf("%w: %+v", ErrUnknownDBType, db)
	}

	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	return New(tx), nil
}

func (uow *UnitOfWork) IsTx() bool {
	return uow.isTx
}

func (uow *UnitOfWork) Commit() (err error) {
	uow.once.Do(func() {
		if !uow.isTx {
			return
		}

		err = uow.IDB.(*sql.Tx).Commit()
	})

	return
}

func (uow *UnitOfWork) Rollback() (err error) {
	uow.once.Do(func() {
		if !uow.isTx {
			return
		}

		err = uow.IDB.(*sql.Tx).Rollback()
	})

	return
}
