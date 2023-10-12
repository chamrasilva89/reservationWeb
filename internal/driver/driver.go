package driver

import (
    "database/sql"
    "time"

    _ "github.com/jackc/pgconn"
    _ "github.com/jackc/pgx/v4"
    _ "github.com/jackc/pgx/v4/stdlib"
)

// DB represents the database connection.
type DB struct {
    SQL *sql.DB
}

// dbConn is a singleton instance of the DB type.
var dbConn = &DB{}

// Constants for configuring the database connection.
const maxOpenDBConn = 15
const maxIdleDBConn = 5
const maxDBLifeTime = 10 * time.Minute

// ConnectSQL connects to the PostgreSQL database and returns a DB object.
func ConnectSQL(dsn string) (*DB, error) {
	/*Inside ConnectSQL, a new database connection is created using the NewDatabase function. The connection is then configured with the specified maximum connection settings.*/
    d, err := NewDatabase(dsn)
    if err != nil {
        panic(err)
    }
    d.SetMaxOpenConns(maxOpenDBConn)
    d.SetMaxIdleConns(maxIdleDBConn)
    d.SetConnMaxIdleTime(maxDBLifeTime)

    dbConn.SQL = d

    err = testDB(d)
    if err != nil {
        return nil, err
    }
    return dbConn, nil
}

// testDB pings the database to check if it's reachable.
func testDB(d *sql.DB) error {
    err := d.Ping()
    if err != nil {
        return err
    }
    return nil
}

// NewDatabase creates a new SQL database connection and pings it.
func NewDatabase(dsn string) (*sql.DB, error) {
    db, err := sql.Open("pgx", dsn)
    if err != nil {
        return nil, err
    }

    if err = db.Ping(); err != nil {
        return nil, err
    }

    return db, nil
}
