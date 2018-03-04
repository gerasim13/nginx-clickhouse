package nginx

import (
	"github.com/gerasim13/nginx-clickhouse/config"
	"github.com/satyrius/gonx"
	"net/url"
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

func EmptyValue(value_type string, value string) string {
	switch value {
		case "undefined", "Undefined", "none", "None", "null", "Null", "NULL":
			value = "-"
		default:
			break
	}
	if value != "-" {
		return value
	}
	switch value_type {
		case "time", "Time", "int", "Int", "float", "Float":
			return "0"
		default:
			break
	}
	return ""
}

func BoolValue(value string) string {
	switch value {
		case "true", "True", "yes", "YES":
			return "1"
		case "false", "False", "no", "NO":
			return "0"
		default:
			break
	}
	return value
}

func ParseField(value_type string, value string) interface{} {
	value = EmptyValue(value_type, value)
	value = BoolValue(value)
	switch value_type {
		case "time", "Time":
			t, err := time.Parse(config.NginxTimeLayout, value)
			if err == nil {
				return t.Format(config.CHTimeLayout)
			}
			break

		case "int", "Int":
			val, err := strconv.Atoi(value)
			if err == nil {
				return val
			}
			logrus.Error(fmt.Sprintf("Error: failed to convert string to int, %s", value))
			break

		case "float", "Float":
			val, err := strconv.ParseFloat(value, 32)
			if err == nil {
				return val
			}
			logrus.Error(fmt.Sprintf("Error: failed to convert string to float32, %s", value))
			break

		default:
			val, err := url.QueryUnescape(value)
			if err == nil {
				return val
			}
			logrus.Error(fmt.Sprintf("Error: failed to decode string, %s", value))
			break
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
