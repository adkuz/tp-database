package services

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Alex-Kuz/tp-database/src/models"
	"github.com/jackc/pgx"
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

func (ps *PostService) RequiredParents(posts []models.Post) []uint64 {

	parents := make(map[uint64]bool)

	for i := 0; i < len(posts); i++ {
		parents[posts[i].Parent] = true
	}

	for i := 0; i < len(posts); i++ {
		for p := 0; p < len(posts); p++ {
			if posts[i].Parent == posts[p].ID {
				parents[posts[i].Parent] = false
			}

		}
	}

	requiredParents := make([]uint64, 0)

	for id, isRequired := range parents {
		if isRequired {
			requiredParents = append(requiredParents, id)
		}
	}

	return requiredParents
}

func (ps *PostService) GetAllParents(threadId uint64, limit uint64, since string, desc bool) []uint64 {

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

	query := "SELECT id, created, is_edited, parent, message::text, author::text, forum::text, thread, tree_path FROM posts WHERE id = $1;"

	rows := ps.db.Query(query, strconv.FormatUint(id, 10))
	defer rows.Close()

	for rows.Next() {
		post := new(models.Post)
		var selectedTime time.Time
		err := rows.Scan(&post.ID, &selectedTime, &post.IsEdited, &post.Parent,
			&post.Message, &post.Author, &post.Forum, &post.Thread, &post.Path)

		post.Created = selectedTime.UTC().Format(time.RFC3339Nano)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}

		return post
	}

	return nil
}

type uint64Array []uint64

func (a uint64Array) String() (s string) {
	sep := ""
	for _, el := range a {
		s += sep
		sep = ", "
		s += fmt.Sprintf("%d", el)
	}
	return
}

func (ps *PostService) GetPostsParentInfoByIdsArray(idArray []uint64) map[uint64]uint64 {

	idToThread := make(map[uint64]uint64)

	if len(idArray) == 0 {
		return idToThread
	}

	query := "SELECT id, thread FROM posts WHERE id = ANY(ARRAY[" + uint64Array(idArray).String() + "]::BIGINT[]);"

	// println(query)
	rows := ps.db.Query(query)
	defer rows.Close()

	for rows.Next() {
		var id uint64
		var thread uint64
		err := rows.Scan(&id, &thread)
		if err != nil {
			fmt.Println("GetPostsParentInfoByIdsArray: error !")
			panic(err)
		}

		idToThread[id] = thread
		// fmt.Println(idToThread)
	}

	return idToThread
}

/*
func (ps *PostService) AddPost(post *models.Post) (bool, *models.Post) {

	INSERT_QUERY :=
		"insert into posts (created, message, parent, author, forum, thread) values ($1, $2, $3, $4, $5, $6) returning id;"

	err := ps.db.QueryRow(INSERT_QUERY, post.Created, post.Message, post.Parent, post.Author, post.Forum, post.Thread).Scan(&post.ID)

	if err != nil {
		fmt.Println("AddPost:  error after id:", err.Error())
		panic(err)
	}

	// fmt.Println("Created post id=", post.ID, ", parent = ", post.Parent)

	insertQueryForumUsers :=
		"insert into forum_users (forum, username) values ($2, $1) ON conflict (forum, username) do nothing;"

	resultRows := ps.db.QueryRow(insertQueryForumUsers, post.Author, post.Forum)

	if err := resultRows.Scan(); err != nil && err != pgx.ErrNoRows {
		// TODO: move conflicts
		panic(err)
	}

	return true, post
}
*/

