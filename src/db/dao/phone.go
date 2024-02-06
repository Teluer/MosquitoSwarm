package dao

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mosquitoSwarm/src/db/dto"
)

func (d *GormDao) SaveNewPhones(phones []dto.Phone) int {
	log.Info("Saving new phones")
	if d.IsTableEmpty(&dto.Phone{}) {
		log.Debug("Phones table is empty, populating")
		d.Insert(&phones)
		return len(phones)
	}
	log.Debug("Phones table has values, adding new")

	existingPhones := d.SelectPhones()
	var newPhones []dto.Phone

NewLoop:
	for _, phone := range phones {
		for _, existingPhone := range existingPhones {
			if existingPhone == phone.Phone {
				continue NewLoop
			}
		}
		newPhones = append(newPhones, phone)
	}

	if len(newPhones) > 0 {
		log.Info(fmt.Sprintf("Inserting %d new phones", len(newPhones)))
		d.Insert(&newPhones)
	} else {
		log.Info("No new phones found")
	}
	return len(newPhones)
}

func (d *GormDao) SelectPhones() []string {
	var result []string
	d.db.Model(&dto.Phone{}).Pluck("phone", &result)
	return result
}

func (d *GormDao) GetLeastUsedPhone() string {
	var phone dto.Phone
	d.db.Order("used_times asc, rand()").First(&phone)
	phone.UsedTimes++
	d.db.Save(&phone)
	return phone.Phone
}
