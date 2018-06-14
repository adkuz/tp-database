package services

import (
	"fmt"

	"github.com/Alex-Kuz/tp-database/src/models"
	"github.com/jackc/pgx"
)

type UserService struct {
	db        *PostgresDatabase
	tableName string
}

const (
	insertUserQuery = "insert into users (nickname, about, email, fullname) values ($1, $2, $3, $4);"
	updateUserQuery = "UPDATE users SET about = $2, email = $3, fullname = $4  WHERE LOWER(nickname) = LOWER($1);"
)

func remove(slice []int, s int) []int {
	return append(slice[:s], slice[s+1:]...)
}

func MakeUserService(pgdb *PostgresDatabase) UserService {
	return UserService{db: pgdb, tableName: "users"}
}

func (uc *UserService) GetDB() *PostgresDatabase {
	return uc.db
}

func (us *UserService) TableName() string {
	return us.tableName
}

func (uc *UserService) GetUserIDByNickname(nickname string) *string {

	// fmt.Println("UserService::GetUserIDByNickname:  nickname = '", nickname, "'")

	query := "SELECT nickname::text FROM users WHERE LOWER(nickname) = LOWER($1)"

	rows := uc.db.Query(query, nickname)
	defer rows.Close()

	for rows.Next() {
		nickname := new(string)
		err := rows.Scan(&nickname)
		if err != nil {
			panic(err)
		}
		return nickname
	}
	return nil
}

func (uc *UserService) GetUserByNickname(nickname string) *models.User {
	query := "SELECT about::text, email::text, fullname::text, nickname::text FROM users WHERE LOWER(nickname) = LOWER($1)"

	rows := uc.db.Query(query, nickname)
	defer rows.Close()

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

type nicknameSet map[string]bool

func (set nicknameSet) String() (s string) {
	sep := ""
	for str, _ := range set {
		s += sep
		sep = ", "
		s += fmt.Sprintf("LOWER('%s')", str)
	}
	return
}

func (uc *UserService) GetUsersByNicknamesArray(nicknames map[string]bool) []string {

	query := "SELECT nickname::text FROM users WHERE LOWER(nickname) = ANY (ARRAY[" + nicknameSet(nicknames).String() + "])"

	// fmt.Println(nicknames)
	// fmt.Println(query)

	rows := uc.db.Query(query)
	defer rows.Close()

	nicksArray := make([]string, 0, len(nicknames))

	for rows.Next() {
		var nick string
		err := rows.Scan(&nick)
		if err != nil {
			panic(err)
		}
		nicksArray = append(nicksArray, nick)
	}
	return nicksArray
}

func (uc *UserService) GetUserByEmail(email string) *models.User {
	query := "SELECT about::text, email::text, fullname::text, nickname::text FROM users WHERE lower(email) = lower($1)"

	rows := uc.db.Query(query, email)
	defer rows.Close()

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

	query := "SELECT about::text, email::text, fullname::text, nickname::text FROM users WHERE LOWER(email) = LOWER($1) OR LOWER(nickname) = LOWER($2)"

	resultRows := uc.db.Query(query, email, nickname)
	defer resultRows.Close()

	for resultRows.Next() {
		user := new(models.User)
		err := resultRows.Scan(&user.About, &user.Email, &user.Fullname, &user.Nickname)
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

	resultRows := uc.db.QueryRow(insertUserQuery, user.Nickname, user.About, user.Email, user.Fullname)

	if err := resultRows.Scan(); err != nil && err != pgx.ErrNoRows {
		// TODO: move conflicts
		panic(err)
	}

	return true, nil
}

func (uc *UserService) UpdateUser(user *models.User) {

	resultRows := uc.db.QueryRow(updateUserQuery, user.Nickname, user.About, user.Email, user.Fullname)

	if err := resultRows.Scan(); err != nil && err != pgx.ErrNoRows {
		// TODO: move conflicts
		panic(err)
	}
}
