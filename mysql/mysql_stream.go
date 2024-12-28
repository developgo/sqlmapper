package mysql

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/mstgnz/sqlmapper"
	"github.com/mstgnz/sqlmapper/stream"
)

// MySQLStreamParser implements the StreamParser interface for MySQL
type MySQLStreamParser struct {
	mysql *MySQL
}

// NewMySQLStreamParser creates a new MySQL stream parser
func NewMySQLStreamParser() *MySQLStreamParser {
	return &MySQLStreamParser{
		mysql: NewMySQL().(*MySQL),
	}
}

// ParseStream implements the StreamParser interface
func (p *MySQLStreamParser) ParseStream(reader io.Reader, callback func(stream.SchemaObject) error) error {
	streamReader := stream.NewStreamReader(reader, ";")

	for {
		statement, err := streamReader.ReadStatement()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading statement: %v", err)
		}

		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}

		// Parse CREATE TABLE statements
		if strings.HasPrefix(strings.ToUpper(statement), "CREATE TABLE") {
			table, err := p.parseTableStatement(statement)
			if err != nil {
				return err
			}

			err = callback(stream.SchemaObject{
				Type: stream.TableObject,
				Data: table,
			})
			if err != nil {
				return err
			}
			continue
		}

		// Parse CREATE VIEW statements
		if strings.HasPrefix(strings.ToUpper(statement), "CREATE VIEW") {
			view, err := p.parseViewStatement(statement)
			if err != nil {
				return err
			}

			err = callback(stream.SchemaObject{
				Type: stream.ViewObject,
				Data: view,
			})
			if err != nil {
				return err
			}
			continue
		}

		// Parse CREATE FUNCTION statements
		if strings.HasPrefix(strings.ToUpper(statement), "CREATE FUNCTION") {
			function, err := p.parseFunctionStatement(statement)
			if err != nil {
				return err
			}

			err = callback(stream.SchemaObject{
				Type: stream.FunctionObject,
				Data: function,
			})
			if err != nil {
				return err
			}
			continue
		}

		// Parse CREATE PROCEDURE statements
		if strings.HasPrefix(strings.ToUpper(statement), "CREATE PROCEDURE") {
			procedure, err := p.parseProcedureStatement(statement)
			if err != nil {
				return err
			}

			err = callback(stream.SchemaObject{
				Type: stream.ProcedureObject,
				Data: procedure,
			})
			if err != nil {
				return err
			}
			continue
		}

		// Parse CREATE TRIGGER statements
		if strings.HasPrefix(strings.ToUpper(statement), "CREATE TRIGGER") {
			trigger, err := p.parseTriggerStatement(statement)
			if err != nil {
				return err
			}

			err = callback(stream.SchemaObject{
				Type: stream.TriggerObject,
				Data: trigger,
			})
			if err != nil {
				return err
			}
			continue
		}
	}

	return nil
}

// ParseStreamParallel implements parallel processing for MySQL stream parsing
func (p *MySQLStreamParser) ParseStreamParallel(reader io.Reader, callback func(stream.SchemaObject) error, workers int) error {
	streamReader := stream.NewStreamReader(reader, ";")
	statements := make(chan string, workers)
	results := make(chan stream.SchemaObject, workers)
	errors := make(chan error, workers)
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for statement := range statements {
				obj, err := p.parseStatement(statement)
				if err != nil {
					errors <- err
					return
				}
				if obj != nil {
					results <- *obj
				}
			}
		}()
	}

	// Start a goroutine to close results channel after all workers are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Start a goroutine to read statements and send them to workers
	go func() {
		for {
			statement, err := streamReader.ReadStatement()
			if err == io.EOF {
				break
			}
			if err != nil {
				errors <- fmt.Errorf("error reading statement: %v", err)
				break
			}

			statement = strings.TrimSpace(statement)
			if statement == "" {
				continue
			}
			statements <- statement
		}
		close(statements)
	}()

	// Process results and handle errors
	for obj := range results {
		if err := callback(obj); err != nil {
			return err
		}
	}

	// Check for any errors from workers
	select {
	case err := <-errors:
		return err
	default:
		return nil
	}
}

