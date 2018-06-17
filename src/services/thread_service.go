package services

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Alex-Kuz/tp-database/src/models"
	"github.com/jackc/pgx"
)

type ThreadService struct {
	db        *PostgresDatabase
	tableName string
}

func MakeThreadService(pgdb *PostgresDatabase) ThreadService {
	return ThreadService{db: pgdb, tableName: "threads"}
}

func (ts *ThreadService) TableName() string {
	return ts.tableName
}

func (ts *ThreadService) AddThread(thread *models.Thread) (bool, *models.Thread) {

	INSERT_QUERY :=
		"insert into threads (%s author, forum, created, title, message, votes)" +
			" values (%s $1, $2, $3, $4, $5, $6) returning id;"

	if thread.Slug == "" {
		INSERT_QUERY = fmt.Sprintf(INSERT_QUERY, "", "")
	} else {
		INSERT_QUERY = fmt.Sprintf(INSERT_QUERY, "slug, ", "'"+thread.Slug+"', ")
	}

	layout := time.RFC3339Nano
	t, err := time.Parse(layout, thread.Created)
	if err != nil {
		fmt.Println(err)
	}

	thread.Created = t.UTC().Format(time.RFC3339Nano)

	// fmt.Println("AddThread: since: ", thread.Created)

	// fmt.Println("AddThread: INSERT_QUERY: ", INSERT_QUERY)

	err = ts.db.QueryRow(INSERT_QUERY, thread.Author, thread.Forum,
		thread.Created, thread.Title, thread.Message, thread.Votes).Scan(&thread.ID)

	if err != nil {
		fmt.Println("AddThread:  error after id:", err.Error())
		panic(err)
	}

	insertQueryForumUsers :=
		"insert into forum_users (forum, username) values ($2, $1) ON conflict (forum, username) do nothing;"

	resultRows := ts.db.QueryRow(insertQueryForumUsers, thread.Author, thread.Forum)

	if err := resultRows.Scan(); err != nil && err != pgx.ErrNoRows {
		// TODO: move conflicts
		panic(err)
	}

	// fmt.Println("AddForum: id:", thread.ID)

	return true, thread
}

func (ts *ThreadService) UpdateThread(thread *models.Thread) *models.Thread {

	update := "update threads SET title = $2, message = $3 WHERE id = $1;"

	resultRows := ts.db.QueryRow(update, thread.ID, thread.Title, thread.Message)

	if err := resultRows.Scan(); err != nil && err != pgx.ErrNoRows {
		// TODO: move conflicts
		panic(err)
	}

	return thread
}

func (ts *ThreadService) SelectThreads(slug, limit, since string, desc bool) (bool, []models.Thread) {

	limitEndStr := "offsetStr+order+limitEndStr,"
	if limit != "" {
		limitEndStr = " LIMIT " + limit + ""
	}

	comp := " >= "
	order := "ORDER BY th.created "
	if desc {
		comp = " <= "
		order += " DESC "
	} else {
		order += " ASC "
	}

	offsetStr := ""
	if since != "" {
		t, err := time.Parse(time.RFC3339Nano, since)
		if err != nil {
			fmt.Println(err)
		}
		since = t.UTC().Format(time.RFC3339Nano)
		// fmt.Println("SelectThreads: since:", since)
		offsetStr = " AND th.created " + comp + " '" + since + "'"
	}

	tx, err := ts.db.DataBase().Begin()
	if err != nil {
		panic(err)
	}

	forumExists := false
	var threadsCount uint64
	rows, err := tx.Query("SELECT slug::text, threads FROM forums WHERE LOWER(slug) = LOWER($1);", slug)
	if err != nil {
		tx.Rollback()
		panic(err)
	}
	if rows.Next() {
		var foundSlug string
		if err := rows.Scan(&foundSlug, &threadsCount); err != nil {
			tx.Rollback()
			panic(err)
		}
		// fmt.Println("get_forum_slug_threads: threads:", threads)
		forumExists = true
	}

	rows.Close()

	if !forumExists {
		tx.Commit()
		return false, nil
	}

	rows, err = tx.Query(
		fmt.Sprintf(
			"SELECT id, coalesce(slug::text, ''), author::text, forum::text, created, title::text, message::text, votes FROM threads th WHERE LOWER(th.forum) = LOWER($1) %s %s %s;",
			offsetStr,
			order,
			limitEndStr,
		),
		slug,
	)
	defer rows.Close()
	if err != nil {
		tx.Rollback()
		panic(err)
	}

	threads := make([]models.Thread, 0, threadsCount)

	for rows.Next() {
		var thread models.Thread
		var selectedTime time.Time
		err := rows.Scan(
			&thread.ID,
			&thread.Slug,
			&thread.Author,
			&thread.Forum,
			&selectedTime,
			&thread.Title,
			&thread.Message,
			&thread.Votes,
		)
		thread.Created = selectedTime.UTC().Format(time.RFC3339Nano)
		if err != nil {
			tx.Rollback()
			panic(err)
		}
		threads = append(threads, thread)
	}

	tx.Commit()

	if len(threads) == 0 {
		return true, threads
	}

	return true, threads
}

