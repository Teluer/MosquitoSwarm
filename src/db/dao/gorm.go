package dao

import (
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"mosquitoSwarm/src/db/dto"
)

type Database interface {
	OpenDatabaseConnection(string)
	AutoMigrateAll()
	OpenTransaction() Database
	CommitTransaction()
	RollbackTransaction()
	Insert(obj any)
	Delete(obj any)
	IsTableEmpty(obj any) bool
	ValidateDataIntegrity() bool

	SelectPhones() []string
	GetLeastUsedPhone() string
	SaveNewPhones(phones []dto.Phone) (inserted int)
	GetLeastUsedFirstName() string
	GetLeastUsedLastName() string
}

type GormDao struct {
	db *gorm.DB
}

var Dao Database = &GormDao{}

func (d *GormDao) OpenDatabaseConnection(dsn string) {
	log.Info("Opening connection to DB")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Error("Failed to connect to DB")
		panic(err)
	}
	d.db = db
}

func (d *GormDao) AutoMigrateAll() {
	d.db.AutoMigrate(
		&dto.FirstName{},
		&dto.LastName{},
		&dto.Phone{},
		&dto.OrderHistory{})
}

func (d *GormDao) OpenTransaction() Database {
	return &GormDao{d.db.Begin()}
}

func (d *GormDao) CommitTransaction() {
	d.db.Commit()
}

func (d *GormDao) RollbackTransaction() {
	d.db.Rollback()
}

func (d *GormDao) Insert(obj any) {
	log.WithField("obj", obj).Debug("Inserting data")
	d.db.Create(obj)
}

func (d *GormDao) Delete(obj any) {
	log.WithField("obj", obj).Debug("Deleting data")
	d.db.Where("1 = 1").Delete(obj)
}

func (d *GormDao) IsTableEmpty(obj any) bool {
	return d.db.Limit(1).Find(obj).RowsAffected == 0
}

func (d *GormDao) ValidateDataIntegrity() bool {
	return !(d.IsTableEmpty(&dto.FirstName{}) ||
		d.IsTableEmpty(&dto.LastName{}) ||
		d.IsTableEmpty(&dto.Phone{}))
}
