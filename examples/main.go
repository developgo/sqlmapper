package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mstgnz/sqlmapper"
	"github.com/mstgnz/sqlmapper/mysql"
	"github.com/mstgnz/sqlmapper/oracle"
	"github.com/mstgnz/sqlmapper/postgres"
	"github.com/mstgnz/sqlmapper/sqlite"
	"github.com/mstgnz/sqlmapper/sqlserver"
)

// DatabaseType represents supported database types
type DatabaseType string

const (
	MySQL      DatabaseType = "mysql"
	PostgreSQL DatabaseType = "postgresql"
	SQLite     DatabaseType = "sqlite"
	Oracle     DatabaseType = "oracle"
	SQLServer  DatabaseType = "sqlserver"
)

// Parser interface for all database parsers
type Parser interface {
	Parse(content string) (*sqlmapper.Schema, error)
	Generate(schema *sqlmapper.Schema) (string, error)
}

// ParserFactory creates parser instances
type ParserFactory struct{}

// NewParser creates a new parser instance based on database type
func (f *ParserFactory) NewParser(dbType DatabaseType) (Parser, error) {
	switch dbType {
	case MySQL:
		return mysql.NewMySQL(), nil
	case PostgreSQL:
		return postgres.NewPostgreSQL(), nil
	case SQLite:
		return sqlite.NewSQLite(), nil
	case Oracle:
		return oracle.NewOracle(), nil
	case SQLServer:
		return sqlserver.NewSQLServer(), nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}

// Converter handles database conversion operations
type Converter struct {
	factory *ParserFactory
}

// NewConverter creates a new converter instance
func NewConverter() *Converter {
	return &Converter{
		factory: &ParserFactory{},
	}
}

// Convert converts SQL from source database type to target database type
func (c *Converter) Convert(sourceType, targetType DatabaseType, content string) (string, error) {
	// Create source parser
	sourceParser, err := c.factory.NewParser(sourceType)
	if err != nil {
		return "", fmt.Errorf("source parser error: %v", err)
	}

	// Parse source content
	schema, err := sourceParser.Parse(content)
	if err != nil {
		return "", fmt.Errorf("parse error: %v", err)
	}

	// Create target parser
	targetParser, err := c.factory.NewParser(targetType)
	if err != nil {
		return "", fmt.Errorf("target parser error: %v", err)
	}

	// Generate target content
	result, err := targetParser.Generate(schema)
	if err != nil {
		return "", fmt.Errorf("generate error: %v", err)
	}

	return result, nil
}

func main() {
	converter := NewConverter()

	// Example conversions
	examples := []struct {
		sourceType DatabaseType
		targetType DatabaseType
		inputFile  string
		outputFile string
	}{
		{PostgreSQL, MySQL, "examples/files/postgres.sql", "examples/files/output/postgres_to_mysql.sql"},
		{MySQL, PostgreSQL, "examples/files/mysql.sql", "examples/files/output/mysql_to_postgres.sql"},
		{Oracle, MySQL, "examples/files/oracle.sql", "examples/files/output/oracle_to_mysql.sql"},
		{SQLServer, PostgreSQL, "examples/files/sqlserver.sql", "examples/files/output/sqlserver_to_postgres.sql"},
		{SQLite, Oracle, "examples/files/sqlite.sql", "examples/files/output/sqlite_to_oracle.sql"},
	}

	for _, example := range examples {
		fmt.Printf("\n=== %s -> %s Conversion ===\n", example.sourceType, example.targetType)

		// Read source file
		content, err := os.ReadFile(example.inputFile)
		if err != nil {
			log.Fatalf("Failed to read source file: %v", err)
		}

		// Convert
		result, err := converter.Convert(example.sourceType, example.targetType, string(content))
		if err != nil {
			log.Fatalf("Conversion error: %v", err)
		}

		// Write result
		err = os.WriteFile(example.outputFile, []byte(result), 0644)
		if err != nil {
			log.Fatalf("Failed to write output file: %v", err)
		}

		fmt.Printf("Conversion completed: %s\n", example.outputFile)
	}
}

func init() {
	// Create output directory
	outputDir := "examples/files/output"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}
}
