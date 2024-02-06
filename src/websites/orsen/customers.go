package orsen

import (
	log "github.com/sirupsen/logrus"
	"math/rand"
	"mosquitoSwarm/src/db/dao"
	"strings"
)

func CreateRandomCustomer(db dao.Database, phonePrefixes string) (name string, phone string) {
	log.Info("Generating a random customer name/phone combination")
	phone = generatePhone(db, phonePrefixes)
	name = generateName(db)
	return
}

func generateName(db dao.Database) string {
	const firstNameOnlyIncidence = 0.2
	const firstNameAfterLastNameIncidence = 0.6
	const nameLowerCaseIncidence = 0.05

	name := db.GetLeastUsedFirstName()
	if !evaluateProbability(firstNameOnlyIncidence) {
		lastName := db.GetLeastUsedLastName()
		if evaluateProbability(firstNameAfterLastNameIncidence) {
			name = lastName + " " + name
		} else {
			name = name + " " + lastName
		}
	}
	if evaluateProbability(nameLowerCaseIncidence) {
		name = strings.ToLower(name)
	}
	return name
}

func generatePhone(db dao.Database, phonePrefixes string) string {
	phone := db.GetLeastUsedPhone()
	prefixes := strings.Split(phonePrefixes, ";")
	prefixIndex := rand.Intn(len(prefixes))
	phone = prefixes[prefixIndex] + phone
	return phone
}

func evaluateProbability(probability float64) bool {
	return rand.Float64() < probability
}
