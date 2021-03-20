package database

type Account string

func NewAccount(s string) Account {
	return Account(s)
}
