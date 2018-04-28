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


func (ps *PostService) RequiredParents(posts []models.Post) []uint64 {
	fmt.Println("\nRequiredParents.........................................................")

	//bans := make(map[uint64][]uint64)
	parents := make(map[uint64]bool)

	for i := 0; i < len(posts); i++ {
		parents[posts[i].Parent] = true
		//bans[posts[i].ID] = make([]uint64, 0)
	}

	for i := 0; i < len(posts); i++ {
		fmt.Println("\t", i, ": id, parent = ", posts[i].ID, ",", posts[i].Parent)

		for p := 0; p < len(posts); p++ {
			if posts[i].Parent == posts[p].ID {
				parents[posts[i].Parent] = false
			}
		}
	}

	requiredParents := make([]uint64, 0)

	for parent, isRequired := range parents {
		if isRequired {
			requiredParents = append(requiredParents, parent)
		}

	}

	fmt.Println("\trequiredParents:", requiredParents)
	fmt.Println(".......................................................................\n")

	return requiredParents
}

func (ps *PostService) GetPostById(id uint64) *models.Post {
	fmt.Println("GetPostById: query start")

	query := fmt.Sprintf(
		"SELECT id, created, is_edited, parent, message, author, forum, thread, tree_path FROM %s WHERE id = %s;",
		ps.tableName, strconv.FormatUint(id, 10))

	fmt.Println("GetThreadBySlug: query:", query)
	fmt.Println("-----------------------------start------------------------------####################")

	rows := ps.db.Query(query)
	fmt.Println("------------------------------end-------------------------------####################")

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
			id, err:= strconv.ParseUint(ids[i], 10, 64)
			if err != nil {
				panic(err)
			}
			post.Path = append(post.Path, id)
		}
		fmt.Println("============================>", post.Path)
		return post
	}

	rows.Close()

	return nil
}

func (ps *PostService) AddPost(post *models.Post) (bool, *models.Post) {

	INSERT_QUERY:=
		"insert into " + ps.tableName +
			" (created, message, parent, author, forum, thread)" +
			" values ($1, $2, $3, $4, $5, $6) returning id;"


	fmt.Println("AddThread: INSERT_QUERY:", INSERT_QUERY)

	err := ps.db.QueryRow(INSERT_QUERY, post.Created, post.Message, post.Parent,
		post.Author, post.Forum, post.Thread).Scan(&post.ID)

	if err != nil {
		fmt.Println("AddPost:  error after id:", err.Error())
		panic(err)
	}

	fmt.Println("AddPost: id:", post.ID)

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


	fmt.Println("GetPostsFlat: QUERY:", query)

	rows := ps.db.Query(query)
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

	rows.Close()

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


	fmt.Println("GetPostsTreeSort: QUERY:", query)

	rows := ps.db.Query(query)
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
	rows.Close()
	return posts
}

func (ps *PostService) GetPostsParentTreeSort(thread *models.Thread,
	limit, since string, desc bool) []models.Post {

	sinceStr := ""
	if since != "" {
		sinceStr = " AND p.tree_path[1] "
		if desc {
			sinceStr += "< "
		} else {
			sinceStr += "> "
		}
		sinceStr += "(SELECT tree_path[1] FROM posts p WHERE p.id = " + since + " )"
	}

	order := ""
	if desc {
		order = " DESC"
	}

	var count uint64 = 0
	if limit != "" {
		var err error
		count, err = strconv.ParseUint(limit, 10, 64)
		if err != nil {
			panic(err)
		}
	}



	query := fmt.Sprintf(
		"SELECT created, id, message, parent, author, forum, thread FROM posts p WHERE p.thread = %s%s ORDER BY p.tree_path[1] %s, p.id;",
		strconv.FormatUint(thread.ID, 10), sinceStr, order)


	fmt.Println("GetPostsTreeSort: QUERY:", query)

	rows := ps.db.Query(query)
	posts := make([]models.Post, 0)

	parents := make([]uint64, 0)
	for rows.Next() && (count == 0 || count != 0 && uint64(len(parents)) < count) {
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

		if len(parents) != 0 {
			if parents[len(parents)-1] == post.Parent {
				parents = append(parents, post.Parent)
			}
		} else {
			parents = append(parents, post.Parent)
		}

		posts = append(posts, post)
	}

	rows.Close()
	return posts
}



