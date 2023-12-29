package dbx

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

const (
	DEFAULT_DRIVER = "sqlite"
)

var (
	//go:embed schema.sql
	schema string
	//go:embed populate.sql
	populate    string
	DbAdminUser string = os.Getenv("DB_AUTH_NAME")
	DbAdminPwd  string = os.Getenv("DB_AUTH_PWD")
)

type DB struct {
	*sql.DB
}

func Dial(ctx context.Context, driver, uri string) *DB {
	db, err := sql.Open(driver, uri)
	if err != nil {
		fmt.Println("Failed to connect to database:", err)
		panic(err)
	}
	err = db.PingContext(ctx)
	if err != nil {
		fmt.Printf("Pinging with context: %s", err)
		panic(err)
	}

	_, err = db.ExecContext(ctx, schema)
	if err != nil {
		panic(err)
	}
	_, err = db.ExecContext(ctx, populate)
	if err != nil {
		panic(err)
	}

	return &DB{db}
}
