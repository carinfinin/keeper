package models

type User struct {
	ID       int64  `json:"-"`
	Login    string `json:"login" validate:"required,min=2,max=20"`
	PassHash string `json:"password" validate:"required,min=5,max=100"`
}
