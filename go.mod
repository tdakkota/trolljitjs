module github.com/tdakkota/trolljitrs

go 1.16

require (
	github.com/cenkalti/backoff/v4 v4.1.1
	github.com/cristalhq/aconfig v0.16.1
	github.com/cristalhq/aconfig/aconfigdotenv v0.16.1
	github.com/cristalhq/aconfig/aconfigyaml v0.16.1
	github.com/gotd/td v0.43.1
	github.com/k0kubun/pp/v3 v3.0.7
	go.uber.org/multierr v1.7.0
	go.uber.org/zap v1.17.0
	golang.org/x/net v0.0.0-20201021035429-f5854403a974
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
)

replace (
	github.com/gotd/ige v0.2.0 => github.com/tdakkota/ige v0.1.4-0.20210609073156-29c48852d442
	github.com/gotd/td v0.43.1 => github.com/tdakkota/td v0.7.2-0.20210612102946-0a72da6d2a30
)
