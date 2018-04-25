package services



type PostService struct {
	db        *PostgresDatabase
	tableName string
}

func MakePostService(pgdb *PostgresDatabase) PostService {
	return PostService{db: pgdb, tableName: "posts"}
}




