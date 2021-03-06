package avatica

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type authentication int

const (
	none authentication = iota
	basic
	digest
)

// Config is a configuration parsed from a DSN string
type Config struct {
	endpoint             string
	maxRowsTotal         int64
	frameMaxSize         int32
	location             *time.Location
	schema               string
	transactionIsolation uint32

	user     string
	password string

	authentication  authentication
	avaticaUser     string
	avaticaPassword string
}

// ParseDSN parses a DSN string to a Config
func ParseDSN(dsn string) (*Config, error) {

	conf := &Config{
		maxRowsTotal:         -1,
		frameMaxSize:         -1,
		location:             time.UTC,
		transactionIsolation: 0,
	}

	parsed, err := url.ParseRequestURI(dsn)

	if err != nil {
		return nil, fmt.Errorf("Unable to parse DSN: %s", err)
	}

	userInfo := parsed.User

	if userInfo != nil {
		if userInfo.Username() != "" {
			conf.user = userInfo.Username()
		}

		if pass, ok := userInfo.Password(); ok {
			conf.password = pass
		}
	}

	queries := parsed.Query()

	if v := queries.Get("maxRowsTotal"); v != "" {

		maxRowTotal, err := strconv.Atoi(v)

		if err != nil {
			return nil, fmt.Errorf("Invalid value for maxRowsTotal: %s", err)
		}

		conf.maxRowsTotal = int64(maxRowTotal)
	}

	if v := queries.Get("frameMaxSize"); v != "" {

		maxRowTotal, err := strconv.Atoi(v)

		if err != nil {
			return nil, fmt.Errorf("Invalid value for frameMaxSize: %s", err)
		}

		conf.frameMaxSize = int32(maxRowTotal)
	}

	if v := queries.Get("location"); v != "" {

		loc, err := time.LoadLocation(v)

		if err != nil {
			return nil, fmt.Errorf("Invalid value for location: %s", err)
		}

		conf.location = loc
	}

	if v := queries.Get("transactionIsolation"); v != "" {

		isolation, err := strconv.Atoi(v)

		if err != nil {
			return nil, fmt.Errorf("Invalid value for transactionIsolation: %s", err)
		}

		if isolation < 0 || isolation > 8 || isolation&(isolation-1) != 0 {
			return nil, fmt.Errorf("transactionIsolation must be 0, 1, 2, 4 or 8, %d given", isolation)
		}

		conf.transactionIsolation = uint32(isolation)
	}

	if v := queries.Get("authentication"); v != "" {

		auth := strings.ToUpper(v)

		if auth == "BASIC" {
			conf.authentication = basic
		} else if auth == "DIGEST" {
			conf.authentication = digest
		} else {
			return nil, fmt.Errorf("authentication must be either BASIC or DIGEST")
		}

		user := queries.Get("avaticaUser")

		if user == "" {
			return nil, fmt.Errorf("authentication is set to %s, but avaticaUser is empty", v)
		}

		conf.avaticaUser = user

		pass := queries.Get("avaticaPassword")

		if pass == "" {
			return nil, fmt.Errorf("authentication is set to %s, but avaticaPassword is empty", v)
		}

		conf.avaticaPassword = pass
	}

	if parsed.Path != "" {
		conf.schema = strings.TrimPrefix(parsed.Path, "/")
	}

	parsed.User = nil
	parsed.RawQuery = ""
	parsed.Fragment = ""

	conf.endpoint = parsed.String()

	return conf, nil
}
