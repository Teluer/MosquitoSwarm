package websites

import (
	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
	"mosquitoSwarm/src/db/dao"
	"mosquitoSwarm/src/db/dto"
	"mosquitoSwarm/src/websites/web"
	"sync"
)

// UpdateLastNames populates "last_names" DB table.
// If the table is empty, it fetches last names from the given URL and saves them in DB.
func UpdateLastNames(wg *sync.WaitGroup, lastNamesUrl string) {
	log.Info("Updating last names if needed")
	if dao.Dao.IsTableEmpty(&dto.LastName{}) {
		log.Info("Last names table empty, updating")
		names := fetchLastNames(lastNamesUrl)
		dao.Dao.Insert(&names)
	}
	wg.Done()
}

func fetchLastNames(lastNamesUrl string) (dtos []dto.LastName) {
	appendNames := func(_ int, name *goquery.Selection) {
		dtos = append(dtos, dto.LastName{LastName: name.Text()})
	}
	web.GetUnsafe(lastNamesUrl).Find("ol.row").Find("a").
		Each(appendNames)
	return dtos
}