// parseStatement parses a single SQL statement and returns a SchemaObject
func (p *MySQLStreamParser) parseStatement(statement string) (*stream.SchemaObject, error) {
	upperStatement := strings.ToUpper(statement)

	switch {
	case strings.HasPrefix(upperStatement, "CREATE TABLE"):
		table, err := p.parseTableStatement(statement)
		if err != nil {
			return nil, err
		}
		return &stream.SchemaObject{
			Type: stream.TableObject,
			Data: table,
		}, nil

	case strings.HasPrefix(upperStatement, "CREATE VIEW"):
		view, err := p.parseViewStatement(statement)
		if err != nil {
			return nil, err
		}
		return &stream.SchemaObject{
			Type: stream.ViewObject,
			Data: view,
		}, nil

	case strings.HasPrefix(upperStatement, "CREATE FUNCTION"):
		function, err := p.parseFunctionStatement(statement)
		if err != nil {
			return nil, err
		}
		return &stream.SchemaObject{
			Type: stream.FunctionObject,
			Data: function,
		}, nil

	case strings.HasPrefix(upperStatement, "CREATE PROCEDURE"):
		procedure, err := p.parseProcedureStatement(statement)
		if err != nil {
			return nil, err
		}
		return &stream.SchemaObject{
			Type: stream.ProcedureObject,
			Data: procedure,
		}, nil
	}

	return nil, nil
}

// GenerateStream implements the StreamParser interface
func (p *MySQLStreamParser) GenerateStream(schema *sqlmapper.Schema, writer io.Writer) error {
	if schema == nil {
		return fmt.Errorf("schema cannot be nil")
	}

	// Write tables
	for _, table := range schema.Tables {
		stmt := p.mysql.generateTableSQL(table)
		if _, err := writer.Write([]byte(stmt + ";\n\n")); err != nil {
			return err
		}

		// Generate indexes for this table
		for _, index := range table.Indexes {
			stmt := p.mysql.generateIndexSQL(table.Name, index)
			if _, err := writer.Write([]byte(stmt + ";\n")); err != nil {
				return err
			}
		}
	}

	// Write views
	for _, view := range schema.Views {
		stmt := fmt.Sprintf("CREATE VIEW %s AS %s", view.Name, view.Definition)
		if _, err := writer.Write([]byte(stmt + ";\n\n")); err != nil {
			return err
		}
	}

	// Write functions
	for _, function := range schema.Functions {
		if !function.IsProc {
			stmt := fmt.Sprintf("CREATE FUNCTION %s(", function.Name)
			for i, param := range function.Parameters {
				if i > 0 {
					stmt += ", "
				}
				stmt += fmt.Sprintf("%s %s", param.Name, param.DataType)
			}
			stmt += fmt.Sprintf(") RETURNS %s\n%s", function.Returns, function.Body)
			if _, err := writer.Write([]byte(stmt + ";\n\n")); err != nil {
				return err
			}
		}
	}

	// Write procedures
	for _, function := range schema.Functions {
		if function.IsProc {
			stmt := fmt.Sprintf("CREATE PROCEDURE %s(", function.Name)
			for i, param := range function.Parameters {
				if i > 0 {
					stmt += ", "
				}
				stmt += fmt.Sprintf("%s %s", param.Name, param.DataType)
			}
			stmt += fmt.Sprintf(")\n%s", function.Body)
			if _, err := writer.Write([]byte(stmt + ";\n\n")); err != nil {
				return err
			}
		}
	}

	// Write triggers
	for _, trigger := range schema.Triggers {
		stmt := fmt.Sprintf("CREATE TRIGGER %s %s %s ON %s\n%s",
			trigger.Name, trigger.Timing, trigger.Event, trigger.Table, trigger.Body)
		if _, err := writer.Write([]byte(stmt + ";\n\n")); err != nil {
			return err
		}
	}

	return nil
}

