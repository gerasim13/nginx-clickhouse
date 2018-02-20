package nginx

import (
	"github.com/gerasim13/nginx-clickhouse/config"
	"github.com/satyrius/gonx"
	"io"
	"strconv"
	"strings"
	"time"
	"github.com/Sirupsen/logrus"
	"fmt"
)

func GetParser(config *config.Config) (*gonx.Parser, error) {
	// Use nginx config file to extract format by the name
	confString := fmt.Sprintf("%s%s%s%s%s",`
	http {
		log_format  `, config.Nginx.LogType, `  '`, config.Nginx.LogFormat, `';
	}`)
	nginxConfig := strings.NewReader(confString)
	return gonx.NewNginxParser(nginxConfig, config.Nginx.LogType)
}

func ParseField(value_type string, value string) interface{} {
	switch value_type {
		case "time", "Time":
			if value == "-" {
			    value = "0"
			}
			t, err := time.Parse(config.NginxTimeLayout, value)
			if err == nil {
				return t.Format(config.CHTimeLayout)
			}
			return value

		case "int", "Int":
			if value == "-" {
			    value = "0"
			}
			val, err := strconv.Atoi(value)
			if err != nil {
				logrus.Error(fmt.Sprintf("Error to convert string to int, %s", value))
			}
			return val

		case "float", "Float":
			if value == "-" {
			    value = "0"
			}
			val, err := strconv.ParseFloat(value, 32)
			if err != nil {
				logrus.Error(fmt.Sprintf("Error to convert string to float32, %s", value))
			}
			return val

		default:
			if value == "-" {
			    value = ""
			}
			return value
	}
	return value
}

func ParseLogs(parser *gonx.Parser, logLines []string) []gonx.Entry {
	logReader := strings.NewReader(strings.Join(logLines, "\n"))
	reader := gonx.NewParserReader(logReader, parser)

	var logs []gonx.Entry

	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		// Process the record... e.g.
		logs = append(logs, *rec)
	}

	return logs
}
