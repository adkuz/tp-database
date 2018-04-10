package services

import (
	"fmt"

	"github.com/Alex-Kuz/tp-database/src/models"
)



type UserService struct {
	db        *PostgresDatabase
	tableName string
}

func MakeUserService(pgdb *PostgresDatabase) UserService {
	return UserService{db: pgdb, tableName: "users"}
}

func (uc *UserService) GetUserByNickname(nickname string) *models.User {
	query := fmt.Sprintf(
		"SELECT about, email, fullname, nickname FROM %s WHERE nickname = '%s'",
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

func (uc *UserService) AddUser(user *models.User) (bool, []models.User) {
	var users []models.User

	if user := uc.GetUserByNickname(user.Nickname); user != nil {
		users = append(users, *user)
	}

	if user := uc.GetUserByEmail(user.Email); user != nil {
		if len(users) == 0 || len(users) != 0 && users[0] != *user{
			users = append(users, *user)
		}
	}

	if len(users) > 0 {
		return false, users
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

