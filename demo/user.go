package main

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	EnvNum   int    `json:"env_num"`
}

func CreateUser(username, password string) *User {
	user := User{
		Username: username,
		Password: password,
	}
	return &user
}

func (u *User) setEnvNum(num int) {
	u.EnvNum = num
}

func (c *User) Login(username, password string) bool {
	if username == c.Username && password == c.Password {
		return true
	}
	return false
}
