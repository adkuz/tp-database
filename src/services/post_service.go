package services

import (
	"fmt"
	_ "fmt"
	"strconv"
	"strings"

	"github.com/Alex-Kuz/tp-database/src/models"
)

type PostService struct {
	db        *PostgresDatabase
	tableName string
}

func MakePostService(pgdb *PostgresDatabase) PostService {
	return PostService{db: pgdb, tableName: "posts"}
}

type ParentThread struct {
	ParentID uint64
	Thread   uint64
}

func (ps *PostService) TableName() string {
	return ps.tableName
}

func (ps *PostService) RequiredParents(posts []models.Post) map[uint64]uint64 {

	parents := make(map[ParentThread]bool)

	for i := 0; i < len(posts); i++ {
		pt := ParentThread{posts[i].Parent, posts[i].Thread}
		parents[pt] = true
	}

	for i := 0; i < len(posts); i++ {
		// fmt.Println("\t", i, ": id, parent = ", posts[i].ID, ",", posts[i].Parent, ",", posts[i].Thread)

		for p := 0; p < len(posts); p++ {
			if posts[i].Parent == posts[p].ID {
				pt := ParentThread{posts[i].Parent, posts[i].Thread}
				parents[pt] = false
			}
		}
	}

	requiredParents := make(map[uint64]uint64)

	for pt, isRequired := range parents {
		if isRequired {
			requiredParents[pt.ParentID] = pt.Thread
		}
	}

	return requiredParents
}

func (ps *PostService) GetAllParents(threadId uint64,
	limit uint64, since string, desc bool) []uint64 {

	sinceStr := ""
	if since != "" {
		sinceStr = " AND tree_path[1] "
		if desc {
			sinceStr += "< "
		} else {
			sinceStr += "> "
		}
		sinceStr += "(SELECT p.tree_path[1] FROM posts p WHERE p.id = " + since + " )"
	}

	order := " ASC"
	if desc {
		order = " DESC"
	}

	limitStr := ""
	if limit != 0 {
		limitStr = " LIMIT " + strconv.FormatUint(limit, 10)
	}

	query := fmt.Sprintf(
		"SELECT id FROM posts WHERE thread = %s AND parent = 0 %s ORDER BY id %s%s;",
		strconv.FormatUint(threadId, 10), sinceStr, order, limitStr)

	// fmt.Println("GetAllParents: query:", query)

	rows := ps.db.Query(query)
	defer rows.Close()

	parents := make([]uint64, 0)
	for rows.Next() {
		var id uint64

		if err := rows.Scan(&id); err != nil {
			fmt.Println(err)
			panic(err)
		}

		parents = append(parents, id)
	}

	return parents
}

func (ps *PostService) GetPostById(id uint64) *models.Post {
	// fmt.Println("GetPostById: query start")

	query := fmt.Sprintf(
		"SELECT id, created, is_edited, parent, message, author, forum, thread, tree_path FROM posts WHERE id = %s;",
		strconv.FormatUint(id, 10))

	rows := ps.db.Query(query)
	defer rows.Close()

	for rows.Next() {
		post := new(models.Post)
		var tree_path string

		err := rows.Scan(&post.ID, &post.Created, &post.IsEdited, &post.Parent,
			&post.Message, &post.Author, &post.Forum, &post.Thread, &tree_path)

		if err != nil {
			fmt.Println(err)
			panic(err)
		}

		ids := strings.Split(tree_path[1:len(tree_path)-1], ",")
		for i := 0; i < len(ids) && ids[i] != ""; i++ {
			id, err := strconv.ParseUint(ids[i], 10, 64)
			if err != nil {
				panic(err)
			}
			post.Path = append(post.Path, id)
		}
		// fmt.Println("============================>", post.Path)
		return post
	}

	return nil
}

func (ps *PostService) AddPost(post *models.Post) (bool, *models.Post) {

	INSERT_QUERY :=
		"insert into posts (created, message, parent, author, forum, thread) values ($1, $2, $3, $4, $5, $6) returning id;"

	// fmt.Println("AddThread: INSERT_QUERY:", INSERT_QUERY)

	err := ps.db.QueryRow(INSERT_QUERY, post.Created, post.Message, post.Parent,
		post.Author, post.Forum, post.Thread).Scan(&post.ID)

	if err != nil {
		fmt.Println("AddPost:  error after id:", err.Error())
		panic(err)
	}

	// fmt.Println("AddPost: id:", post.ID)

	insertQueryForumUsers :=
		"insert into forum_users (username, forum) select $1, $2 " +
			"where not exists (select * from forum_users where lower(username) = lower($3) and lower(forum) = lower($4));"

	insertQueryUserForum, err := ps.db.Prepare(insertQueryForumUsers)
	defer insertQueryUserForum.Close()
	if err != nil {
		panic(err)
	}

	_, err = insertQueryUserForum.Exec(post.Author, post.Forum, post.Author, post.Forum)
	if err != nil {
		fmt.Println("AddForum:  error:", err.Error())
		panic(err)
	}

	return true, post
}

