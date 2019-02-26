# Project Title

Test task for RonteLTD. Task was to write a service that check services availability and response time every minute and provides an api for it.

## Getting Started

just do

```
go build
```

then launch resulting executable.

### Flags

The only flag is -configPath which specifies path to config. Config should be in .toml format.

### Configuration

config file should have following sections:

```
	[HTTP] -- API Http configuration
	Address = "127.0.0.1:8003" -- to which address bind http server
	WriteTimeoutMilliseconds = 1000 -- http server write timeout in milliseconds
	ReadTimeoutMilliseconds = 1000 -- http server read timeout in milliseconds
	AuthLogin = "admin" -- basic auth login for /service/statistics
	AuthPassword = "admin" -- basic auth password for /service/statistics

	[AvailabilityCheck] -- Configuration for availability check service
	CheckPerMinutes = 1 - how usually launch availability check in minutes
	TimeoutMilliseconds = 10000 - http request to service timeout
	Sites = ["xvideos.com", "ya.ru", "rambler.ru"] -- which services to check
``` 

### API

API has following routes:

GET /services/slowest - request slowest service. 404 if none of services available.
GET /services/fastest - request fastest service. 404 if none of services available.
GET /services/statistics - request service request statistics. Requires basic authorization.
GET /services/{name} - request service response_time and availability. 404 if none match.