func (ps *PostService) AddSomePosts(posts models.PostsArray, requiredParents []uint64) (bool, models.PostsArray) {

	fmt.Print("(")
	defer fmt.Print(")")

	addedPostsArr := make(models.PostsArray, 0, len(posts))

	tx, err := ps.db.DataBase().Begin()
	if err != nil {
		panic(err)
	}

	treadId := posts[0].Thread

	_, err = tx.Prepare(
		"insert_posts",
		"insert into posts (created, message, parent, author, forum, thread) values ($1, $2, $3, $4, $5, $6) returning id;",
	)
	/*
		_, err = tx.Prepare(
			"insert_forum_users",
			"insert into forum_users (forum, username) values ($2, $1) ON conflict (forum, username) do nothing;",
		)
	*/

	if len(requiredParents) != 0 {
		fmt.Println("SELECT id, thread FROM posts WHERE id = ANY(ARRAY[" + uint64Array(requiredParents).String() + "]::BIGINT[]);")

		rows, err := tx.Query("SELECT id, thread FROM posts WHERE id = ANY(ARRAY[" + uint64Array(requiredParents).String() + "]::BIGINT[]);")
		if err == nil {
			anyResults := false
			for rows.Next() {
				var id, thread uint64
				if rows.Scan(&id, &thread); err != nil {
					if err == pgx.ErrNoRows {
						tx.Rollback()
						return false, nil
					}
					panic(err)
				}
				if thread != treadId {
					tx.Rollback()
					return false, nil
				}
				anyResults = true
				fmt.Printf(" %d->%d", id, thread)
			}
			if !anyResults {
				tx.Rollback()
				return false, nil
			}
		} else {
			panic(err)
		}
	}

	for i := 0; i < len(posts); i++ {
		post := posts[i]
		row := tx.QueryRow("insert_posts", post.Created, post.Message, post.Parent, post.Author, post.Forum, post.Thread)
		fmt.Printf("on read: %d: status = %d", len(posts), tx.Status())
		err := row.Scan(&posts[i].ID)
		if err != nil {
			tx.Rollback()
			panic(err)
		}
		addedPostsArr = append(addedPostsArr, posts[i])
	}

	/*
		for i := 0; i < len(posts); i++ {
			post := &posts[i]
			_, err := tx.Exec("insert_forum_users", post.Author, post.Forum)
			if err != nil {
				tx.Rollback()
				panic(err)
			}
		}
	*/

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		panic(err)
	}

	fmt.Println("end: status = ", tx.Status())
	if tx.Status() != pgx.TxStatusCommitSuccess {
		fmt.Println("==============================================================")
	}

	return true, addedPostsArr
}

func (ps *PostService) GetPostsFlat(thread *models.Thread, limit, since string, desc bool) []models.Post {

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
		"SELECT created, id, message::text, parent, author::text, forum::text, thread FROM posts p WHERE p.thread = %s%s ORDER BY p.created %s, p.id %s%s;",
		strconv.FormatUint(thread.ID, 10), sinceStr, order, order, limitStr)

	rows := ps.db.Query(query)
	defer rows.Close()
	posts := make([]models.Post, 0)

	for rows.Next() {
		//var parent uint64
		var post models.Post
		var selectedTime time.Time

		err := rows.Scan(
			&selectedTime,
			&post.ID,
			&post.Message,
			&post.Parent,
			&post.Author,
			&post.Forum,
			&post.Thread,
		)
		post.Created = selectedTime.UTC().Format(time.RFC3339Nano)
		if err != nil {
			panic(err)
		}

		posts = append(posts, post)
	}

	return posts
}

func (ps *PostService) GetPostsTreeSort(thread *models.Thread, limit, since string, desc bool) []models.Post {

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
		"SELECT created, id, message::text, parent, author::text, forum::text, thread FROM posts p WHERE p.thread = %s%s ORDER BY p.tree_path %s, p.id %s%s;",
		strconv.FormatUint(thread.ID, 10), sinceStr, order, order, limitStr)

	// fmt.Println("GetPostsTreeSort: QUERY:", query)

	rows := ps.db.Query(query)
	defer rows.Close()
	posts := make([]models.Post, 0)

	for rows.Next() {
		//var parent uint64
		var post models.Post
		var selectedTime time.Time
		err := rows.Scan(
			&selectedTime,
			&post.ID,
			&post.Message,
			&post.Parent,
			&post.Author,
			&post.Forum,
			&post.Thread,
		)
		post.Created = selectedTime.UTC().Format(time.RFC3339Nano)

		if err != nil {
			panic(err)
		}

		posts = append(posts, post)
	}
	return posts
}

func (ps *PostService) GetPostsParentTreeSort(thread *models.Thread, limit, since string, desc bool) []models.Post {

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

	posts := make([]models.Post, 0)

	for i := 0; i < len(parents); i++ {

		/*
			three_path
			three_path[1]
			thread
		*/

		query := fmt.Sprintf(
			"SELECT created, id, message::text, parent, author::text, forum::text, thread FROM posts WHERE tree_path[1] = %s AND thread = %s%s ORDER BY tree_path, id;",
			strconv.FormatUint(parents[i], 10), strconv.FormatUint(thread.ID, 10), sinceStr)

		rows := ps.db.Query(query)

		for rows.Next() {
			var post models.Post
			var selectedTime time.Time
			err := rows.Scan(
				&selectedTime,
				&post.ID,
				&post.Message,
				&post.Parent,
				&post.Author,
				&post.Forum,
				&post.Thread,
			)
			post.Created = selectedTime.UTC().Format(time.RFC3339Nano)
			if err != nil {
				panic(err)
			}
			posts = append(posts, post)
		}
		rows.Close()
	}

	return posts
}

func (ps *PostService) UpdatePost(post *models.Post) *models.Post {
	update := "update posts SET message = $2, is_edited = true WHERE id = $1;"

	resultRows := ps.db.QueryRow(update, post.ID, post.Message)
	if err := resultRows.Scan(); err != nil && err != pgx.ErrNoRows {
		// TODO: move conflicts
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
