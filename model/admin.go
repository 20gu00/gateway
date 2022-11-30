package model

import (
	"gorm.io/gorm"
	"time"
)

type Admin struct {
	gorm.Model
	//ID       int    `gorm:"primary_key" json:"id"`
	UserName string `json:"username"`
	Salt     string `json:"salt"`
	Password string `json:"password"`
	Token    string `gorm:"type:varchar(500)" json:"token"`
	IsDelete int    `json:"is_delete"`
}

func (*Admin) TableName() string {
	return "admin"
}

type AdminSessionInfo struct {
	ID        int       `json:"id"`
	UserName  string    `json:"user_name"`
	LoginTime time.Time `json:"login_time"`
}

type AdminInfoOutput struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	LoginTime time.Time `json:"login_time"`
	Avatar    string    `json:"avatar"`
	//Introduction string    `json:"introduction"`
	//Roles        []string  `json:"roles"`
}
