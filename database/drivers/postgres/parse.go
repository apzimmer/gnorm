package postgres // import "gnorm.org/gnorm/database/drivers/postgres"

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"

	// register postgres driver
	_ "github.com/lib/pq"
	"github.com/pkg/errors"

	"gnorm.org/gnorm/database"
	"gnorm.org/gnorm/database/drivers/postgres/pg"
)

//go:generate go get github.com/xoxo-go/xoxo
//go:generate xoxo pgsql://$DB_USER:$DB_PASSWORD@$DB_HOST/$DB_NAME?sslmode=$DB_SSL_MODE --schema information_schema -o pg --template-path ./templates

// Parse reads the postgres schemas for the given schemas and converts them into
// database.Info structs.
func Parse(log *log.Logger, conn string, schemaNames []string) (*database.Info, error) {
	log.Println("connecting to postgres with DSN", conn)
	db, err := sql.Open("postgres", conn)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	sch := make([]sql.NullString, len(schemaNames))
	for x := range schemaNames {
		sch[x] = sql.NullString{String: schemaNames[x], Valid: true}
	}
	log.Println("querying table schemas for", schemaNames)
	tables, err := pg.QueryTable(db, pg.TableTableSchemaWhere.In(sch), pg.UnOrdered)
	if err != nil {
		return nil, err
	}

	schemas := make(map[string]map[string][]*database.Column, len(schemaNames))
	for _, name := range schemaNames {
		schemas[name] = map[string][]*database.Column{}
	}

	for _, t := range tables {
		s, ok := schemas[t.TableSchema.String]
		if !ok {
			log.Printf("Should be impossible: table %q references unknown schema %q", t.TableName.String, t.TableSchema.String)
			continue
		}
		s[t.TableName.String] = nil
	}

	columns, err := pg.QueryColumn(db, pg.ColumnTableSchemaWhere.In(sch), pg.UnOrdered)
	if err != nil {
		return nil, err
	}

	for _, c := range columns {
		schema, ok := schemas[c.TableSchema.String]
		if !ok {
			log.Printf("Should be impossible: column %q references unknown schema %q", c.ColumnName.String, c.TableSchema.String)
			continue
		}
		_, ok = schema[c.TableName.String]
		if !ok {
			log.Printf("Should be impossible: column %q references unknown table %q in schema %q", c.ColumnName.String, c.TableName.String, c.TableSchema.String)
			continue
		}

		col := toDBColumn(c, log)
		schema[c.TableName.String] = append(schema[c.TableName.String], col)
	}

	enums, err := queryEnums(db, schemaNames)
	if err != nil {
		return nil, err
	}

	res := &database.Info{Schemas: make([]*database.Schema, 0, len(schemas))}
	for _, schema := range schemaNames {
		tables := schemas[schema]
		s := &database.Schema{
			DBName: schema,
			Enums:  enums[schema],
		}
		for tname, columns := range tables {
			s.Tables = append(s.Tables, &database.Table{DBName: tname, DBSchema: schema, Columns: columns})
		}
		res.Schemas = append(res.Schemas, s)
	}

	return res, nil
}

func toDBColumn(c *pg.Column, log *log.Logger) *database.Column {
	col := &database.Column{
		DBName:     c.ColumnName.String,
		Nullable:   bool(c.IsNullable),
		HasDefault: c.ColumnDefault.String != "",
		Orig:       *c,
	}

	typ := c.DataType.String
	switch typ {
	case "ARRAY":
		col.IsArray = true
		// when it's an array, postges prepends an underscore to the standard
		// name.
		typ = c.UdtName.String[1:]

	case "USER-DEFINED":
		col.UserDefined = true
		typ = c.UdtName.String
	}

	length, newtyp, err := calculateLength(typ)
	switch {
	case err != nil:
		log.Println(err)
	case length > 0:
		col.Length = length
		typ = newtyp
	}
	col.DBType = typ

	return col
}

// calculateLength tries to convert a type that contains a length specification
// to a length number and a type name without the brackets.  Thus varchar[32]
// would return 32, "varchar".  It's up to the consumer to understand that
// sometimes length is a maximum and sometimes it's a requirement (i.e.
// varchar[32] vs char[32]), since this information is intrinsic to the type
// name.
func calculateLength(typ string) (length int, newtyp string, err error) {
	idx := strings.Index(typ, "[")
	if idx == -1 {
		// no length indicated
		return 0, "", nil
	}
	end := strings.LastIndex(typ, "]")
	// we expect the length of the type to be the end of the name.
	if end == len(typ)-1 {
		lstr := typ[idx+1 : end]
		l, err := strconv.Atoi(lstr)
		if err != nil {
			return 0, "", err
		}
		return l, typ[:idx], nil
	}
	// something wonky with the brackets
	return 0, "", errors.New("unknown bracket format in type name")
}

func queryEnums(db *sql.DB, schemas []string) (map[string][]*database.Enum, error) {
	const q = `
	SELECT      n.nspname, t.typname as type 
	FROM        pg_type t 
	LEFT JOIN   pg_catalog.pg_namespace n ON n.oid = t.typnamespace 
	WHERE       (t.typrelid = 0 OR (SELECT c.relkind = 'c' FROM pg_catalog.pg_class c WHERE c.oid = t.typrelid)) 
	AND     NOT EXISTS(SELECT 1 FROM pg_catalog.pg_type el WHERE el.oid = t.typelem AND el.typarray = t.oid)
	AND     n.nspname IN (%s)`
	spots := make([]string, len(schemas))
	vals := make([]interface{}, len(schemas))
	for x := range schemas {
		spots[x] = fmt.Sprintf("$%v", x+1)
		vals[x] = schemas[x]
	}
	query := fmt.Sprintf(q, strings.Join(spots, ", "))
	rows, err := db.Query(query, vals...)
	defer rows.Close()
	if err != nil {
		return nil, errors.WithMessage(err, "error querying enum names")
	}
	ret := map[string][]*database.Enum{}
	for rows.Next() {
		var name, schema string
		if err := rows.Scan(&schema, &name); err != nil {
			return nil, errors.WithMessage(err, "error scanning enum name into string")
		}
		vals, err := queryValues(db, schema, name)
		if err != nil {
			return nil, err
		}
		enum := &database.Enum{
			DBName:   name,
			DBSchema: schema,
			Values:   vals,
		}
		ret[schema] = append(ret[schema], enum)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.WithMessage(err, "error reading enum names")
	}
	return ret, nil
}

func queryValues(db *sql.DB, schema, enum string) ([]*database.EnumValue, error) {
	rows, err := db.Query(`
	SELECT
	e.enumlabel,
	e.enumsortorder
	FROM pg_type t
	JOIN ONLY pg_namespace n ON n.oid = t.typnamespace
	LEFT JOIN pg_enum e ON t.oid = e.enumtypid
	WHERE n.nspname = $1 AND t.typname = $2`, schema, enum)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to query enum values for %s.%s", schema, enum)
	}
	defer rows.Close()
	var vals []*database.EnumValue
	for rows.Next() {
		var name string
		var val int
		if err := rows.Scan(&name, &val); err != nil {
			return nil, errors.Wrapf(err, "failed reading enum values for %s.%s", schema, enum)
		}

		vals = append(vals, &database.EnumValue{DBName: name, Value: val})
	}
	return vals, nil
}