func (ps *PostService) GetPostsFlat(thread *models.Thread,
	limit, since string, desc bool) []models.Post {

	sinceStr := ""
	if since != "" {
		sinceStr = " AND p.id "
		if desc {
			sinceStr += "< "
		} else {
			sinceStr += "> "
		}
		sinceStr += since
	}

	order := ""
	if desc {
		order = " DESC"
	}

	limitStr := ""
	if limit != "" {
		limitStr = " LIMIT " + limit
	}

	query := fmt.Sprintf(
		"SELECT created, id, message, parent, author, forum, thread FROM posts p WHERE p.thread = %s%s ORDER BY p.created %s, p.id %s%s;",
		strconv.FormatUint(thread.ID, 10), sinceStr, order, order, limitStr)

	// fmt.Println("GetPostsFlat: QUERY:", query)

	rows := ps.db.Query(query)
	defer rows.Close()
	posts := make([]models.Post, 0)

	for rows.Next() {
		//var parent uint64
		var post models.Post
		err := rows.Scan(
			&post.Created,
			&post.ID,
			&post.Message,
			&post.Parent,
			&post.Author,
			&post.Forum,
			&post.Thread,
		)
		if err != nil {
			panic(err)
		}

		posts = append(posts, post)
	}

	return posts
}

func (ps *PostService) GetPostsTreeSort(thread *models.Thread,
	limit, since string, desc bool) []models.Post {

	sinceStr := ""
	if since != "" {
		sinceStr = " AND p.tree_path "
		if desc {
			sinceStr += "< "
		} else {
			sinceStr += "> "
		}
		sinceStr += "(SELECT tree_path FROM posts p WHERE p.id = " + since + " )"
	}

	order := ""
	if desc {
		order = " DESC"
	}

	limitStr := ""
	if limit != "" {
		limitStr = " LIMIT " + limit
	}

	query := fmt.Sprintf(
		"SELECT created, id, message, parent, author, forum, thread FROM posts p WHERE p.thread = %s%s ORDER BY p.tree_path %s, p.id %s%s;",
		strconv.FormatUint(thread.ID, 10), sinceStr, order, order, limitStr)

	// fmt.Println("GetPostsTreeSort: QUERY:", query)

	rows := ps.db.Query(query)
	defer rows.Close()
	posts := make([]models.Post, 0)

	for rows.Next() {
		//var parent uint64
		var post models.Post
		err := rows.Scan(
			&post.Created,
			&post.ID,
			&post.Message,
			&post.Parent,
			&post.Author,
			&post.Forum,
			&post.Thread,
		)
		if err != nil {
			panic(err)
		}

		posts = append(posts, post)
	}
	return posts
}

func (ps *PostService) GetPostsParentTreeSort(thread *models.Thread,
	limit, since string, desc bool) []models.Post {

	sinceStr := ""
	if since != "" {
		sinceStr = " AND tree_path[1] "
		if desc {
			sinceStr += "< "
		} else {
			sinceStr += "> "
		}
		sinceStr += "(SELECT p.tree_path[1] FROM posts p WHERE p.id = " + since + " )"
	}

	var count uint64 = 0
	if limit != "" {
		var err error
		count, err = strconv.ParseUint(limit, 10, 64)
		if err != nil {
			panic(err)
		}
	}

	parents := ps.GetAllParents(thread.ID, count, since, desc)
	// fmt.Println("GetPostsParentTreeSort: GetAllParents ->", parents)

	posts := make([]models.Post, 0)

	for i := 0; i < len(parents); i++ {

		query := fmt.Sprintf(
			"SELECT created, id, message, parent, author, forum, thread, tree_path FROM posts WHERE tree_path[1] = %s AND thread = %s%s ORDER BY tree_path, id;",
			strconv.FormatUint(parents[i], 10), strconv.FormatUint(thread.ID, 10), sinceStr)

		// fmt.Println("--> GetPostsTreeSort: QUERY:", query)
		// fmt.Println("<-- GetPostsTreeSort: posts:")
		rows := ps.db.Query(query)
		var path string

		for rows.Next() {
			var post models.Post
			err := rows.Scan(
				&post.Created,
				&post.ID,
				&post.Message,
				&post.Parent,
				&post.Author,
				&post.Forum,
				&post.Thread,
				&path,
			)
			if err != nil {
				panic(err)
			}
			fmt.Println("\t\t\t\t", ": parent, id, path:", post.Parent, post.ID, path)
			posts = append(posts, post)
		}
		rows.Close()
	}

	return posts
}

func (ps *PostService) UpdatePost(post *models.Post) *models.Post {
	update :=
		"update posts SET message = $2, is_edited = true WHERE id = $1;"

	updateQuery, err := ps.db.Prepare(update)
	if err != nil {
		panic(err)
	}
	defer updateQuery.Close()

	_, err = updateQuery.Exec(post.ID, post.Message)
	if err != nil {
		panic(err)
	}

	post.IsEdited = true

	return post
}

func (ps *PostService) CountOnForum(forum *models.Forum) uint64 {
	query := fmt.Sprintf("SELECT COUNT(*) FROM posts WHERE LOWER(forum) = LOWER('%s');", forum.Slug)
	rows := ps.db.Query(query)
	defer rows.Close()

	var count uint64 = 0

	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			fmt.Println(err)
			panic(err)
		}
	}

	return count
}
