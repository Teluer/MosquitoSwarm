package dto

type LastName struct {
	LastName  string `gorm:"primaryKey;type:varchar(70)"`
	UsedTimes int
}
