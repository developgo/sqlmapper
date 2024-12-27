package mysql

import (
	"strings"
	"testing"

	"github.com/mstgnz/sqlporter"
	"github.com/stretchr/testify/assert"
)

func TestMySQL_Parse(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantErr  bool
		validate func(*testing.T, *sqlporter.Schema)
	}{
		{
			name:    "Empty content",
			content: "",
			wantErr: true,
		},
		{
			name: "SET and Configuration Commands",
			content: `
				SET NAMES utf8mb4;
				SET character_set_client = utf8mb4;
				SET time_zone = '+00:00';
				SET foreign_key_checks = 1;`,
			wantErr: false,
			validate: func(t *testing.T, schema *sqlporter.Schema) {
				assert.NotNil(t, schema)
			},
		},
		{
			name: "CREATE DATABASE",
			content: `
				CREATE DATABASE IF NOT EXISTS testdb
				DEFAULT CHARACTER SET utf8mb4
				DEFAULT COLLATE utf8mb4_unicode_ci;`,
			wantErr: false,
			validate: func(t *testing.T, schema *sqlporter.Schema) {
				assert.Equal(t, "testdb", schema.Name)
			},
		},
		{
			name: "CREATE TABLE with All Features",
			content: `
				CREATE TABLE orders (
					order_id INT AUTO_INCREMENT PRIMARY KEY,
					customer_id INT NOT NULL,
					order_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					status ENUM('new', 'processing', 'completed') CHECK (status IN ('new', 'processing', 'completed')),
					total_amount DECIMAL(10,2) DEFAULT 0.00,
					notes TEXT,
					CONSTRAINT fk_customer FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE,
					CONSTRAINT uq_order UNIQUE (order_id, customer_id)
				) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
				
				ALTER TABLE orders COMMENT = 'Store orders table';
				ALTER TABLE orders MODIFY COLUMN status ENUM('new', 'processing', 'completed') COMMENT 'Order current status';`,
			wantErr: false,
			validate: func(t *testing.T, schema *sqlporter.Schema) {
				assert.Len(t, schema.Tables, 1)
				table := schema.Tables[0]
				assert.Equal(t, "orders", table.Name)
				assert.Len(t, table.Columns, 6)
				assert.Len(t, table.Constraints, 4) // PK, FK, CHECK, UNIQUE
			},
		},
		{
			name: "ALTER TABLE Operations",
			content: `
				ALTER TABLE employees ADD COLUMN email VARCHAR(100);
				ALTER TABLE employees MODIFY COLUMN salary DECIMAL(10,2) DEFAULT 5000;
				ALTER TABLE employees MODIFY COLUMN active BOOLEAN NOT NULL;
				ALTER TABLE employees DROP COLUMN notes;
				ALTER TABLE employees ADD CONSTRAINT chk_salary CHECK (salary > 0);
				ALTER TABLE employees RENAME TO staff;`,
			wantErr: false,
			validate: func(t *testing.T, schema *sqlporter.Schema) {
				// ALTER komutları şu an için parse edilmiyor
			},
		},
		{
			name: "DROP TABLE",
			content: `
				DROP TABLE IF EXISTS old_employees;
				DROP TABLE employees CASCADE;`,
			wantErr: false,
			validate: func(t *testing.T, schema *sqlporter.Schema) {
				// DROP komutları şu an için parse edilmiyor
			},
		},
		{
			name: "CREATE INDEX with Variations",
			content: `
				CREATE INDEX idx_employee_name ON employees(last_name, first_name);
				CREATE UNIQUE INDEX idx_employee_email ON employees(email);
				CREATE INDEX idx_employee_salary ON employees(salary DESC);
				CREATE FULLTEXT INDEX idx_employee_bio ON employees(bio);`,
			wantErr: false,
			validate: func(t *testing.T, schema *sqlporter.Schema) {
				// Index'ler tabloya bağlı olduğu için önce tablo oluşturulmalı
			},
		},
		{
			name: "CREATE VIEW",
			content: `
				CREATE OR REPLACE VIEW employee_summary AS
				SELECT 
					department_id,
					COUNT(*) as employee_count,
					AVG(salary) as avg_salary
				FROM employees
				GROUP BY department_id;`,
			wantErr: false,
			validate: func(t *testing.T, schema *sqlporter.Schema) {
				assert.Len(t, schema.Views, 1)
				assert.Equal(t, "employee_summary", schema.Views[0].Name)
			},
		},
		{
			name: "CREATE FUNCTION and PROCEDURE",
			content: `
				DELIMITER //
				CREATE FUNCTION update_employee_status(emp_id INT)
				RETURNS BOOLEAN
				BEGIN
					UPDATE employees SET updated_at = CURRENT_TIMESTAMP WHERE id = emp_id;
					RETURN TRUE;
				END //

				CREATE PROCEDURE process_payroll(IN month INT, IN year INT)
				BEGIN
					-- Process payroll logic
					COMMIT;
				END //
				DELIMITER ;`,
			wantErr: false,
			validate: func(t *testing.T, schema *sqlporter.Schema) {
				assert.Len(t, schema.Functions, 2)
				assert.Equal(t, "update_employee_status", schema.Functions[0].Name)
				assert.Equal(t, "process_payroll", schema.Functions[1].Name)
			},
		},
		{
			name: "CREATE TRIGGER",
			content: `
				DELIMITER //
				CREATE TRIGGER before_employee_update
				BEFORE UPDATE ON employees
				FOR EACH ROW
				BEGIN
					SET NEW.updated_at = CURRENT_TIMESTAMP;
				END //

				CREATE TRIGGER after_salary_change
				AFTER UPDATE ON employees
				FOR EACH ROW
				BEGIN
					IF OLD.salary <> NEW.salary THEN
						INSERT INTO salary_history (employee_id, old_salary, new_salary)
						VALUES (NEW.id, OLD.salary, NEW.salary);
					END IF;
				END //
				DELIMITER ;`,
			wantErr: false,
			validate: func(t *testing.T, schema *sqlporter.Schema) {
				assert.Len(t, schema.Triggers, 2)
				assert.Equal(t, "before_employee_update", schema.Triggers[0].Name)
				assert.Equal(t, "after_salary_change", schema.Triggers[1].Name)
			},
		},
		{
			name: "GRANT and REVOKE",
			content: `
				GRANT SELECT, INSERT, UPDATE ON employees TO 'app_user'@'localhost';
				GRANT ALL PRIVILEGES ON testdb.* TO 'admin'@'%' WITH GRANT OPTION;
				GRANT EXECUTE ON PROCEDURE process_payroll TO 'payroll_user'@'localhost';
				REVOKE UPDATE ON employees FROM 'app_user'@'localhost';`,
			wantErr: false,
			validate: func(t *testing.T, schema *sqlporter.Schema) {
				assert.Len(t, schema.Permissions, 4)
			},
		},
		{
			name: "Foreign Key Constraints",
			content: `
				CREATE TABLE departments (
					id INT AUTO_INCREMENT PRIMARY KEY,
					name VARCHAR(100) NOT NULL
				);

				CREATE TABLE employees (
					id INT AUTO_INCREMENT PRIMARY KEY,
					department_id INT,
					name VARCHAR(100) NOT NULL,
					CONSTRAINT fk_department FOREIGN KEY (department_id)
						REFERENCES departments(id)
						ON DELETE SET NULL
						ON UPDATE CASCADE
				);`,
			wantErr: false,
			validate: func(t *testing.T, schema *sqlporter.Schema) {
				assert.Len(t, schema.Tables, 2)
				assert.Len(t, schema.Tables[1].Constraints, 2) // PK ve FK
			},
		},
		{
			name: "INSERT INTO",
			content: `
				INSERT INTO departments (name) VALUES ('IT'), ('HR'), ('Sales');
				
				INSERT INTO employees (department_id, name) 
				VALUES 
					(1, 'John Doe'),
					(1, 'Jane Smith'),
					(2, 'Bob Wilson');`,
			wantErr: false,
			validate: func(t *testing.T, schema *sqlporter.Schema) {
				// INSERT komutları şu an için parse edilmiyor
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMySQL()
			schema, err := m.Parse(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.validate != nil {
				tt.validate(t, schema)
			}
		})
	}
}

