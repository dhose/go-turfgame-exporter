# go-turfgame-exporter
Prometheus exporter for [turfgame.com](https://turfgame.com) user statistics.

## Environment variables
| Variable name        | Default                                 | Description                                                     |
| -------------------- |---------------------------------------- | --------------------------------------------------------------- |
| TURF_USERS           |                                         | Comma separated list of Turf usernames                          |
| TURF_API_USERS_URL   | https://api.turfgame.com/unstable/users | Turfgame API endpoint                                           |
| POLL_INTERVAL_SEC    | 300                                     | Time in seconds between each update of data from turfgame.com   |
| HTTPD_PORT           | 9097                                    | Network port used to expose metrics (defaults to :9097/metrics) |
