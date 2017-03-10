package sqlutil

import (
	"database/sql"
	"fmt"
	"strings"
)

const Separator = ",\n"

func CreateTable(db *sql.DB, model interface{}) error {
	t, err := typeOf(model)
	if err != nil {
		return err
	}

	schema, err := metadata.Schema(t)
	if err != nil {
		return err
	}

	definitions := []string{}
	tableName := strings.ToLower(t.Name())
	tablePK := []string{}

	for _, column := range schema.Columns {
		definition := strings.TrimRight(fmt.Sprintf(" %s %s %s", column.Name, column.DataType, column.Constraint.String()), " ")
		definitions = append(definitions, definition)

		if column.Constraint&ColumnConstraintPrimaryKey != 0 {
			tablePK = append(tablePK, column.Name)
		}
	}

	definitions = append(definitions, fmt.Sprintf(" CONSTRAINT %s_pk PRIMARY KEY(%s)", tableName, strings.Join(tablePK, Separator)))
	statement := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n%s\n)", tableName, strings.Join(definitions, Separator))

	if _, err = db.Exec(statement); err != nil {
		return err
	}

	for _, index := range schema.Indexes {
		statement := fmt.Sprintf("CREATE INDEX %s ON %s (%s)", index.Name, tableName, strings.Join(index.Columns, ","))
		if _, err := db.Exec(statement); err != nil {
			return err
		}
	}

	return nil
}
