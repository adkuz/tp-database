package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"github.com/Alex-Kuz/tp-database/src/controllers"
	"github.com/Alex-Kuz/tp-database/src/router"
	"github.com/Alex-Kuz/tp-database/src/services"
	"github.com/gorilla/mux"
)

func doesNotImplements(responceWriter http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(responceWriter, "This method does not have implements.")
	fmt.Println("Endpoint Hit: homePage")
}

var (
	postgresConfig = services.Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "12345",
		DBName:   "forum_tp",
	}

	SchemaFile = "src/sql/dbscheme.sql"

	PostgresService services.PostgresDatabase

	ForumRouter *mux.Router = nil
)

func getParams(regEx, url string) (paramsMap map[string]string) {

	var compRegEx = regexp.MustCompile(regEx)
	match := compRegEx.FindStringSubmatch(url)

	paramsMap = make(map[string]string)
	for i, name := range compRegEx.SubexpNames() {
		if i > 0 && i <= len(match) {
			paramsMap[name] = match[i]
		}
	}
	return
}

func readConfig(dbLine *string) services.Config {

	fmt.Println(*dbLine)

	reg :=
		`(?P<db>[a-zA-Z][a-zA-Z0-9]*)://(?P<username>[a-zA-Z0-9_]+):(?P<password>[a-zA-Z0-9_]+)@(?P<host>[a-zA-Z][a-zA-Z0-9]*):(?P<port>[0-9]{4})/(?P<db_name>[a-zA-Z_][a-z_A-Z0-9]*)`
	paramsMap := getParams(reg, *dbLine)

	fmt.Println(paramsMap)

	port, err := strconv.ParseUint(paramsMap["port"], 10, 16)
	if err != nil {
		panic("Port error: not integer")
	}

	return services.Config{
		Host:     paramsMap["host"],
		Port:     uint16(port & 0xFFFF),
		User:     paramsMap["username"],
		Password: paramsMap["password"],
		DBName:   paramsMap["db_name"],
	}
}

func init() {

	fmt.Println("Connecting to database server...")

	if len(os.Args) > 1 {
		postgresConfig = readConfig(&(os.Args[1]))
	}

	fmt.Println("postgresConfig: ", postgresConfig)

	connectionConfig := services.MakeConnectionConfig(postgresConfig)
	fmt.Println("prepared postgresConfig: ", connectionConfig)
	PostgresService = services.Connect(connectionConfig)
	PostgresService.Setup(SchemaFile)

	fmt.Println("Successfuly connection")

	fmt.Println("Initialization API...")
	forumAPI := controllers.MakeForumAPI(&PostgresService)

	fmt.Println("Creating router...")
	ForumRouter = router.CreateRouter("/api", &forumAPI)
}

func main() {
	fmt.Println("Starting server...")

	log.Fatal(http.ListenAndServe(":5000", ForumRouter))
}