// parseTableStatement parses a CREATE TABLE statement
func (p *MySQLStreamParser) parseTableStatement(statement string) (*sqlmapper.Table, error) {
	// Create a temporary schema for parsing
	tempSchema := &sqlmapper.Schema{}
	p.mysql.schema = tempSchema

	// Parse the table using the existing MySQL parser
	if err := p.mysql.parseTables(statement); err != nil {
		return nil, err
	}

	// Check if any table was parsed
	if len(tempSchema.Tables) == 0 {
		return nil, fmt.Errorf("no table found in statement")
	}

	// Return the first table
	return &tempSchema.Tables[0], nil
}

// parseViewStatement parses a CREATE VIEW statement
func (p *MySQLStreamParser) parseViewStatement(statement string) (*sqlmapper.View, error) {
	// Create a temporary schema for parsing
	tempSchema := &sqlmapper.Schema{}
	p.mysql.schema = tempSchema

	// Parse the view using the existing MySQL parser
	if err := p.mysql.parseViews(statement); err != nil {
		return nil, err
	}

	// Check if any view was parsed
	if len(tempSchema.Views) == 0 {
		return nil, fmt.Errorf("no view found in statement")
	}

	// Return the first view
	return &tempSchema.Views[0], nil
}

// parseFunctionStatement parses a CREATE FUNCTION statement
func (p *MySQLStreamParser) parseFunctionStatement(statement string) (*sqlmapper.Function, error) {
	// Create a temporary schema for parsing
	tempSchema := &sqlmapper.Schema{}
	p.mysql.schema = tempSchema

	// Parse the function using the existing MySQL parser
	if err := p.mysql.parseFunctions(statement); err != nil {
		return nil, err
	}

	// Check if any function was parsed
	if len(tempSchema.Functions) == 0 {
		return nil, fmt.Errorf("no function found in statement")
	}

	// Return the first function
	return &tempSchema.Functions[0], nil
}

// parseProcedureStatement parses a CREATE PROCEDURE statement
func (p *MySQLStreamParser) parseProcedureStatement(statement string) (*sqlmapper.Procedure, error) {
	// Create a temporary schema for parsing
	tempSchema := &sqlmapper.Schema{}
	p.mysql.schema = tempSchema

	// Parse the procedure using the existing MySQL parser
	if err := p.mysql.parseFunctions(statement); err != nil {
		return nil, err
	}

	// Check if any function was parsed
	if len(tempSchema.Functions) == 0 {
		return nil, fmt.Errorf("no procedure found in statement")
	}

	// Find the first procedure (function with IsProc=true)
	for _, fn := range tempSchema.Functions {
		if fn.IsProc {
			proc := &sqlmapper.Procedure{
				Name:       fn.Name,
				Parameters: fn.Parameters,
				Body:       fn.Body,
				Schema:     fn.Schema,
			}
			return proc, nil
		}
	}

	return nil, fmt.Errorf("no procedure found in statement")
}

// parseTriggerStatement parses a CREATE TRIGGER statement
func (p *MySQLStreamParser) parseTriggerStatement(statement string) (*sqlmapper.Trigger, error) {
	// Create a temporary schema for parsing
	tempSchema := &sqlmapper.Schema{}
	p.mysql.schema = tempSchema

	// Parse the trigger using the existing MySQL parser
	if err := p.mysql.parseTriggers(statement); err != nil {
		return nil, err
	}

	// Check if any trigger was parsed
	if len(tempSchema.Triggers) == 0 {
		return nil, fmt.Errorf("no trigger found in statement")
	}

	// Return the first trigger
	return &tempSchema.Triggers[0], nil
}
