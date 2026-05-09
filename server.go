package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/dhose/go-turfgame-exporter/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

type userCounterState struct {
	totalPoints      int
	taken            int
	rank             int
	uniqueZonesTaken int
	medalsTaken      int
	blocktime        int
}

type server struct {
	cfg          Config
	client       *http.Client
	m            *metrics.Metrics
	mux          *http.ServeMux
	prevCounters map[string]userCounterState
}

func newServer(cfg Config) (*server, error) {
	if len(cfg.TurfUsers) == 0 {
		return nil, fmt.Errorf("TURF_USERS cannot be empty")
	}
	reg := prometheus.NewRegistry()
	reg.MustRegister(prometheus.NewGoCollector(), prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	m := metrics.NewMetrics(reg)
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	return &server{
		cfg:          cfg,
		client:       &http.Client{Timeout: 10 * time.Second},
		m:            m,
		mux:          mux,
		prevCounters: make(map[string]userCounterState),
	}, nil
}

func (s *server) run(ctx context.Context) error {
	go s.poll(ctx)
	srv := &http.Server{Addr: ":" + s.cfg.HttpPort, Handler: s.mux}
	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background()) //nolint:errcheck
	}()
	return srv.ListenAndServe()
}

func (s *server) poll(ctx context.Context) {
	jsonBody, _ := json.Marshal(buildUserList(s.cfg.TurfUsers))
	s.fetchAndUpdate(jsonBody)
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(s.cfg.PollIntervalSec) * time.Second):
			s.fetchAndUpdate(jsonBody)
		}
	}
}

func buildUserList(users []string) []map[string]string {
	list := make([]map[string]string, 0, len(users))
	for _, u := range users {
		list = append(list, map[string]string{"name": u})
	}
	return list
}

func (s *server) fetchAndUpdate(jsonBody []byte) {
	start := time.Now()
	resp, err := s.client.Post(s.cfg.TurfApiEndpoint, "application/json", bytes.NewBuffer(jsonBody))
	s.m.HTTPRequestDuration.WithLabelValues(s.cfg.TurfApiEndpoint).Observe(time.Since(start).Seconds())

	if err != nil {
		log.Printf("request error: %v", err)
		s.m.APIRequestsTotal.WithLabelValues("error").Inc()
		return
	}
	defer resp.Body.Close()

	s.m.APIRequestsTotal.WithLabelValues(strconv.Itoa(resp.StatusCode)).Inc()
	log.Printf("request to %s: status=%d duration=%.3fs", s.cfg.TurfApiEndpoint, resp.StatusCode, time.Since(start).Seconds())

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("read error: %v", err)
		return
	}

	var users []User
	if err := json.Unmarshal(body, &users); err != nil {
		log.Printf("unmarshal error: %v", err)
		return
	}

	s.updateMetrics(users)
	s.m.LastSuccessfulScrape.Set(float64(time.Now().Unix()))
}

func addDelta(counter prometheus.Counter, prev, current int) {
	if delta := current - prev; delta > 0 {
		counter.Add(float64(delta))
	}
}

func (s *server) updateMetrics(users []User) {
	for _, u := range users {
		s.m.UserPoints.WithLabelValues(u.Name).Set(float64(u.Points))
		s.m.UserZonesOwned.WithLabelValues(u.Name).Set(float64(len(u.Zones)))
		s.m.UserPointsPerHour.WithLabelValues(u.Name).Set(float64(u.PointsPerHour))
		s.m.UserPlace.WithLabelValues(u.Name).Set(float64(u.Place))
		s.m.UserInfo.WithLabelValues(u.Name, strconv.Itoa(u.Id), u.Country, u.Region.Name, strconv.Itoa(u.Region.Id)).Set(1)
		s.m.UserZoneRetakeRatio.WithLabelValues(u.Name).Set(zoneRetakeRatio(u.Taken, u.UniqueZonesTaken))

		prev := s.prevCounters[u.Name]
		addDelta(s.m.UserPointsTotal.WithLabelValues(u.Name), prev.totalPoints, u.TotalPoints)
		addDelta(s.m.UserTaken.WithLabelValues(u.Name), prev.taken, u.Taken)
		addDelta(s.m.UserRank.WithLabelValues(u.Name), prev.rank, u.Rank)
		addDelta(s.m.UserUniqueZonesTaken.WithLabelValues(u.Name), prev.uniqueZonesTaken, u.UniqueZonesTaken)
		addDelta(s.m.UserMedalsTaken.WithLabelValues(u.Name), prev.medalsTaken, len(u.Medals))
		addDelta(s.m.UserBlocktimeSeconds.WithLabelValues(u.Name), prev.blocktime, u.Blocktime)
		s.prevCounters[u.Name] = userCounterState{
			totalPoints:      u.TotalPoints,
			taken:            u.Taken,
			rank:             u.Rank,
			uniqueZonesTaken: u.UniqueZonesTaken,
			medalsTaken:      len(u.Medals),
			blocktime:        u.Blocktime,
		}
	}
}

func zoneRetakeRatio(taken, uniqueZonesTaken int) float64 {
	if uniqueZonesTaken == 0 {
		return math.NaN()
	}
	return float64(taken) / float64(uniqueZonesTaken)
}
