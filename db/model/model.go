package model

type Order struct {
	OrderId       string `gorm:"primaryKey;index;unique"`
	UserId        string
	OrderTime     string
	OrderQuantity int
}

type Goods struct {
	Gid      string `gorm:"primaryKey;index;unique"`
	Name     string
	Quantity int
}
