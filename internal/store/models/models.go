package models

import "time"

// User.
type User struct {
	ID         int64  `json:"-"`
	Login      string `json:"login" validate:"required,min=2,max=20"`
	PassHash   string `json:"password" validate:"required,min=5,max=100"`
	DeviceID   int64  `json:"-"`
	DeviceName string `json:"-"`
	Salt       string `json:"-"`
}

// AuthResponse возвращается с сервера в роутах
// - /auth
// - /register
// - /refresh.
type AuthResponse struct {
	Access  string `json:"access_token"`
	Refresh string `json:"refresh_token"`
	Salt    string `json:"salt"`
}

// Item представляет запись зашифрованных данных.
type Item struct {
	UID         string    `json:"uid"`
	Type        string    `json:"type"`
	Data        []byte    `json:"data"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
	Description string    `json:"description"`
	IsDeleted   bool      `json:"is_deleted"`
}

// Login передаётся насервер в роутах
// - /auth
// - /register
// так же при сохранени в локальную бд.
type Login struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Card.
type Card struct {
	Number string `json:"number"`
	Expiry string `json:"expiry"`
	CCV    string `json:"ccv"`
}

// File.
type File struct {
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	Content []byte `json:"content"`
}

// Token.
type Token struct {
	Access  string `json:"access_token"`
	Refresh string `json:"refresh_token"`
}

// LastSync передаётся с сервера в роуте
// - /last_sync
type LastSync struct {
	Update time.Time `json:"update_at"`
}