func (ts *ThreadService) GetThreadBySlug(slug string) *models.Thread {

	// fmt.Println("GetThreadBySlug: query start")

	query := fmt.Sprintf(
		"SELECT id, slug::text, author::text, forum::text, created, title::text, message::text, votes FROM threads WHERE LOWER(slug) = LOWER('%s');",
		slug)

	// fmt.Println("GetThreadBySlug: query:", query)
	// fmt.Println("-----------------------------start------------------------------####################")

	rows := ts.db.Query(query)
	defer rows.Close()
	// fmt.Println("------------------------------end-------------------------------####################")

	for rows.Next() {
		thread := new(models.Thread)
		var selectedTime time.Time
		err := rows.Scan(
			&thread.ID,
			&thread.Slug,
			&thread.Author,
			&thread.Forum,
			&selectedTime,
			&thread.Title,
			&thread.Message,
			&thread.Votes,
		)
		thread.Created = selectedTime.UTC().Format(time.RFC3339Nano)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
		return thread
	}
	return nil
}

func (ts *ThreadService) GetThreadIDBySlugOrId(slugOrID string) (uint64, bool) {

	threadId, err := strconv.ParseUint(slugOrID, 10, 64)

	var rows *pgx.Rows
	if err == nil {
		rows = ts.db.Query("SELECT id FROM threads WHERE id = $1;", threadId)
	} else {
		rows = ts.db.Query("SELECT id FROM threads WHERE lower(slug) = lower('$1');", slugOrID)
	}
	defer rows.Close()

	for rows.Next() {
		var id uint64
		err := rows.Scan(&id)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
		return id, true
	}

	return 0, false
}

func (ts *ThreadService) GetThreadById(id uint64) *models.Thread {

	query := fmt.Sprintf(
		"SELECT id, coalesce(slug::text, ''), author::text, forum::text, created, title::text, message::text, votes FROM %s WHERE id = %s;",
		ts.tableName, strconv.FormatUint(id, 10))

	rows := ts.db.Query(query)
	defer rows.Close()

	for rows.Next() {
		thread := new(models.Thread)
		var selectedTime time.Time
		err := rows.Scan(
			&thread.ID,
			&thread.Slug,
			&thread.Author,
			&thread.Forum,
			&selectedTime,
			&thread.Title,
			&thread.Message,
			&thread.Votes,
		)
		thread.Created = selectedTime.UTC().Format(time.RFC3339Nano)
		if err != nil {
			fmt.Println("GetThreadBySlug: query:", query)
			fmt.Println(err)
			panic(err)
		}
		return thread
	}

	return nil
}

func (ts *ThreadService) Vote(thread *models.Thread, vote models.Vote) *models.Thread {

	addVoteStr := "+ 1"
	if vote.Voice == -1 {
		addVoteStr = "- 1"
	}

	tx, err := ts.db.DataBase().Begin()
	if err != nil {
		panic(err)
	}

	_, err = tx.Prepare(
		"get_vote",
		"SELECT id, voice FROM votes WHERE lower(username) = lower($1) AND thread = $2;",
	)

	_, err = tx.Prepare(
		"insert_vote",
		"INSERT INTO votes (username, voice, thread) VALUES ($1, $2, $3) returning id;",
	)

	_, err = tx.Prepare(
		"update_vote",
		"UPDATE votes SET voice = $1 WHERE id = $2;",
	)

	voice := new(int32)
	voteId := new(uint64)

	rows, err := tx.Query("get_vote", vote.Nickname, thread.ID)
	if err != nil {
		tx.Rollback()
		fmt.Println("Vote:  error:", err.Error())
		panic(err)
	}

	if rows.Next() {
		err := rows.Scan(voteId, voice)
		if err != nil {
			tx.Rollback()
			panic(err)
		}
	} else {
		voice, voteId = nil, nil
	}

	rows.Close()

	if voice != nil {
		if *voice == vote.Voice {
			return thread
		}

		_, err := tx.Exec("update_vote", vote.Voice, *voteId)
		if err != nil {
			tx.Rollback()
			fmt.Println("Vote:  error:", err.Error())
			panic(err)
		}

		addVoteStr = "+ 2"
		if vote.Voice == -1 {
			addVoteStr = "- 2"
		}
	} else {

		var id uint64
		err := tx.QueryRow("insert_vote", vote.Nickname, vote.Voice, thread.ID).Scan(&id)
		if err != nil {
			tx.Rollback()
			fmt.Println("Vote:  error:", err.Error())
			panic(err)
		}
	}

	err = tx.QueryRow("UPDATE threads SET votes = votes "+addVoteStr+" WHERE id = $1 returning votes;", thread.ID).Scan(&thread.Votes)
	if err != nil {
		tx.Rollback()
		fmt.Println("Vote:  error:", err.Error())
		panic(err)
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		panic(err)
	}

	if tx.Status() != pgx.TxStatusCommitSuccess {
		fmt.Println("==============================================================")
	}

	return thread
}

func (ts *ThreadService) getVote(username string, threadId uint64) (*int32, *uint64) {

	query := "SELECT id, voice FROM votes WHERE lower(username) = lower($1) AND thread = $2;"

	rows := ts.db.Query(query, username, threadId)
	defer rows.Close()

	for rows.Next() {
		voice := new(int32)
		id := new(uint64)
		err := rows.Scan(id, voice)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
		return voice, id
	}

	return nil, nil
}
