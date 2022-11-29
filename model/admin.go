package model

import "gorm.io/gorm"

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
