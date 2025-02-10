package data

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"reflect"
	"time"

	"github.com/ahmetkoprulu/rtrp/common/utils"
	"github.com/golang-migrate/migrate"
	_ "github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	dbContextPg *PgDbContext
)

// PgDbContext represents a PostgreSQL database context
type PgDbContext struct {
	*pgxpool.Pool
	connectionString string
}

// QueryRunner interface for both Pool and Tx
type QueryRunner interface {
	Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error)
}

// TxFn is a function that will be called with a QueryRunner that is either the
// pool or an active transaction
type TxFn func(QueryRunner) error

func LoadPostgres(databaseUrl, DatabaseName string) error {
	u, err := url.Parse(databaseUrl)
	if err != nil {
		return err
	}

	u.Path = "/" + DatabaseName

	err = InitializePostgreDb(u.String())
	if err != nil {
		return err
	}

	return nil
}

func InitializePostgreDb(connectionString string) error {
	if dbContextPg != nil {
		return nil
	}

	m, err := migrate.New("file://migrations", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		utils.Logger.Fatal("Failed to start server", utils.Logger.String("error", err.Error()))
	}

	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return fmt.Errorf("unable to parse connection string: %v", err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %v", err)
	}

	dbContextPg = &PgDbContext{Pool: pool, connectionString: connectionString}
	return nil
}

func NewPgDbContext() (*PgDbContext, error) {
	if dbContextPg.Pool == nil {
		return nil, errors.New("PostgresDbContext is not initialized")
	}

	return dbContextPg, nil
}

// WithTransaction executes a function within a transaction
func (db *PgDbContext) WithTransaction(ctx context.Context, fn TxFn) error {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			panic(p)
		}
	}()

	err = fn(tx)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

// ScanRow scans a single row into a struct
func (db *PgDbContext) ScanRow(row pgx.Row, dest interface{}) error {
	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("dest must be a pointer to a struct")
	}

	v = v.Elem()
	t := v.Type()
	fields := make([]interface{}, 0)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		pgTag := field.Tag.Get("pg")
		if pgTag == "" {
			continue
		}

		fields = append(fields, v.Field(i).Addr().Interface())
	}

	return row.Scan(fields...)
}

// ScanRows scans multiple rows into a slice of structs
func (db *PgDbContext) ScanRows(rows pgx.Rows, dest interface{}) error {
	defer rows.Close()

	sliceValue := reflect.ValueOf(dest)
	if sliceValue.Kind() != reflect.Ptr || sliceValue.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("dest must be a pointer to a slice")
	}

	sliceType := sliceValue.Elem().Type()
	elementType := sliceType.Elem()

	slice := reflect.MakeSlice(sliceType, 0, 0)

	for rows.Next() {
		element := reflect.New(elementType).Elem()
		fields := make([]interface{}, 0)

		for i := 0; i < elementType.NumField(); i++ {
			field := elementType.Field(i)
			if field.Tag.Get("pg") == "" {
				continue
			}
			fields = append(fields, element.Field(i).Addr().Interface())
		}

		if err := rows.Scan(fields...); err != nil {
			return err
		}

		slice = reflect.Append(slice, element)
	}

	sliceValue.Elem().Set(slice)
	return rows.Err()
}

func (db *PgDbContext) GenerateNewId() string {
	return uuid.New().String()
}
