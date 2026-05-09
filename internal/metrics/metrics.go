package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const namespace = "turfgame"

type Metrics struct {
	APIRequestsTotal     *prometheus.CounterVec
	UserZonesOwned       *prometheus.GaugeVec
	UserPointsPerHour    *prometheus.GaugeVec
	UserPoints           *prometheus.GaugeVec
	UserBlocktime        *prometheus.GaugeVec
	UserTaken            *prometheus.GaugeVec
	UserTotalPoints      *prometheus.GaugeVec
	UserRank             *prometheus.GaugeVec
	UserPlace            *prometheus.GaugeVec
	UserUniqueZonesTaken *prometheus.GaugeVec
	UserMedalsTaken      *prometheus.GaugeVec
	UserRegion           *prometheus.GaugeVec
	UserInfo             *prometheus.GaugeVec
	HTTPRequestDuration  *prometheus.HistogramVec
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	return &Metrics{
		APIRequestsTotal: promauto.With(reg).NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "api_requests_total",
			Help:      "Total number of requests to Turfgame API",
		}, []string{"code"}),
		UserZonesOwned: promauto.With(reg).NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "user_zones_owned",
			Help:      "Number of zones owned",
		}, []string{"user"}),
		UserPointsPerHour: promauto.With(reg).NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "user_points_per_hour",
			Help:      "Number of points received per hour",
		}, []string{"user"}),
		UserPoints: promauto.With(reg).NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "user_points",
			Help:      "Number of points received in this round",
		}, []string{"user"}),
		UserBlocktime: promauto.With(reg).NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "user_blocktime",
			Help:      "The users blocktime",
		}, []string{"user"}),
		UserTaken: promauto.With(reg).NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "user_taken",
			Help:      "Number of zones taken",
		}, []string{"user"}),
		UserTotalPoints: promauto.With(reg).NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "user_total_points",
			Help:      "The users total points",
		}, []string{"user"}),
		UserRank: promauto.With(reg).NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "user_rank",
			Help:      "The users rank",
		}, []string{"user"}),
		UserPlace: promauto.With(reg).NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "user_place",
			Help:      "The users place",
		}, []string{"user"}),
		UserUniqueZonesTaken: promauto.With(reg).NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "user_unique_zones_taken",
			Help:      "Number of unique zones the user has taken",
		}, []string{"user"}),
		UserMedalsTaken: promauto.With(reg).NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "user_medals_taken",
			Help:      "Number of medals the user has taken",
		}, []string{"user"}),
		UserRegion: promauto.With(reg).NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "user_region",
			Help:      "The users current region",
		}, []string{"user", "region"}),
		UserInfo: promauto.With(reg).NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "user_info",
			Help:      "Static metadata about the user, always 1",
		}, []string{"user", "user_id", "country", "region", "region_id"}),
		HTTPRequestDuration: promauto.With(reg).NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "A histogram of the HTTP request durations in seconds.",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		}, []string{"url"}),
	}
}
