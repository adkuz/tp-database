package services

import (
	"fmt"

	"github.com/Alex-Kuz/tp-database/src/models"
	"github.com/jackc/pgx"
)

type ForumService struct {
	db        *PostgresDatabase
	tableName string
}

func MakeForumService(pgdb *PostgresDatabase) ForumService {
	return ForumService{db: pgdb, tableName: "forums"}
}

func (fs *ForumService) TableName() string {
	return fs.tableName
}

func (fs *ForumService) GetForumBySlug(slug string) *models.Forum {
	query := fmt.Sprintf(
		"SELECT slug, author, title, threads, posts FROM forums WHERE LOWER(slug) = LOWER('%s')", slug)

	rows := fs.db.Query(query)
	defer rows.Close()

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

func (fs *ForumService) SlugBySlug(slug string) *string {
	query := fmt.Sprintf(
		"SELECT slug FROM forums WHERE LOWER(slug) = LOWER('%s')", slug)

	rows := fs.db.Query(query)
	defer rows.Close()

	for rows.Next() {
		str := new(string)
		err := rows.Scan(str)
		if err != nil {
			panic(err)
		}
		return str
	}
	return nil
}

func (fs *ForumService) IncThreadsCountBySlug(slug string) bool {
	UPDATE_QUERY := "UPDATE forums SET threads = threads + 1 WHERE LOWER(slug) = LOWER($1);"

	resultRows := fs.db.QueryRow(UPDATE_QUERY, slug)

	if err := resultRows.Scan(); err != nil && err != pgx.ErrNoRows {
		panic(err)
	}

	// fmt.Println("")
	// fmt.Println("IncThreadsCountBySlug:  slug =", slug)

	return true
}

func (fs *ForumService) AddForum(forum *models.Forum) (bool, *models.Forum) {

	if conflictForum := fs.GetForumBySlug(forum.Slug); conflictForum != nil {
		return false, conflictForum
	}

	INSERT_QUERY := "insert into forums (slug, author, title, threads, posts) values ($1, $2, $3, $4, $5);"

	resultRows := fs.db.QueryRow(INSERT_QUERY, forum.Slug, forum.User, forum.Title, forum.Threads, forum.Posts)

	if err := resultRows.Scan(); err != nil && err != pgx.ErrNoRows {
		panic(err)
	}

	// fmt.Println("AddForum:  forum.User =", forum.User)

	return true, forum
}

func (fs *ForumService) GetUsers(forum *models.Forum, since, limit string, desc bool) []models.User {

	sinceStr := ""
	if since != "" {
		sinceStr = " AND LOWER(u.nickname) "
		if desc {
			sinceStr += "< "
		} else {
			sinceStr += "> "
		}
		sinceStr += "LOWER('" + since + "')"
	}

	order := " ASC"
	if desc {
		order = " DESC"
	}

	limitStr := ""
	if limit != "" {
		limitStr = " LIMIT " + limit
	}

	query := fmt.Sprintf(
		"SELECT nickname, fullname, about, email FROM users u JOIN forum_users uf ON LOWER(u.nickname) = LOWER(uf.username)"+
			" WHERE LOWER(uf.forum) = LOWER('%s') %s ORDER BY LOWER(u.nickname) %s %s;",
		forum.Slug, sinceStr, order, limitStr)

	// fmt.Println("GetUsers: QUERY:", query)

	rows := fs.db.Query(query)
	defer rows.Close()

	users := make([]models.User, 0)

	for rows.Next() {
		//var parent uint64
		var user models.User
		err := rows.Scan(
			&user.Nickname,
			&user.Fullname,
			&user.About,
			&user.Email,
		)
		fmt.Println(user)
		if err != nil {
			panic(err)
		}

		users = append(users, user)
	}
	return users
}
