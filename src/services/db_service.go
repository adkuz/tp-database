package services

import (
	"fmt"
	"database/sql"

	_ "github.com/lib/pq"

	"os"
	"strings"
)


type Database interface {
	Setup(string)
	Execute(string)
	Result()
	Close()
	Query(string)
	QueryRow(query string, args ...interface{})
}



type Config struct {
	// Address that locates our postgres instance
	Host string
	// Port to connect to
	Port string
	// User that has access to the database
	User string
	// Password so that the user can login
	Password string
	// Database to connect to (must have been created priorly)
	DBName string
}

type PostgresDatabase struct {
	Connection *sql.DB
	LastResult sql.Result
}


func MakeConnectionString(config Config) string {
	return fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		config.User, config.Password, config.DBName, config.Host, config.Port)
}

func (pgdb *PostgresDatabase) QueryRow (query string, args ...interface{}) *sql.Row {
	return pgdb.Connection.QueryRow(query, args...)
}

func (pgdb *PostgresDatabase) Prepare(query string) (*sql.Stmt, error) {
	return pgdb.Connection.Prepare(query)
}

func Connect(dbConnectionString string) PostgresDatabase {
	pgdb := PostgresDatabase{Connection: nil, LastResult: nil}

	// Here one heed lib/pq
	db, err := sql.Open("postgres", dbConnectionString)
	if err != nil {
		fmt.Println("DB connection error: ", err)
		panic(err)
	}

	pgdb.Connection, db = db, pgdb.Connection

	if err := pgdb.Connection.Ping(); err != nil {
		fmt.Println("DB ping error: ", err)
		panic(err)
	}
	return pgdb
}

func (pgdb *PostgresDatabase) Setup(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Setupfile opening error: ", err)
		panic(err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		fmt.Println("Error after opening setupfile: ", err)
		panic(err)
	}

	bs := make([]byte, stat.Size())
	_, err = file.Read(bs)
	if err != nil {
		fmt.Println("Error after opening setupfile: ", err)
		panic(err)
	}

	commands := strings.Split(string(bs), ";")
	for i := range commands {
		if commands[i] != "" {
			//fmt.Println(i, ": ", commands[i]+";")
			pgdb.Execute(commands[i] + ";")
		}
	}
}

func (pgdb *PostgresDatabase) Execute(query string, args ...interface{}) sql.Result {

	res, err := pgdb.Connection.Exec(query, args...)
	if err != nil {
		panic(err)
	}
	return res
}


func (pgdb *PostgresDatabase) Query(query string, args ...interface{}) *sql.Rows {

	res, err := pgdb.Connection.Query(query, args...)
	if err != nil {
		panic(err)
	}
	return res
}




func (pgdb *PostgresDatabase) Result() sql.Result {
	return pgdb.LastResult
}

func (pgdb *PostgresDatabase) Close() {
	pgdb.Connection.Close()
}
