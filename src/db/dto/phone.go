package dto

type Phone struct {
	Phone     string `gorm:"primaryKey;type:varchar(15)"`
	UsedTimes int
}

func (Phone) TableName() string {
	return "phones"
}
