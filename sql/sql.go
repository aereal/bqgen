package sql

import (
	"bytes"
	"fmt"
	"io"

	"cloud.google.com/go/bigquery"
)

type Options struct {
	TableOptions string
}

func NewBuilder(opts *Options) *Builder {
	if opts == nil {
		opts = &Options{}
	}
	return &Builder{opts: opts}
}

type Builder struct {
	schemata []*schema
	opts     *Options
}

type schema struct {
	bqSchema bigquery.Schema
	name     string
}

func (b *Builder) Generate() (string, error) {
	out := new(bytes.Buffer)
	for _, s := range b.schemata {
		fmt.Fprintf(out, "CREATE TABLE `%s` (\n", s.name)
		for _, field := range s.bqSchema {
			if err := b.flushColumnDDL(out, field); err != nil {
				return "", err
			}
		}
		fmt.Fprintf(out, ")")
		if b.opts.TableOptions != "" {
			fmt.Fprintf(out, " %s", b.opts.TableOptions)
		}
		fmt.Fprintln(out, ";")
	}
	return out.String(), nil
}

func (b *Builder) Consume(bqSchema bigquery.Schema, tableName string) {
	b.schemata = append(b.schemata, &schema{bqSchema: bqSchema, name: tableName})
}

func (b *Builder) flushColumnDDL(out io.Writer, field *bigquery.FieldSchema) error {
	ct, err := columnType(field)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "  `%s` %s,\n", field.Name, ct)
	if err != nil {
		return err
	}
	return nil
}

func columnType(field *bigquery.FieldSchema) (string, error) {
	if field.Repeated {
		return "", fmt.Errorf("currently repeated field is not supported")
	}
	var sqlType string
	switch field.Type {
	case bigquery.StringFieldType:
		sqlType = "TEXT"
	case bigquery.BytesFieldType:
		sqlType = "BLOB"
	case bigquery.IntegerFieldType:
		sqlType = "BIGINT"
	case bigquery.FloatFieldType:
		sqlType = "DECIMAL" // TODO
	case bigquery.BooleanFieldType:
		sqlType = "BOOLEAN"
	case bigquery.TimestampFieldType:
		sqlType = "TIMESTAMP"
	case bigquery.RecordFieldType:
		// TODO
	case bigquery.DateFieldType:
		sqlType = "DATE"
	case bigquery.TimeFieldType:
		sqlType = "TIME"
	case bigquery.DateTimeFieldType:
		sqlType = "DATETIME"
	case bigquery.NumericFieldType:
		// TODO
	case bigquery.GeographyFieldType:
		// TODO
	default:
		return "", fmt.Errorf("Unknown type: %s", field.Type)
	}
	if field.Required {
		sqlType += " NOT NULL"
	}
	return sqlType, nil
}
