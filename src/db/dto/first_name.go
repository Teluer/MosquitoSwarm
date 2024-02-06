package dto

type FirstName struct {
	FirstName string `gorm:"primaryKey;type:varchar(70)"`
	UsedTimes int
}
