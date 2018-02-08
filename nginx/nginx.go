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
	nginxConfig := strings.NewReader(fmt.Sprintf("%s%s%s%s", `
		http {
			log_format  '`, config.Nginx.LogType, `'  '`, config.Nginx.LogFormat, `';
		}
	`))
	return gonx.NewNginxParser(nginxConfig, config.Nginx.LogType)
}

func ParseField(value_type string, value string) interface{} {
	switch value_type {
		case "time", "Time":
			t, err := time.Parse(config.NginxTimeLayout, value)
			if err == nil {
				return t.Format(config.CHTimeLayout)
			}
			return value

		case "int", "Int":
			val, err := strconv.Atoi(value)
			if err != nil {
				logrus.Error("Error to convert string to int")
			}
			return val

		case "float", "Float":
			val, err := strconv.ParseFloat(value, 32)
			if err != nil {
				logrus.Error("Error to convert string to float32")
			}
			return val

		default:
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
