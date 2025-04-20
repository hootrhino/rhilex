// Copyright (C) 2025 wwhai
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package xmanager

import (
	"fmt"
	"slices"
	"strings"
)

// SchemaDefine defines the structure of the table fields.
type SchemaDefine struct {
	Fields map[string]string
}

// DataCenter represents the data center object with a connection to the database.
type DataCenter struct {
	// You can add DB connection or other necessary fields here
}

// NewDataCenter creates a new DataCenter instance.
func NewDataCenter() *DataCenter {
	return &DataCenter{}
}

// CreateTable generates a DDL statement to create a table based on the schema definition.
func (dc *DataCenter) CreateTable(tableName string, schema SchemaDefine) (string, error) {
	// Validate the schema
	if err := validateSchema(schema); err != nil {
		return "", err
	}

	// Ensure the first field is always 'ts' (timestamp)
	if _, exists := schema.Fields["ts"]; !exists {
		// Add the ts field with proper type if it doesn't exist
		schema.Fields["ts"] = "INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))"
	} else {
		// Force 'ts' field to have the correct timestamp type (INTEGER)
		if schema.Fields["ts"] != "INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))" {
			schema.Fields["ts"] = "INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))"
		}
	}

	// Generate DDL statement
	var ddlParts []string
	ddlParts = append(ddlParts, fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (", sanitizeFieldName(tableName)))

	// Add fields from schema
	var i int
	for fieldName, fieldType := range schema.Fields {
		// Sanitize field name to avoid SQL injection and keyword conflicts
		fieldName = sanitizeFieldName(fieldName)
		ddlParts = append(ddlParts, fmt.Sprintf("%s %s", fieldName, sanitizeFieldType(fieldType)))

		// Add a comma after each field except the last one
		i++
		if i < len(schema.Fields) {
			ddlParts = append(ddlParts, ", ")
		}
	}

	ddlParts = append(ddlParts, ");")
	ddl := strings.Join(ddlParts, "")

	return ddl, nil
}

// validateSchema validates the schema to ensure no invalid field names or types.
func validateSchema(schema SchemaDefine) error {
	for fieldName, fieldType := range schema.Fields {
		if !isValidFieldName(fieldName) {
			return fmt.Errorf("invalid field name: %s", fieldName)
		}
		if !isValidFieldType(fieldType) {
			return fmt.Errorf("invalid field type: %s", fieldType)
		}
	}
	return nil
}

// isValidFieldName checks if the field name is valid (not a SQL reserved keyword).
func isValidFieldName(fieldName string) bool {
	// Example: a list of SQL reserved keywords could be used here.
	// For simplicity, let's assume "SELECT" and "TABLE" are reserved.
	reservedKeywords := []string{"SELECT", "TABLE", "INSERT", "DELETE", "UPDATE"}
	for _, keyword := range reservedKeywords {
		if strings.ToUpper(fieldName) == keyword {
			return false
		}
	}
	return true
}

// isValidFieldType checks if the field type is a valid SQL type.
func isValidFieldType(fieldType string) bool {
	// You can expand this check for more field types if needed.
	validTypes := []string{"TEXT", "INTEGER", "REAL", "BLOB", "NUMERIC"}
	return slices.Contains(validTypes, strings.ToUpper(fieldType))
}

// sanitizeFieldName ensures that the field name doesn't clash with SQL keywords or injection.
func sanitizeFieldName(name string) string {
	// Wrapping field name with backticks to avoid keyword conflicts.
	return fmt.Sprintf("`%s`", name)
}

// sanitizeFieldType ensures that the field type is safe to use in a SQL statement.
func sanitizeFieldType(fieldType string) string {
	// Simply returning the field type as is, but this can be extended.
	return fieldType
}
