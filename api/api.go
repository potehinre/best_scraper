package api

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/potehinre/best_scraper/availability"
	"github.com/potehinre/best_scraper/config"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var stats requestStatistics

type requestStatistics struct {
	mu    sync.Mutex
	stats map[string]int
}

func (r *requestStatistics) Init(sites []string) {
	r.mu.Lock()
	r.stats = map[string]int{}
	for _, site := range sites {
		r.stats[site] = 0
	}
	r.mu.Unlock()
}

func (r *requestStatistics) Inc(site string) {
	r.mu.Lock()
	if count, ok := r.stats[site]; ok {
		count += 1
		r.stats[site] = count
	}
	r.mu.Unlock()
}

func (r *requestStatistics) GetStats() map[string]int {
	r.mu.Lock()
	res := map[string]int{}
	for k, v := range r.stats {
		res[k] = v
	}
	r.mu.Unlock()
	return res
}

func Init(conf *config.Config) {
	stats.Init(conf.AvailabilityCheck.Sites)
	r := mux.NewRouter()
	r.HandleFunc("/services/slowest", jsonDecorator(slowestServiceHandler)).Methods("GET")
	r.HandleFunc("/services/fastest", jsonDecorator(fastestServiceHandler)).Methods("GET")
	r.HandleFunc("/services/statistics", basicAuthDecorator(jsonDecorator(statistics), conf.HTTP.AuthLogin, conf.HTTP.AuthPassword)).Methods("GET")
	r.HandleFunc("/services/{name}", jsonDecorator(serviceResponseTimeHandler)).Methods("GET")
	http.Handle("/", r)
	srv := &http.Server{
		Handler:      r,
		Addr:         conf.HTTP.Address,
		WriteTimeout: time.Duration(conf.HTTP.WriteTimeoutMilliseconds) * time.Millisecond,
		ReadTimeout:  time.Duration(conf.HTTP.ReadTimeoutMilliseconds) * time.Millisecond,
	}
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.WithFields(log.Fields{"address": conf.HTTP.Address, "error": err}).Fatal("Listening on address")
	}
}

func jsonDecorator(f func(w http.ResponseWriter, r *http.Request) (bool, interface{}, error)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		found, res, err := f(w, r)
		if !found {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resB, err := json.Marshal(res)
		if err != nil {
			log.WithFields(log.Fields{"data": res, "error": err}).Error("can't marshal to json")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, string(resB))
	}
}

func basicAuthDecorator(f func(w http.ResponseWriter, r *http.Request), username, password string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

		user, pass, authOK := r.BasicAuth()
		if authOK == false {
			http.Error(w, "Not authorized", http.StatusUnauthorized)
			return
		}

		if user != username || pass != password {
			http.Error(w, "Not authorized", http.StatusUnauthorized)
			return
		}
		f(w, r)
	}
}

func serviceResponseTimeHandler(w http.ResponseWriter, r *http.Request) (bool, interface{}, error) {
	vars := mux.Vars(r)
	name := vars["name"]
	found, availability := availability.Availability.Get(name)
	if !found {
		return false, nil, nil
	}
	stats.Inc(name)
	return true, availability, nil
}

type fastestSlowestServiceResponse struct {
	Name         string `json:"site_name"`
	ResponseTime int    `json:"response_time"`
}

func slowestServiceHandler(w http.ResponseWriter, r *http.Request) (bool, interface{}, error) {
	found, name, respTime := availability.Availability.MaxResponseTimeSite()
	if !found {
		return false, nil, nil
	}
	return true, &fastestSlowestServiceResponse{name, respTime}, nil
}

func fastestServiceHandler(w http.ResponseWriter, r *http.Request) (bool, interface{}, error) {
	found, name, respTime := availability.Availability.MinResponseTimeSite()
	if !found {
		return false, nil, nil
	}
	return true, &fastestSlowestServiceResponse{name, respTime}, nil
}

func statistics(w http.ResponseWriter, r *http.Request) (bool, interface{}, error) {
	stats := stats.GetStats()
	return true, stats, nil
}
