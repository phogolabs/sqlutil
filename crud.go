package sqlutil

import (
	"database/sql"
	"fmt"
	"strings"
)

func QueryRow(db *sql.DB, model interface{}) error {
	t, err := typeOf(model)
	if err != nil {
		return err
	}

	schema, err := metadata.Schema(t)
	if err != nil {
		return err
	}

	v := valueOf(model)
	columns := []string{}
	values := make([]interface{}, 0)

	for _, column := range schema.Columns {
		value := v.Field(column.Index).Addr().Interface()
		expression := fmt.Sprintf("%s = ?", column.Name)

		if column.Constraint&ColumnConstraintPrimaryKey != 0 {
			columns = append(columns, expression)
			values = append(values, value)
		}
	}

	statement := fmt.Sprintf("SELECT * FROM %s WHERE %s", schema.Table, strings.Join(columns, ","))
	row := db.QueryRow(statement, values...)
	return Scan(&RowScanner{row}, model)
}

func Insert(db *sql.DB, model interface{}) (int64, error) {
	t, err := typeOf(model)
	if err != nil {
		return 0, err
	}

	schema, err := metadata.Schema(t)
	if err != nil {
		return 0, err
	}

	v := valueOf(model)
	columns := []string{}
	values := make([]interface{}, 0)
	placeholders := []string{}

	for _, column := range schema.Columns {
		value := v.Field(column.Index).Addr().Interface()
		values = append(values, value)
		columns = append(columns, column.Name)
		placeholders = append(placeholders, "?")
	}

	statement := fmt.Sprintf("INSERT INTO %s (%s) VALUES(%s)", schema.Table, strings.Join(columns, ","), strings.Join(placeholders, ","))
	return execSQL(db, statement, values...)
}

func Update(db *sql.DB, model interface{}) (int64, error) {
	t, err := typeOf(model)
	if err != nil {
		return 0, err
	}

	schema, err := metadata.Schema(t)
	if err != nil {
		return 0, err
	}

	v := valueOf(model)
	columns := []string{}
	values := make([]interface{}, 0)
	conditionValues := make([]interface{}, 0)
	conditions := []string{}

	for _, column := range schema.Columns {
		value := v.Field(column.Index).Addr().Interface()
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
	statement := fmt.Sprintf("UPDATE %s SET %s WHERE %s", schema.Table, strings.Join(columns, ","), strings.Join(conditions, ","))
	return execSQL(db, statement, values...)
}

func Delete(db *sql.DB, model interface{}) (int64, error) {
	t, err := typeOf(model)
	if err != nil {
		return 0, err
	}

	schema, err := metadata.Schema(t)
	if err != nil {
		return 0, err
	}

	v := valueOf(model)
	columns := []string{}
	values := make([]interface{}, 0)

	for _, column := range schema.Columns {
		value := v.Field(column.Index).Addr().Interface()
		expression := fmt.Sprintf("%s = ?", column.Name)

		if column.Constraint&ColumnConstraintPrimaryKey != 0 {
			columns = append(columns, expression)
			values = append(values, value)
		}
	}

	statement := fmt.Sprintf("DELETE FROM %s WHERE %s", schema.Table, strings.Join(columns, ","))
	return execSQL(db, statement, values...)
}

func execSQL(db *sql.DB, statement string, values ...interface{}) (int64, error) {
	result, err := db.Exec(statement, values...)
	if err != nil {
		return 0, err
	}

	cnt, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return cnt, nil
}
