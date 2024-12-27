package oracle

import (
	"testing"

	"github.com/mstgnz/sqlporter"
	"github.com/stretchr/testify/assert"
)

func TestOracle_Parse(t *testing.T) {
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
			name: "CREATE TABLE with All Features",
			content: `
				CREATE TABLE users (
					id NUMBER DEFAULT users_seq.NEXTVAL PRIMARY KEY,
					username VARCHAR2(50) NOT NULL UNIQUE,
					email VARCHAR2(100) NOT NULL,
					password VARCHAR2(255) NOT NULL,
					status VARCHAR2(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
					created_at TIMESTAMP DEFAULT SYSTIMESTAMP,
					updated_at TIMESTAMP DEFAULT SYSTIMESTAMP
				);
				
				CREATE TABLE posts (
					id NUMBER DEFAULT posts_seq.NEXTVAL PRIMARY KEY,
					user_id NUMBER NOT NULL,
					title VARCHAR2(255) NOT NULL,
					content CLOB,
					status VARCHAR2(20) DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'archived')),
					created_at TIMESTAMP DEFAULT SYSTIMESTAMP,
					updated_at TIMESTAMP DEFAULT SYSTIMESTAMP,
					CONSTRAINT fk_posts_users FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
				);`,
			wantErr: false,
			validate: func(t *testing.T, schema *sqlporter.Schema) {
				assert.Len(t, schema.Tables, 2)

				// Users tablosu kontrolü
				usersTable := schema.Tables[0]
				assert.Equal(t, "users", usersTable.Name)
				assert.Len(t, usersTable.Columns, 7)

				// Posts tablosu kontrolü
				postsTable := schema.Tables[1]
				assert.Equal(t, "posts", postsTable.Name)
				assert.Len(t, postsTable.Columns, 7)

				// Foreign key kontrolü
				fkFound := false
				for _, constraint := range postsTable.Constraints {
					if constraint.Type == "FOREIGN KEY" {
						fkFound = true
						assert.Equal(t, "fk_posts_users", constraint.Name)
						assert.Equal(t, []string{"user_id"}, constraint.Columns)
						assert.Equal(t, "users", constraint.RefTable)
						assert.Equal(t, []string{"id"}, constraint.RefColumns)
						assert.Equal(t, "CASCADE", constraint.DeleteRule)
					}
				}
				assert.True(t, fkFound, "Foreign key constraint not found")
			},
		},
		{
			name: "CREATE SEQUENCE",
			content: `
				CREATE SEQUENCE users_seq START WITH 1 INCREMENT BY 1;
				CREATE SEQUENCE posts_seq START WITH 1 INCREMENT BY 1;`,
			wantErr: false,
			validate: func(t *testing.T, schema *sqlporter.Schema) {
				assert.Len(t, schema.Sequences, 2)
				assert.Equal(t, "users_seq", schema.Sequences[0].Name)
				assert.Equal(t, "posts_seq", schema.Sequences[1].Name)
			},
		},
		{
			name: "CREATE VIEW",
			content: `
				CREATE OR REPLACE VIEW active_users_view AS
				SELECT 
					u.*,
					COUNT(p.id) as post_count,
					MAX(p.created_at) as last_post_date
				FROM users u
				LEFT JOIN posts p ON u.id = p.user_id
				WHERE u.status = 'active'
				GROUP BY u.id, u.username, u.email, u.password, u.status, u.created_at, u.updated_at;`,
			wantErr: false,
			validate: func(t *testing.T, schema *sqlporter.Schema) {
				assert.Len(t, schema.Views, 1)
				assert.Equal(t, "active_users_view", schema.Views[0].Name)
			},
		},
		{
			name: "CREATE TRIGGER",
			content: `
				CREATE OR REPLACE TRIGGER users_update_timestamp
				BEFORE UPDATE ON users
				FOR EACH ROW
				BEGIN
					:NEW.updated_at := SYSTIMESTAMP;
				END;
				/
				
				CREATE OR REPLACE TRIGGER posts_update_timestamp
				BEFORE UPDATE ON posts
				FOR EACH ROW
				BEGIN
					:NEW.updated_at := SYSTIMESTAMP;
				END;
				/`,
			wantErr: false,
			validate: func(t *testing.T, schema *sqlporter.Schema) {
				assert.Len(t, schema.Triggers, 2)

				// Users trigger kontrolü
				usersTrigger := schema.Triggers[0]
				assert.Equal(t, "users_update_timestamp", usersTrigger.Name)
				assert.Equal(t, "users", usersTrigger.Table)
				assert.Equal(t, "BEFORE", usersTrigger.Timing)
				assert.Equal(t, "UPDATE", usersTrigger.Event)
				assert.True(t, usersTrigger.ForEachRow)

				// Posts trigger kontrolü
				postsTrigger := schema.Triggers[1]
				assert.Equal(t, "posts_update_timestamp", postsTrigger.Name)
				assert.Equal(t, "posts", postsTrigger.Table)
				assert.Equal(t, "BEFORE", postsTrigger.Timing)
				assert.Equal(t, "UPDATE", postsTrigger.Event)
				assert.True(t, postsTrigger.ForEachRow)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewOracle()
			schema, err := p.Parse(tt.content)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, schema)
			}
		})
	}
}

func TestOracle_Generate(t *testing.T) {
	tests := []struct {
		name     string
		schema   *sqlporter.Schema
		wantErr  bool
		validate func(*testing.T, string)
	}{
		{
			name:    "Empty schema",
			schema:  nil,
			wantErr: true,
		},
		{
			name: "Full schema",
			schema: &sqlporter.Schema{
				Name: "example_db",
				Tables: []sqlporter.Table{
					{
						Name: "users",
						Columns: []sqlporter.Column{
							{Name: "id", DataType: "NUMBER", DefaultValue: "users_seq.NEXTVAL", IsPrimaryKey: true},
							{Name: "username", DataType: "VARCHAR2", Length: 50, IsNullable: false, IsUnique: true},
							{Name: "email", DataType: "VARCHAR2", Length: 100, IsNullable: false},
							{Name: "status", DataType: "VARCHAR2", Length: 20, DefaultValue: "'active'"},
						},
					},
				},
				Sequences: []sqlporter.Sequence{
					{Name: "users_seq", StartValue: 1, IncrementBy: 1},
				},
				Views: []sqlporter.View{
					{
						Name: "active_users_view",
						Definition: `SELECT u.*, COUNT(p.id) as post_count
							FROM users u LEFT JOIN posts p ON u.id = p.user_id
							WHERE u.status = 'active'
							GROUP BY u.id`,
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "CREATE TABLE users")
				assert.Contains(t, output, "CREATE SEQUENCE users_seq")
				assert.Contains(t, output, "CREATE VIEW active_users_view")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewOracle()
			output, err := o.Generate(tt.schema)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, output)
			}
		})
	}
}
