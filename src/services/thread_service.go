package services

import (
	"database/sql"
	"fmt"

	"github.com/Alex-Kuz/tp-database/src/models"
)

type ThreadService struct {
	db        *PostgresDatabase
	tableName string
}

func MakeThreadService(pgdb *PostgresDatabase) ThreadService {
	return ThreadService{db: pgdb, tableName: "threads"}
}


func (ts *ThreadService) AddThread(thread *models.Thread) (bool, *models.Thread) {

	INSERT_QUERY_WITH_SLUG :=
		"insert into " + ts.tableName +
			" (slug, author, forum, created, title, message, votes)" +
				"values ($1, $2, $3, $4, $5, $6, $7) returning id;"

	INSERT_QUERY_WITHOUT_SLUG :=
		"insert into " + ts.tableName +
			" (author, forum, created, title, message, votes)" +
			"values ($1, $2, $3, $4, $5, $6) returning id;"


	var result sql.Result

	insertQuery, err := ts.db.Prepare(INSERT_QUERY_WITH_SLUG)

	if len(thread.Slug) != 0 {
		insertQuery, err = ts.db.Prepare(INSERT_QUERY_WITH_SLUG)
		if err != nil {
			panic(err)
		}
		result, err = insertQuery.Exec(thread.Slug, thread.Author, thread.Forum,
			thread.Created, thread.Title, thread.Message, thread.Votes)

		if err != nil {
			fmt.Println("\nAddForum:  thread.Slug:", thread.Slug)
			fmt.Println("AddForum:  error:", err.Error())
			panic(err)
		}
	} else {
		insertQuery, err = ts.db.Prepare(INSERT_QUERY_WITHOUT_SLUG)
		if err != nil {
			panic(err)
		}
		result, err = insertQuery.Exec(thread.Author, thread.Forum,
			thread.Created, thread.Title, thread.Message, thread.Votes)

		if err != nil {
			fmt.Println("\nAddForum:  thread.Slug:", thread.Slug)
			fmt.Println("AddForum:  error:", err.Error())
			panic(err)
		}
	}

	id, err := result.RowsAffected()

	if err != nil {
		fmt.Println("AddForum:  error after id:", err.Error())
		panic(err)
	}

	thread.ID = id

	return true, thread

}


