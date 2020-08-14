package mysql

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

func NewMySQLSource(dsn string) *MySQLSource {
	return &MySQLSource{dsn: dsn}
}

type MySQLSource struct {
	dsn string
}

func (s *MySQLSource) Populate(ctx context.Context) ([]*TableDefinition, error) {
	db, err := sql.Open("mysql", s.dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	tableNames, err := db.QueryContext(ctx, "show tables")
	if err != nil {
		return nil, err
	}
	defer tableNames.Close()
	tables := []*TableDefinition{}
	for tableNames.Next() {
		td := &TableDefinition{Columns: []*ColumnDefinition{}}
		if err := tableNames.Scan(&td.Name); err != nil {
			return nil, err
		}
		cols, err := db.QueryContext(ctx, fmt.Sprintf("show columns from `%s`", td.Name))
		if err != nil {
			return nil, err
		}
		defer cols.Close()
		for cols.Next() {
			cd := &ColumnDefinition{}
			var isNullable string
			err := cols.Scan(&cd.Field, &cd.ColumnType, &isNullable, &cd.Key, &cd.DefaultValue, &cd.Extra)
			if err != nil {
				return nil, err
			}
			if isNullable == "YES" {
				cd.IsNullable = true
			}
			td.Columns = append(td.Columns, cd)
		}
		tables = append(tables, td)
	}
	return tables, nil
}

type ColumnDefinition struct {
	Field        string
	ColumnType   string
	IsNullable   bool
	Key          string
	DefaultValue interface{}
	Extra        string
}

type TableDefinition struct {
	Name    string
	Columns []*ColumnDefinition
}
