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
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestCreateTable(t *testing.T) {
	// Create an in-memory SQLite database for testing
	db, err := sql.Open("sqlite3", ":memory:")
	assert.NoError(t, err)
	defer db.Close()

	// Initialize DataCenter
	dc := NewDataCenter()

	// Define schema
	schema := SchemaDefine{
		Fields: map[string]string{
			"name": "TEXT",
			"age":  "INTEGER",
		},
	}

	// Generate DDL
	ddl, err := dc.CreateTable("users", schema)
	assert.NoError(t, err)

	t.Log("Generated DDL:", ddl)
	// Execute DDL to create the table
	_, err = db.Exec(ddl)
	assert.NoError(t, err)

	// Check if the table is created
	var tableExists int
	err = db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='users';").Scan(&tableExists)
	assert.NoError(t, err)
	assert.Equal(t, 1, tableExists)
}

func TestCreateTableWithInvalidFieldName(t *testing.T) {
	// Create an in-memory SQLite database for testing
	db, err := sql.Open("sqlite3", ":memory:")
	assert.NoError(t, err)
	defer db.Close()

	// Initialize DataCenter
	dc := NewDataCenter()

	// Define schema with an invalid field name
	schema := SchemaDefine{
		Fields: map[string]string{
			"name":   "TEXT",
			"age":    "INTEGER",
			"select": "TEXT", // "select" is a SQL keyword, invalid field name
		},
	}

	// Attempt to generate DDL
	ddl, err := dc.CreateTable("users", schema)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid field name")
	assert.Empty(t, ddl)
}

func TestCreateTableWithInvalidFieldType(t *testing.T) {
	// Create an in-memory SQLite database for testing
	db, err := sql.Open("sqlite3", ":memory:")
	assert.NoError(t, err)
	defer db.Close()

	// Initialize DataCenter
	dc := NewDataCenter()

	// Define schema with an invalid field type
	schema := SchemaDefine{
		Fields: map[string]string{
			"name": "TEXT",
			"age":  "INVALID_TYPE", // "INVALID_TYPE" is not a valid SQL type
		},
	}

	// Attempt to generate DDL
	ddl, err := dc.CreateTable("users", schema)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid field type")
	assert.Empty(t, ddl)
}
