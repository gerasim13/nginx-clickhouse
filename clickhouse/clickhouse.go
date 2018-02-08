package clickhouse

import (
	"github.com/gerasim13/nginx-clickhouse/nginx"
	"github.com/mintance/go-clickhouse"
	"github.com/satyrius/gonx"
	"github.com/Sirupsen/logrus"
	"net/url"
	"reflect"
	config "github.com/gerasim13/nginx-clickhouse/config"
)

var clickHouseStorage *clickhouse.Conn

func Save(config *config.Config, logs []gonx.Entry) error {

	storage, err := getStorage(config)

	if err != nil {
		return err
	}

	rows, columns := buildRows(config.ClickHouse.Columns, logs)

	query, err := clickhouse.BuildMultiInsert(
		config.ClickHouse.Db+"."+config.ClickHouse.Table,
		columns,
		rows,
	)

	if err != nil {
		return err
	}

	return query.Exec(storage)
}

func getColumns(columns []config.ColumnDescription) []string {

	keys := reflect.ValueOf(columns).MapKeys()
	stringColumns := make([]string, len(keys))

	for i := 0; i < len(keys); i++ {
		stringColumns[i] = keys[i].String()
	}

	return stringColumns
}

func buildRows(columns []config.ColumnDescription, data []gonx.Entry) (
	rows clickhouse.Rows, cols clickhouse.Columns) {

	for _, logEntry := range data {
		row := clickhouse.Row{}

		for _, column := range columns {
			value, err := logEntry.Field(column.Key)
			if err != nil {
				logrus.Errorf("error to build %s row: %v", column.Key, err)
			}
			row = append(row, nginx.ParseField(column.Type, value))
			cols = append(cols, column.Name)
		}
		rows = append(rows, row)
	}

	return rows, cols
}

func getStorage(config *config.Config) (*clickhouse.Conn, error) {

	if clickHouseStorage != nil {
		return clickHouseStorage, nil
	}

	cHTTP := clickhouse.NewHttpTransport()
	conn := clickhouse.NewConn(config.ClickHouse.Host+":"+config.ClickHouse.Port, cHTTP)

	params := url.Values{}
	params.Add("user", config.ClickHouse.Credentials.User)
	params.Add("password", config.ClickHouse.Credentials.Password)
	conn.SetParams(params)

	if err := conn.Ping(); err != nil {
		return nil, err
	}

	return conn, nil
}
