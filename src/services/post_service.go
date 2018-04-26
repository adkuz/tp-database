package services

import (
	_ "fmt"

	"github.com/Alex-Kuz/tp-database/src/models"
)

type PostService struct {
	db        *PostgresDatabase
	tableName string
}

func MakePostService(pgdb *PostgresDatabase) PostService {
	return PostService{db: pgdb, tableName: "posts"}
}



func (ps *PostService) AddPost(post *models.Post) (bool, *models.Post) {

	/*INSERT_QUERY:=
		"insert into " + ps.tableName +
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
	*/
	return true, post
}



