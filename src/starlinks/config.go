package starlinks

import (
	"errors"
	"flag"
	"fmt"
	"strings"
)

type Options struct {

	// Server
	Listen  string
	NetType string
	//Domain      string

	// API Server
	APIListen  string
	APINetType string

	// Redis
	RedisDail string

	// SQL
	SQLDail string
	SQLType string

	// Secret
	Secret string
}

func parseNetPath(path string) (string, string, error) {
	var err error
	var domain, address string

	parsed_path := strings.SplitN(path, ":", 2)
	switch len(parsed_path) {
	case 1:
		domain = "tcp"
		address = parsed_path[0]
	case 2:
		if parsed_path[0] == "" {
			domain = "tcp"
		} else {
			domain = parsed_path[0]
		}
		address = parsed_path[1]
	default:
		err = errors.New("Not a valid network path:" + path)
	}

	if domain == "" {
		if err != nil {
			err = fmt.Errorf("Not a valid network domain %v", domain)
		}
		return "", "", err
	}

	return domain, address, nil
}

func parseArgs() (*Options, error) {
	var err error

	listen := flag.String("listen", "0.0.0.0:80", "bind address to accept link requests.")
	//domain := flag.String("domain", "", "host name of http server")

	api_listen := flag.String("api_listen", "127.0.0.1:23278", "Bind address of RESTful config API.")
	//api_domain := flag.String("api_domai")
	redis_dail := flag.String("redis_dail", "127.0.0.1:2379", "Redis cache server.")
	sql_dail := flag.String("sql_dail", "unix:/var/run/mysql.sock", "SQL storage server")
	sql_type := flag.String("sql_type", "mysql", "SQL storage type.")
	secret := flag.String("secret", "starstudio", "Secret string for mapping ID.")
	help := flag.Bool("help", false, "Print help information.")

	flag.Parse()
	if *help {
		flag.PrintDefaults()
		return nil, nil
	}

	opt := new(Options)
	if opt.NetType, opt.Listen, err = parseNetPath(*listen); err != nil {
		if opt.NetType != "tcp" {
			return nil, fmt.Errorf("Link Server: http cannot run over %v", opt.NetType)
		}
		return nil, errors.New("Link Server: " + err.Error())
	}
	if opt.APIListen, opt.APINetType, err = parseNetPath(*api_listen); err != nil {
		if opt.NetType != "tcp" {
			return nil, fmt.Errorf("API Server: http cannot run over %v", opt.APINetType)
		}
		return nil, errors.New("API Server" + err.Error())
	}
	opt.RedisDail = *redis_dail
	opt.SQLDail = *sql_dail
	opt.SQLType = *sql_type
	//opt.Domain = *domain
	opt.Secret = *secret

	return opt, nil
}
