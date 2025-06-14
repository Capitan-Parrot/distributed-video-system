package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
)

type querier interface {
	Query(query string, args ...any) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type txKey struct{}

// InTx выполняет функцию в транзакции
func (d *Database) InTx(ctx context.Context, fn func(ctx context.Context) error) error {
	// Проверяем, есть ли уже транзакция в контексте
	if tx := d.txFromCtx(ctx); tx != nil {
		return fn(ctx) // Уже в транзакции, выполняем функцию
	}

	// Начинаем новую транзакцию
	tx, err := d.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Добавляем транзакцию в контекст
	ctx = context.WithValue(ctx, txKey{}, tx)

	// Выполняем функцию
	err = fn(ctx)
	if err != nil {
		// При ошибке откатываем транзакцию
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Printf("cannot rollback transaction: %v", rollbackErr)
		}
		return err
	}

	// Коммитим транзакцию
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// querier возвращает текущую транзакцию или соединение с БД
func (d *Database) querier(ctx context.Context) querier {
	if tx := d.txFromCtx(ctx); tx != nil {
		return tx
	}
	return d.DB
}

// txFromCtx извлекает транзакцию из контекста
func (d *Database) txFromCtx(ctx context.Context) *sql.Tx {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		return tx
	}
	return nil
}
