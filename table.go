package sqlutil

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

type Table struct {
	schema *Schema
	model  interface{}
	t      reflect.Type
	v      reflect.Value
}

func Model(model interface{}) *Table {
	t, err := typeOf(model)
	if err != nil {
		panic(err)
	}

	schema, err := metadata.Schema(t)
	if err != nil {
		panic(err)
	}

	v := valueOf(model)

	return &Table{
		model:  model,
		t:      t,
		v:      v,
		schema: schema,
	}
}

func (t *Table) QueryRow(db *sql.DB) error {
	columns := []string{}
	values := make([]interface{}, 0)

	for _, column := range t.schema.Columns {
		value := t.v.Field(column.Index).Addr().Interface()
		expression := fmt.Sprintf("%s = ?", column.Name)

		if column.Constraint&ColumnConstraintPrimaryKey != 0 {
			columns = append(columns, expression)
			values = append(values, value)
		}
	}

	statement := fmt.Sprintf("SELECT * FROM %s WHERE %s", t.schema.Table, strings.Join(columns, ","))
	row := db.QueryRow(statement, values...)
	return Scan(&RowScanner{row}, t.model)
}

func (t *Table) Insert(db *sql.DB) (int64, error) {
	columns := []string{}
	values := make([]interface{}, 0)
	placeholders := []string{}

	for _, column := range t.schema.Columns {
		value := t.v.Field(column.Index).Addr().Interface()
		values = append(values, value)
		columns = append(columns, column.Name)
		placeholders = append(placeholders, "?")
	}

	statement := fmt.Sprintf("INSERT INTO %s (%s) VALUES(%s)", t.schema.Table, strings.Join(columns, ","), strings.Join(placeholders, ","))
	return execSQL(db, statement, values...)
}

func (t *Table) Update(db *sql.DB) (int64, error) {
	columns := []string{}
	values := make([]interface{}, 0)
	conditionValues := make([]interface{}, 0)
	conditions := []string{}

	for _, column := range t.schema.Columns {
		value := t.v.Field(column.Index).Addr().Interface()
		expression := fmt.Sprintf("%s = ?", column.Name)

		if column.Constraint&ColumnConstraintPrimaryKey != 0 {
			conditions = append(conditions, expression)
			conditionValues = append(values, value)
		} else {
			columns = append(columns, expression)
			values = append(values, value)
		}
	}

	values = append(values, conditionValues...)
	statement := fmt.Sprintf("UPDATE %s SET %s WHERE %s", t.schema.Table, strings.Join(columns, ","), strings.Join(conditions, ","))
	return execSQL(db, statement, values...)
}

func (t *Table) Delete(db *sql.DB) (int64, error) {
	columns := []string{}
	values := make([]interface{}, 0)

	for _, column := range t.schema.Columns {
		value := t.v.Field(column.Index).Addr().Interface()
		expression := fmt.Sprintf("%s = ?", column.Name)

		if column.Constraint&ColumnConstraintPrimaryKey != 0 {
			columns = append(columns, expression)
			values = append(values, value)
		}
	}

	statement := fmt.Sprintf("DELETE FROM %s WHERE %s", t.schema.Table, strings.Join(columns, ","))
	return execSQL(db, statement, values...)
}