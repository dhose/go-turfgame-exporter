package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sethvargo/go-envconfig"
)

type Config struct {
	TurfApiEndpoint string   `env:"TURF_API_USERS_URL, default=https://api.turfgame.com/v5/users"`
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
		[]string{"code"},
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

	requestDurations = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_request_duration_seconds",
			Help: "A histogram of the HTTP request durations in seconds.",
			// Bucket configuration: the first bucket includes all requests finishing in 0.05 seconds, the last one includes all requests finishing in 10 seconds.
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"url"},
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
	prometheus.MustRegister(requestDurations)

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

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	go fetchData(c, client, users, ch)

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

func fetchData(c Config, client http.Client, users []map[string]string, ch chan []User) <-chan []User {
	json_body, _ := json.Marshal(users)
	var turfData []User

	for {
		requestStart := time.Now()
		resp, err := client.Post(c.TurfApiEndpoint, "application/json", bytes.NewBuffer(json_body))
		duration := time.Since(requestStart)
		requestDurations.WithLabelValues(c.TurfApiEndpoint).Observe(duration.Seconds())

		if err != nil {
			log.Printf("An error occured %v", err)
			turfgameApiRequestsTotal.WithLabelValues("error").Inc()
		} else {
			log.Printf("Sucessfully called %s in %v seconds", c.TurfApiEndpoint, duration.Seconds())
			turfgameApiRequestsTotal.WithLabelValues(strconv.Itoa(resp.StatusCode)).Inc()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Println(err)
			}

			err = json.Unmarshal(body, &turfData)

			if err != nil {
				log.Println(err)
			}

			ch <- turfData
		}

		time.Sleep(time.Duration(c.PollIntervalSec) * time.Second)
	}
}
