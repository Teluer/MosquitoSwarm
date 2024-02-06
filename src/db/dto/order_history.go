package dto

import "time"

type OrderHistory struct {
	Name          string `gorm:"type:varchar(100)"`
	Phone         string `gorm:"type:varchar(30)"`
	ItemId        string `gorm:"type:varchar(50)"`
	OrderDateTime time.Time
}

func (OrderHistory) TableName() string {
	return "order_history"
}
