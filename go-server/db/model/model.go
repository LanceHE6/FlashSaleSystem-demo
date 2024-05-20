package model

import "time"

type MyModel struct {
	CreatedAt time.Time `gorm:"autoCreateTime:milli"`
	UpdatedAt time.Time `gorm:"autoUpdateTime:milli"`
}

type Order struct {
	MyModel
	OrderId       string `gorm:"primary_key;index"`
	UserId        string
	OrderQuantity int
	OrderTime     string
}

type Goods struct {
	MyModel
	Gid      string `gorm:"primary_key;index"`
	Name     string
	Quantity int
}
