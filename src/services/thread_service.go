package services

type ThreadService struct {
	db        *PostgresDatabase
	tableName string
}

func MakeThreadService(pgdb *PostgresDatabase) ThreadService {
	return ThreadService{db: pgdb, tableName: "forums"}
}

/*
func (ts *ThreadService) AddThread(thread *models.Thread) (bool, *models.Thread) {

}
*/

