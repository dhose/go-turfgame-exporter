# go-turfgame-exporter
Prometheus exporter for [turfgame.com](https://turfgame.com) user statistics.

## Environment variables
| Variable name        | Default                           | Required | Description                                                   |
| -------------------- | --------------------------------- | -------- | ------------------------------------------------------------- |
| TURF_USERS           |                                   | Yes      | Comma separated list of Turf usernames                        |
| TURF_API_USERS_URL   | https://api.turfgame.com/v5/users | No       | Turfgame API endpoint                                         |
| POLL_INTERVAL_SEC    | 300                               | No       | Time in seconds between each update of data from turfgame.com |
| HTTPD_PORT           | 9097                              | No       | Network port used to expose metrics                           |

Metrics are served at `http://localhost:<HTTPD_PORT>/metrics` and refreshed every `POLL_INTERVAL_SEC` seconds.

## Exported metrics

| Metric | Type | Labels | Description |
| ------ | ---- | ------ | ----------- |
| `turfgame_user_zones_owned` | Gauge | `user` | Number of zones owned |
| `turfgame_user_points_per_hour` | Gauge | `user` | Number of points received per hour |
| `turfgame_user_points` | Gauge | `user` | Number of points received in this round |
| `turfgame_user_blocktime` | Gauge | `user` | The user's blocktime |
| `turfgame_user_taken` | Gauge | `user` | Number of zones taken |
| `turfgame_user_total_points` | Gauge | `user` | The user's total points |
| `turfgame_user_rank` | Gauge | `user` | The user's rank |
| `turfgame_user_place` | Gauge | `user` | The user's place |
| `turfgame_user_unique_zones_taken` | Gauge | `user` | Number of unique zones the user has taken |
| `turfgame_user_medals_taken` | Gauge | `user` | Number of medals the user has taken |
| `turfgame_user_info` | Gauge | `user`, `user_id`, `country`, `region`, `region_id` | Static user metadata, always 1 |
| `turfgame_user_zone_retake_ratio` | Gauge | `user` | Ratio of total zones taken to unique zones (1.0 = explorer, higher = stationary) |
| `turfgame_last_successful_scrape_timestamp_seconds` | Gauge | — | Unix timestamp of the last successful Turfgame API scrape |
| `turfgame_api_requests_total` | Counter | `code` | Total number of requests to the Turfgame API |
| `http_request_duration_seconds` | Histogram | `url` | HTTP request durations in seconds |
