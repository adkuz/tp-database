package services

import (
	"fmt"
	"github.com/Alex-Kuz/tp-database/src/models"
)


type ForumService struct {
	db        *PostgresDatabase
	tableName string
}

func MakeForumService(pgdb *PostgresDatabase) ForumService {
	return ForumService{db: pgdb, tableName: "forums"}
}



func (fs *ForumService) GetForumBySlug(slug string) *models.Forum {
	query := fmt.Sprintf(
		"SELECT slug, author_id, title, threads, posts FROM %s WHERE slug = '%s'",
			fs.tableName, slug)

	rows := fs.db.Query(query)

	for rows.Next() {
		forum := new(models.Forum)
		err := rows.Scan(&forum.Slug, &forum.User, &forum.Title, &forum.Threads, &forum.Posts)
		if err != nil {
			panic(err)
		}
		return forum
	}
	return nil
}

func (fs *ForumService) AddForum(forum *models.Forum, authorId uint64) (bool, *models.Forum) {

	if conflictForum := fs.GetForumBySlug(forum.Slug); conflictForum != nil {
		return false, conflictForum
	}

	INSERT_QUERY :=
		"insert into " + fs.tableName + " (slug, author_id, title, threads, posts) values ($1, $2, $3, $4, $5);"

	insertQuery, err := fs.db.Prepare(INSERT_QUERY)
	defer insertQuery.Close()
	if err != nil {
		panic(err)
	}

	_, err = insertQuery.Exec(forum.Slug, authorId, forum.Title, forum.Threads, forum.Posts)
	if err != nil {
		panic(err)
	}

	return true, forum
}
