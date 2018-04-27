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

	for i := 0; i < len(posts); i++ {
		if parents[posts[i].Parent] {
			requiredParents = append(requiredParents, posts[i].Parent)
		}

	}

	fmt.Println("\trequiredParents:", requiredParents)
	fmt.Println(".......................................................................\n")

	return requiredParents
}

func (ps *PostService) GetPostById(id uint64) *models.Post {
	fmt.Println("GetPostById: query start")

	query := fmt.Sprintf(
		"SELECT id, created, is_edited, parent, author, forum, thread, tree_path FROM %s WHERE id = %s;",
		ps.tableName, strconv.FormatUint(id, 10))

	fmt.Println("GetThreadBySlug: query:", query)
	fmt.Println("-----------------------------start------------------------------####################")

	rows := ps.db.Query(query)
	fmt.Println("------------------------------end-------------------------------####################")

	for rows.Next() {
		post := new(models.Post)
		var tree_path string
		err := rows.Scan(&post.ID, &post.Created, &post.IsEdited, &post.Parent, &post.Author,
			&post.Forum, &post.Thread, &tree_path)
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

	return nil
}

func (ps *PostService) AddPost(post *models.Post) (bool, *models.Post) {

	INSERT_QUERY:=
		"insert into " + ps.tableName +
			" (created, parent, author, forum, thread)" +
			" values ($1, $2, $3, $4, $5) returning id;"


	fmt.Println("AddThread: INSERT_QUERY:", INSERT_QUERY)

	err := ps.db.QueryRow(INSERT_QUERY, post.Created, post.Parent,
		post.Author, post.Forum, post.Thread).Scan(&post.ID)

	if err != nil {
		fmt.Println("AddPost:  error after id:", err.Error())
		panic(err)
	}

	fmt.Println("AddPost: id:", post.ID)

	return true, post
}





