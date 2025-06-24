package models

import "time"

type User struct {
	ID         int64  `json:"-"`
	Login      string `json:"login" validate:"required,min=2,max=20"`
	PassHash   string `json:"password" validate:"required,min=5,max=100"`
	DeviceID   int64  `json:"-"`
	DeviceName string `json:"-"`
	Salt       string `json:"-"`
}

type AuthResponse struct {
	Access  string `json:"access_token"`
	Refresh string `json:"refresh_token"`
	Salt    string `json:"salt"`
}

type Item struct {
	UID         string    `json:"uid"`
	Type        string    `json:"type"`
	Data        []byte    `json:"data"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
	Description string    `json:"description"`
}

type Login struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
type Card struct {
	Number string `json:"number"`
	Expiry string `json:"expiry"`
	CCV    string `json:"ccv"`
}
type File struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Path string `json:"path"`
}
