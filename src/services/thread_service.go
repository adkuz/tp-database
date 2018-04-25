package services

import (
	"fmt"
	"time"

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

	INSERT_QUERY:=
		"insert into " + ts.tableName +
			" (%s author, forum, created, title, message, votes)" +
				" values (%s $1, $2, $3, $4, $5, $6) returning id;"

	if thread.Slug == "" {
		INSERT_QUERY = fmt.Sprintf(INSERT_QUERY, "", "")
	} else {
		INSERT_QUERY = fmt.Sprintf(INSERT_QUERY, "slug, ", "'" + thread.Slug + "', ")
	}

	fmt.Println("AddThread: INSERT_QUERY:", INSERT_QUERY)

	err := ts.db.QueryRow(INSERT_QUERY, thread.Author, thread.Forum,
		thread.Created, thread.Title, thread.Message, thread.Votes).Scan(&thread.ID)

	if err != nil {
		fmt.Println("AddThread:  error after id:", err.Error())
		panic(err)
	}
	//for result.Next() {
	//	err := result.Scan(&thread.ID)
	//	if err != nil {
	//		fmt.Println("AddThread:  error after id:", err.Error())
	//		panic(err)
	//	}
	//}

	fmt.Println("AddForum: id:", thread.ID)

	return true, thread
}

func (ts *ThreadService) SelectThreads(slug, limit, since string, desc bool) (bool, []models.Thread) {

	limitStr := ""
	if limit != "" {
		fmt.Println("SelectThreads: limit:", limit)
		limitStr = "LIMIT " + limit
	}


	comp := " >= "
	order := "ORDER BY th.created "
	if desc {
		comp = " <= "
		order += "DESC "
	}

	offsetStr := ""
	if since != "" {
		layout := "2006-01-02T15:04:05.000Z"
		t, err := time.Parse(layout, since)
		if err != nil {
			fmt.Println(err)
		}
		since = t.UTC().Format(time.RFC3339)
		fmt.Println("SelectThreads: since:", since)
		offsetStr = "AND th.created " + comp + " '" + since + "'"
	}

	

	query := fmt.Sprintf(
		"SELECT id, slug, author, forum, created, title, message, votes FROM %s th WHERE LOWER(th.forum) = LOWER('%s') %s %s %s;",
		ts.tableName, slug, offsetStr, order,  limitStr)

	fmt.Println("SelectThreads: query:", query)


	rows := ts.db.Query(query)
	threads := make([]models.Thread, 0)

	for rows.Next() {
		var thread models.Thread
		err := rows.Scan(&thread.ID, &thread.Slug, &thread.Author, &thread.Forum, &thread.Created,
			&thread.Title, &thread.Message, &thread.Votes)
		if err != nil {
			panic(err)
		}
		threads = append(threads, thread)
	}

	if len(threads) == 0 {
		return false, nil
	}

	return true, threads
}

func (ts *ThreadService) GetThreadBySlug(slug string) *models.Thread {

	fmt.Println("GetThreadBySlug: query start")

	query := fmt.Sprintf(
		"SELECT slug, author, forum, created, title, message, votes FROM %s WHERE LOWER(slug) = LOWER('%s');",
			ts.tableName, slug)

	fmt.Println("GetThreadBySlug: query:", query)
	fmt.Println("-----------------------------start------------------------------####################")

	rows := ts.db.Query(query)
	fmt.Println("------------------------------end-------------------------------####################")

	for rows.Next() {
		thread := new(models.Thread)
		err := rows.Scan(&thread.Slug, &thread.Author, &thread.Forum, &thread.Created,
			&thread.Title, &thread.Message, &thread.Votes)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
		return thread
	}

	return nil
}