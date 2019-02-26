package availability

import (
	"io/ioutil"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/potehinre/best_scraper/config"

	log "github.com/sirupsen/logrus"
)

const schema = "http://"

var Availability SitesAvailability

type SitesAvailability struct {
	sites sync.Map
}

func (s SitesAvailability) Get(site string) (bool, *SiteAvailability) {
	saI, ok := s.sites.Load(site)
	if !ok {
		return false, nil
	}
	sa := saI.(*SiteAvailability)
	return true, sa
}

func (s *SitesAvailability) set(site string, avail *SiteAvailability) {
	s.sites.Store(site, avail)
}

func (s SitesAvailability) MaxResponseTimeSite() (bool, string, int) {
	max := 0
	maxSite := ""
	s.sites.Range(func(key, value interface{}) bool {
		site := key.(string)
		sa := value.(*SiteAvailability)
		if !sa.Available {
			return true
		}
		if sa.ResponseTime > max {
			max = sa.ResponseTime
			maxSite = site
		}
		return true
	})
	if maxSite == "" {
		return false, "", 0
	}
	return true, maxSite, max
}

func (s SitesAvailability) MinResponseTimeSite() (bool, string, int) {
	min := math.MaxInt64
	minSite := ""
	s.sites.Range(func(key, value interface{}) bool {
		site := key.(string)
		sa := value.(*SiteAvailability)
		if !sa.Available {
			return true
		}
		if sa.ResponseTime < min {
			min = sa.ResponseTime
			minSite = site
		}
		return true
	})
	if minSite == "" {
		return false, "", 0
	}
	return true, minSite, min
}

func Check(availConf *config.AvailabilityCheck) {
	client := http.Client{
		Timeout: time.Duration(availConf.TimeoutMilliseconds) * time.Millisecond,
	}
	var wg sync.WaitGroup
	wg.Add(len(availConf.Sites))
	for _, site := range availConf.Sites {
		go checkSiteAvailability(site, &client, &wg)
	}
	wg.Wait()
}

func setSiteUnavailable(site string) {
	Availability.set(site, &SiteAvailability{0, false})
}

type SiteAvailability struct {
	ResponseTime int  `json:"response_time"`
	Available    bool `json:"available"`
}

func checkSiteAvailability(site string, client *http.Client, wg *sync.WaitGroup) {
	defer wg.Done()
	timeBegin := time.Now()
	resp, err := client.Get(schema + site)
	siteLog := log.WithFields(log.Fields{"site": site})
	if err != nil {
		siteLog.WithError(err).Info("error making http Get to site")
		setSiteUnavailable(site)
		return
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	timeEnd := time.Since(timeBegin)
	if err != nil {
		siteLog.WithError(err).Info("error reading response body")
		setSiteUnavailable(site)
		return
	}
	if resp.StatusCode%500 < 100 {
		siteLog.WithFields(log.Fields{"statusCode": resp.StatusCode}).Info("site unavailable")
		setSiteUnavailable(site)
	}
	Availability.set(site, &SiteAvailability{int(timeEnd), true})
}
