package services

import (
	"fmt"
	"github.com/Alex-Kuz/tp-database/src/models"
)



type UserService struct {
	db        *PostgresDatabase
	tableName string
}





func remove(slice []int, s int) []int {
	return append(slice[:s], slice[s+1:]...)
}

func MakeUserService(pgdb *PostgresDatabase) UserService {
	return UserService{db: pgdb, tableName: "users"}
}


func (uc *UserService) GetUserIDByNickname(nickname string) *uint64 {
	query := fmt.Sprintf(
		"SELECT id FROM %s WHERE LOWER(nickname) = LOWER('%s')",
		uc.tableName, nickname)

	rows := uc.db.Query(query)

	for rows.Next() {
		userId := new(uint64)
		err := rows.Scan(&userId)
		if err != nil {
			panic(err)
		}
		return userId
	}
	return nil
}

func (uc *UserService) GetUserByNickname(nickname string) *models.User {
	query := fmt.Sprintf(
		"SELECT about, email, fullname, nickname FROM %s WHERE LOWER(nickname) = LOWER('%s')",
		uc.tableName, nickname)

	rows := uc.db.Query(query)

	for rows.Next() {
		user := new(models.User)
		err := rows.Scan(&user.About, &user.Email, &user.Fullname, &user.Nickname)
		if err != nil {
			panic(err)
		}
		return user
	}
	return nil
}

func (uc *UserService) GetUserByEmail(email string) *models.User {
	query := fmt.Sprintf(
		"SELECT about, email, fullname, nickname FROM %s WHERE email = '%s'",
		uc.tableName, email)

	rows := uc.db.Query(query)

	for rows.Next() {
		user := new(models.User)
		err := rows.Scan(&user.About, &user.Email, &user.Fullname, &user.Nickname)
		if err != nil {
			panic(err)
		}
		return user
	}
	return nil
}

func (uc *UserService) GetUsersByEmailOrNick(email, nickname string) []models.User {
	users := make([]models.User, 0)

	query := fmt.Sprintf(
		"SELECT about, email, fullname, nickname FROM %s WHERE LOWER(email) = LOWER('%s') OR LOWER(nickname) = LOWER('%s')",
			uc.tableName, email, nickname)

	rows := uc.db.Query(query)

	for rows.Next() {
		user := new(models.User)
		err := rows.Scan(&user.About, &user.Email, &user.Fullname, &user.Nickname)
		if err != nil {
			panic(err)
		}

		users = append(users, *user)
	}
	return users
}

func (uc *UserService) AddUser(user *models.User) (bool, []models.User) {
	conflictUsers := uc.GetUsersByEmailOrNick(user.Email, user.Nickname)

	if len(conflictUsers) == 2 && conflictUsers[0] == conflictUsers[1] {
		conflictUsers = conflictUsers[:1]
	}

	if len(conflictUsers) > 0 {
		return false, conflictUsers
	}

	INSERT_QUERY :=
		"insert into " + uc.tableName + " (nickname, about, email, fullname) values ($1, $2, $3, $4);"

	insertQuery, err := uc.db.Prepare(INSERT_QUERY)
	if err != nil {
		panic(err)
	}
	defer insertQuery.Close()

	_, err = insertQuery.Exec(user.Nickname, user.About, user.Email, user.Fullname)
	if err != nil {
		panic(err)
	}

	return true, nil
}

func (uc *UserService) UpdateUser(user *models.User)  {


	fmt.Println("to update {about: ",user.About,
		", email: ", user.Email,
		", fullname: ", user.Fullname,
		", nickname: ", user.Nickname,
		"}")

	UPDATE_QUERY :=
		"update " + uc.tableName + " SET about = $2, email = $3, fullname = $4 " +
			"WHERE nickname = $1;"

	updateQuery, err := uc.db.Prepare(UPDATE_QUERY)
	if err != nil {
		panic(err)
	}
	defer updateQuery.Close()

	_, err = updateQuery.Exec(user.Nickname, user.About, user.Email, user.Fullname)
	if err != nil {
		panic(err)
	}

}

