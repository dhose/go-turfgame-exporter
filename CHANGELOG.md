# [1.0.0](https://github.com/dhose/go-turfgame-exporter/compare/v0.8.0...v1.0.0) (2026-05-10)


* feat(metrics)!: convert six user gauge metrics to counters ([eb2ef9a](https://github.com/dhose/go-turfgame-exporter/commit/eb2ef9a1c067f14bdc539eac9de5a4e96231ac91))


### BREAKING CHANGES

* all six metrics are renamed due to counter semantics
and the automatic _total suffix added by the Prometheus client library:
  turfgame_user_total_points      -> turfgame_user_points_total
  turfgame_user_blocktime         -> turfgame_user_blocktime_seconds_total
  turfgame_user_taken             -> turfgame_user_taken_total
  turfgame_user_rank              -> turfgame_user_rank_total
  turfgame_user_unique_zones_taken -> turfgame_user_unique_zones_taken_total
  turfgame_user_medals_taken      -> turfgame_user_medals_taken_total

# [0.8.0](https://github.com/dhose/go-turfgame-exporter/compare/v0.7.9...v0.8.0) (2026-05-09)


### Features

* **metrics:** add last_successful_scrape_timestamp_seconds metric ([c8c4dac](https://github.com/dhose/go-turfgame-exporter/commit/c8c4dac29b4fdb0b8accae3390f8c891ff9b7aea))
* **metrics:** add turfgame_user_info metric ([f0d1874](https://github.com/dhose/go-turfgame-exporter/commit/f0d18742a3fb38df2f1c9bf35cf22c9ea91c3c2b))
* **metrics:** add turfgame_user_zone_retake_ratio metric ([6db11f2](https://github.com/dhose/go-turfgame-exporter/commit/6db11f2b49d241bf576b198ab08043312cbe7c17))

## [0.7.9](https://github.com/dhose/go-turfgame-exporter/compare/v0.7.8...v0.7.9) (2026-05-09)


### Bug Fixes

* **deps:** update dependencies ([bba0243](https://github.com/dhose/go-turfgame-exporter/commit/bba0243f7e46b458053cc6b37354cb5bd978d510))
* update go to v1.26.3 ([28da873](https://github.com/dhose/go-turfgame-exporter/commit/28da873d02acdc345524a45dbd3a58117c1f63ad))

## [0.7.8](https://github.com/dhose/go-turfgame-exporter/compare/v0.7.7...v0.7.8) (2026-03-07)


### Bug Fixes

* **deps:** update dependencies ([108cf4b](https://github.com/dhose/go-turfgame-exporter/commit/108cf4bb29e3e48cf0f3051de467e24f610f5f33))
* update go to 1.26.1 ([62a2647](https://github.com/dhose/go-turfgame-exporter/commit/62a2647a97c5bb6568c6ef64977c724f75163b00))

## [0.7.7](https://github.com/dhose/go-turfgame-exporter/compare/v0.7.6...v0.7.7) (2025-09-06)


### Bug Fixes

* bump version to 1.25 ([43b86c6](https://github.com/dhose/go-turfgame-exporter/commit/43b86c6569bc5abaed283bf71e26cdc26e0b3608))

## [0.7.6](https://github.com/dhose/go-turfgame-exporter/compare/v0.7.5...v0.7.6) (2025-09-06)


### Bug Fixes

* **deps:** update module github.com/prometheus/client_golang to v1.23.2 ([6779668](https://github.com/dhose/go-turfgame-exporter/commit/67796689a3837950f8e2db5579fed29a00562f4d))
