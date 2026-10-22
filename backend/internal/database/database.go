// Pacote database cuida do pool pgx, de transações e das migrações.
package database

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Connect abre o pool de conexões e confirma que o banco responde.
func Connect(ctx context.Context, url string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("criando pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("conectando ao banco: %w", err)
	}
	return pool, nil
}

// WithTx executa fn dentro de uma transação, com commit em caso de sucesso e
// rollback em caso de erro ou pânico.
func WithTx(ctx context.Context, pool *pgxpool.Pool, fn func(tx pgx.Tx) error) (err error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("iniciando transação: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()
	if err = fn(tx); err != nil {
		return err
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("confirmando transação: %w", err)
	}
	return nil
}

// Migrator aplica as migrações embutidas no binário.
type Migrator struct{ db *sql.DB }

// NewMigrator cria um Migrator a partir da URL do banco.
func NewMigrator(pool *pgxpool.Pool) *Migrator {
	goose.SetBaseFS(migrationsFS)
	_ = goose.SetDialect("postgres")
	return &Migrator{db: stdlib.OpenDBFromPool(pool)}
}

// Up aplica todas as migrações pendentes.
func (m *Migrator) Up(ctx context.Context) error {
	return goose.UpContext(ctx, m.db, "migrations")
}

// Down desfaz a última migração.
func (m *Migrator) Down(ctx context.Context) error {
	return goose.DownContext(ctx, m.db, "migrations")
}

// Status imprime o estado das migrações.
func (m *Migrator) Status(ctx context.Context) error {
	return goose.StatusContext(ctx, m.db, "migrations")
}

// Close libera a conexão database/sql (não fecha o pool original).
func (m *Migrator) Close() error { return m.db.Close() }

// IsUniqueViolation informa se o erro é uma violação de unicidade (23505).
func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// IsForeignKeyViolation informa se o erro é uma violação de chave estrangeira (23503).
func IsForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}
