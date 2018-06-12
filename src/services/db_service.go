package services

import (
	"fmt"

	"github.com/jackc/pgx"

	"os"
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
	Port uint16
	// User that has access to the database
	User string
	// Password so that the user can login
	Password string
	// Database to connect to (must have been created priorly)
	DBName string
}

type PostgresDatabase struct {
	Connections *pgx.ConnPool
}

func MakeConnectionConfig(config Config) pgx.ConnConfig {
	return pgx.ConnConfig{
		Host:     config.Host,
		User:     config.User,
		Password: config.Password,
		Database: config.DBName,
		Port:     config.Port,
	}
}

func (pgdb *PostgresDatabase) DataBase() *pgx.ConnPool {
	return pgdb.Connections
}

func (pgdb *PostgresDatabase) QueryRow(query string, args ...interface{}) *pgx.Row {
	return pgdb.Connections.QueryRow(query, args...)
}

func Connect(connectionConfig pgx.ConnConfig) PostgresDatabase {
	pgdb := PostgresDatabase{Connections: nil}

	conns, err := pgx.NewConnPool(
		pgx.ConnPoolConfig{
			ConnConfig:     connectionConfig,
			MaxConnections: 42,
		},
	)
	if err != nil {
		fmt.Println("DB connection error: ", err)
		panic(err)
	}

	pgdb.Connections = conns

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

	command := string(bs)
	pgdb.Execute(command)
}

func (pgdb *PostgresDatabase) Execute(query string, args ...interface{}) pgx.CommandTag {

	res, err := pgdb.Connections.Exec(query, args...)
	if err != nil {
		panic(err)
	}
	return res
}

func (pgdb *PostgresDatabase) Query(query string, args ...interface{}) *pgx.Rows {

	res, err := pgdb.Connections.Query(query, args...)

	if err != nil {
		pgdb.Connections.Stat()
		fmt.Println("Query: " + query)
		fmt.Print("Args: ")
		fmt.Println(args)
		fmt.Println(err)
		panic(err)
	}
	return res
}

func (pgdb *PostgresDatabase) Close() {
	pgdb.Connections.Close()
}
