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

	"github.com/dhose/go-turfgame-exporter/internal/metrics"
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

func main() {
	ctx := context.Background()
	var c Config

	if err := envconfig.Process(ctx, &c); err != nil {
		log.Fatal(err)
	}

	m := metrics.NewMetrics(prometheus.DefaultRegisterer)
	go backgroundJob(c, m)

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":"+c.HttpPort, nil)
}

func backgroundJob(c Config, m *metrics.Metrics) {
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

	go fetchData(c, client, users, ch, m)

	for {
		data := <-ch

		for _, user := range data {
			m.UserPoints.WithLabelValues(user.Name).Set(float64(user.Points))
			m.UserZonesOwned.WithLabelValues(user.Name).Set(float64(len(user.Zones)))
			m.UserPointsPerHour.WithLabelValues(user.Name).Set(float64(user.PointsPerHour))
			m.UserBlocktime.WithLabelValues(user.Name).Set(float64(user.Blocktime))
			m.UserTaken.WithLabelValues(user.Name).Set(float64(user.Taken))
			m.UserTotalPoints.WithLabelValues(user.Name).Set(float64(user.TotalPoints))
			m.UserRank.WithLabelValues(user.Name).Set(float64(user.Rank))
			m.UserPlace.WithLabelValues(user.Name).Set(float64(user.Place))
			m.UserUniqueZonesTaken.WithLabelValues(user.Name).Set(float64(user.UniqueZonesTaken))
			m.UserMedalsTaken.WithLabelValues(user.Name).Set(float64(len(user.Medals)))
			m.UserRegion.WithLabelValues(user.Name, user.Region.Name).Set(1)
		}
	}
}

func fetchData(c Config, client http.Client, users []map[string]string, ch chan []User, m *metrics.Metrics) <-chan []User {
	json_body, _ := json.Marshal(users)
	var turfData []User

	for {
		requestStart := time.Now()
		resp, err := client.Post(c.TurfApiEndpoint, "application/json", bytes.NewBuffer(json_body))
		duration := time.Since(requestStart)
		m.HTTPRequestDuration.WithLabelValues(c.TurfApiEndpoint).Observe(duration.Seconds())

		if err != nil {
			log.Printf("An error occured %v", err)
			m.APIRequestsTotal.WithLabelValues("error").Inc()
		} else {
			log.Printf("The request to %s completed with status code %v and took %v seconds", c.TurfApiEndpoint, resp.StatusCode, duration.Seconds())
			m.APIRequestsTotal.WithLabelValues(strconv.Itoa(resp.StatusCode)).Inc()

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
