package dao

import (
	dto2 "mosquitoSwarm/src/db/dto"
)

func (d *GormDao) GetLeastUsedFirstName() string {
	var name dto2.FirstName
	d.db.Order("used_times asc, rand()").First(&name)
	name.UsedTimes++
	d.db.Save(&name)
	return name.FirstName
}

func (d *GormDao) GetLeastUsedLastName() string {
	var name dto2.LastName
	d.db.Order("used_times asc, rand()").First(&name)
	name.UsedTimes++
	d.db.Save(&name)
	return name.LastName
}