func TestMySQL_Generate(t *testing.T) {
	tests := []struct {
		name    string
		schema  *sqlporter.Schema
		want    string
		wantErr bool
	}{
		{
			name:    "Nil schema",
			schema:  nil,
			wantErr: true,
		},
		{
			name: "Basic schema with one table",
			schema: &sqlporter.Schema{
				Tables: []sqlporter.Table{
					{
						Name: "users",
						Columns: []sqlporter.Column{
							{Name: "id", DataType: "INT", AutoIncrement: true, IsPrimaryKey: true},
							{Name: "name", DataType: "VARCHAR", Length: 100, IsNullable: false},
							{Name: "email", DataType: "VARCHAR", Length: 255, IsNullable: false, IsUnique: true},
						},
					},
				},
			},
			want: strings.TrimSpace(`
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE
);`),
			wantErr: false,
		},
		{
			name: "Schema with table and indexes",
			schema: &sqlporter.Schema{
				Tables: []sqlporter.Table{
					{
						Name: "products",
						Columns: []sqlporter.Column{
							{Name: "id", DataType: "INT", AutoIncrement: true, IsPrimaryKey: true},
							{Name: "name", DataType: "VARCHAR", Length: 100, IsNullable: false},
							{Name: "price", DataType: "DECIMAL", Length: 10, Scale: 2, IsNullable: true},
						},
						Indexes: []sqlporter.Index{
							{Name: "idx_name", Columns: []string{"name"}},
							{Name: "idx_price", Columns: []string{"price"}, IsUnique: true},
						},
					},
				},
			},
			want: strings.TrimSpace(`
CREATE TABLE products (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    price DECIMAL(10,2)
);
CREATE INDEX idx_name ON products(name);
CREATE UNIQUE INDEX idx_price ON products(price);`),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMySQL()
			got, err := m.Generate(tt.schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != "" {
				assert.Equal(t, tt.want, strings.TrimSpace(got))
			}
		})
	}
}
