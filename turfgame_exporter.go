package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sethvargo/go-envconfig"
)

type Config struct {
	TurfApiEndpoint string   `env:"TURF_API_USERS_URL, default=https://api.turfgame.com/unstable/users"`
	TurfUsers       []string `env:"TURF_USERS, required"`
	PollIntervalSec int      `env:"POLL_INTERVAL_SEC, default=300"`
	HttpPort        string   `env:"HTTPD_PORT, default=9097"`
}

type User struct {
	Country          string `json:"country"`
	Medals           []int  `json:"medals"`
	Zones            []int  `json:"zones"`
	PointsPerHour    int    `json:"pointsPerHour"`
	Points           int    `json:"points"`
	Blocktime        int    `json:"blocktime"`
	Taken            int    `json:"taken"`
	Name             string `json:"name"`
	TotalPoints      int    `json:"totalPoints"`
	Rank             int    `json:"rank"`
	Id               int    `json:"id"`
	Place            int    `json:"place"`
	UniqueZonesTaken int    `json:"uniqueZonesTaken"`
	Region           Region `json:"region"`
}

type Region struct {
	Name string `json:"name"`
	Id   int    `json:"id"`
}

// Metrics
var (
	turfgameApiRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "turfgame_api_requests_total",
			Help: "Total number of requests to Turfgame API",
		},
		[]string{"status"},
	)

	zonesOwned = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "turfgame_user_zones_owned",
			Help: "Number of zones owned",
		},
		[]string{"user"},
	)

	pointsPerHour = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "turfgame_user_points_per_hour",
			Help: "Number of points received per hour",
		},
		[]string{"user"},
	)

	roundPoints = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "turfgame_user_points",
			Help: "Number of points received in this round",
		},
		[]string{"user"},
	)

	blocktime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "turfgame_user_blocktime",
			Help: "The users blocktime",
		},
		[]string{"user"},
	)

	takenZones = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "turfgame_user_taken",
			Help: "Number of zones taken",
		},
		[]string{"user"},
	)

	totalPoints = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "turfgame_user_total_points",
			Help: "The users total points",
		},
		[]string{"user"},
	)

	userRank = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "turfgame_user_rank",
			Help: "The users rank",
		},
		[]string{"user"},
	)

	place = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "turfgame_user_place",
			Help: "The users place",
		},
		[]string{"user"},
	)

	uniqueZones = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "turfgame_user_unique_zones_taken",
			Help: "Number of unique zones the user has taken",
		},
		[]string{"user"},
	)

	medalsTaken = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "turfgame_user_medals_taken",
			Help: "Number of medals the user has taken",
		},
		[]string{"user"},
	)

	region = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "turfgame_user_region",
			Help: "The users current region",
		},
		[]string{"user", "region"},
	)
)

func main() {
	ctx := context.Background()
	var c Config

	if err := envconfig.Process(ctx, &c); err != nil {
		log.Fatal(err)
	}

	go backgroundJob(c)

	prometheus.MustRegister(turfgameApiRequestsTotal)
	prometheus.MustRegister(roundPoints)
	prometheus.MustRegister(zonesOwned)
	prometheus.MustRegister(pointsPerHour)
	prometheus.MustRegister(blocktime)
	prometheus.MustRegister(takenZones)
	prometheus.MustRegister(totalPoints)
	prometheus.MustRegister(userRank)
	prometheus.MustRegister(place)
	prometheus.MustRegister(uniqueZones)
	prometheus.MustRegister(medalsTaken)
	prometheus.MustRegister(region)

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":"+c.HttpPort, nil)
}

func backgroundJob(c Config) {
	if len(c.TurfUsers) == 0 {
		log.Fatal("TURF_USERS cannot be an empty string")
	}

	var users []map[string]string
	ch := make(chan []User)

	for _, u := range c.TurfUsers {
		user := map[string]string{
			"name": u,
		}
		users = append(users, user)
	}

	go fetchData(c, users, ch)

	for {
		data := <-ch

		for _, user := range data {
			roundPoints.WithLabelValues(user.Name).Set(float64(user.Points))
			zonesOwned.WithLabelValues(user.Name).Set(float64(len(user.Zones)))
			pointsPerHour.WithLabelValues(user.Name).Set(float64(user.PointsPerHour))
			blocktime.WithLabelValues(user.Name).Set(float64(user.Blocktime))
			takenZones.WithLabelValues(user.Name).Set(float64(user.Taken))
			totalPoints.WithLabelValues(user.Name).Set(float64(user.TotalPoints))
			userRank.WithLabelValues(user.Name).Set(float64(user.Rank))
			place.WithLabelValues(user.Name).Set(float64(user.Place))
			uniqueZones.WithLabelValues(user.Name).Set(float64(user.UniqueZonesTaken))
			medalsTaken.WithLabelValues(user.Name).Set(float64(len(user.Medals)))
			region.WithLabelValues(user.Name, user.Region.Name).Set(1)
		}
	}
}

func fetchData(c Config, users []map[string]string, ch chan []User) <-chan []User {
	json_body, _ := json.Marshal(users)
	var turfData []User

	turfgameApiRequestsTotal.WithLabelValues("ok")
	turfgameApiRequestsTotal.WithLabelValues("error")

	for {
		resp, err := http.Post(c.TurfApiEndpoint, "application/json", bytes.NewBuffer(json_body))

		if err != nil {
			turfgameApiRequestsTotal.WithLabelValues("error").Inc()
			log.Printf("An Error Occured %v", err)
		} else {
			turfgameApiRequestsTotal.WithLabelValues("ok").Inc()
			log.Printf("Sucessfully called %s", c.TurfApiEndpoint)

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				turfgameApiRequestsTotal.WithLabelValues("error").Inc()
				log.Println(err)
			}

			err = json.Unmarshal(body, &turfData)

			if err != nil {
				turfgameApiRequestsTotal.WithLabelValues("error").Inc()
				log.Println(err)
			}

			ch <- turfData
		}

		time.Sleep(time.Duration(c.PollIntervalSec) * time.Second)
	}
}
