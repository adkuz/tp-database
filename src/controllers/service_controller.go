package controllers

import "net/http"

func ServiceStatus(respWriter http.ResponseWriter, request *http.Request) {
	respWriter.Header().Set("Content-Type", "application/json; charset=utf-8")

	status := make(map[string]uint64)

	count := func(tablename string) uint64 {

		rows := UserService.GetDB().Query("SELECT COUNT(*) FROM " + tablename)
		defer rows.Close()

		for rows.Next() {
			var count uint64
			err := rows.Scan(&count)
			if err != nil {
				panic(err)
			}
			return count
		}
		return 0
	}

	status["user"] = count(UserService.TableName())
	status["forum"] = count(ForumService.TableName())
	status["thread"] = count(ThreadService.TableName())
	status["post"] = count(PostService.TableName())

	respWriter.WriteHeader(http.StatusOK)
	writeJSONBody(&respWriter, status)
}

func ServiceClear(respWriter http.ResponseWriter, request *http.Request) {
	respWriter.Header().Set("Content-Type", "application/json; charset=utf-8")

	drop := func(tablename string) {
		// fmt.Println("--start dropping " + tablename)

		rows, err := UserService.GetDB().DataBase().Query("DELETE FROM " + tablename + ";")
		defer rows.Close()
		if err != nil {
			panic(err)
		} /* else {
			fmt.Println("++dropped " + tablename)
		}*/
	}

	drop("forum_users")
	drop("posts")
	drop("votes")
	drop("threads")
	drop("forums")
	drop("users")

	respWriter.WriteHeader(http.StatusOK)
}
