package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/bigquery"
	"github.com/aereal/bqgen/source/mysql"
	"github.com/aereal/bqgen/sql"
)

func NewApp(out, errOut io.Writer) *App {
	if out == nil {
		out = os.Stdout
	}
	if errOut == nil {
		errOut = os.Stderr
	}
	return &App{out: out, errOut: errOut}
}

type App struct {
	out    io.Writer
	errOut io.Writer
}

func (c *App) Run(argv []string) error {
	fs := flag.NewFlagSet(argv[0], flag.ContinueOnError)
	fs.SetOutput(c.errOut)
	var (
		tableOptions string
		dsn          string
	)
	fs.StringVar(&tableOptions, "table-options", "", "additional table options")
	fs.StringVar(&dsn, "dsn", "", "DSN string")
	err := fs.Parse(argv[1:])
	if err == flag.ErrHelp {
		return nil
	}
	if err != nil {
		return err
	}
	if dsn != "" {
		return c.cmdMySQL(dsn, fs.Args()...)
	}
	return c.cmdMain(tableOptions, fs.Args()...)
}

func (c *App) cmdMain(tableOptions string, args ...string) error {
	b := sql.NewBuilder(&sql.Options{TableOptions: tableOptions})
	for _, schemaFile := range args {
		raw, err := ioutil.ReadFile(schemaFile)
		if err != nil {
			return err
		}
		schema, err := bigquery.SchemaFromJSON(raw)
		if err != nil {
			return err
		}
		b.Consume(schema, buildTableName(schemaFile))
	}
	ddl, err := b.Generate()
	if err != nil {
		return err
	}
	fmt.Fprintln(c.out, ddl)
	return nil
}

func (c *App) cmdMySQL(dsn string, args ...string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	src := mysql.NewMySQLSource(dsn)
	tables, err := src.Populate(ctx)
	if err != nil {
		return err
	}
	for _, t := range tables {
		for _, c := range t.Columns {
			log.Printf("table:%s %#v", t.Name, c)
		}
	}
	return nil
}

func buildTableName(fn string) string {
	base := filepath.Base(fn)
	ext := filepath.Ext(fn)
	return strings.Replace(base, ext, "", 1)
}
