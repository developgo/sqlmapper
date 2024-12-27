package oracle

// Oracle'dan diğer veritabanı tiplerine veri tipi dönüşüm haritaları
var (
	// OracleToMySQL Oracle'dan MySQL'e veri tipi dönüşümleri
	OracleToMySQL = map[string]string{
		"NUMBER":        "decimal",
		"NUMBER(1)":     "boolean",
		"NUMBER(3)":     "tinyint",
		"NUMBER(5)":     "smallint",
		"NUMBER(10)":    "int",
		"NUMBER(19)":    "bigint",
		"BINARY_FLOAT":  "float",
		"BINARY_DOUBLE": "double",
		"FLOAT":         "double",
		"CHAR":          "char",
		"NCHAR":         "char",
		"VARCHAR2":      "varchar",
		"NVARCHAR2":     "varchar",
		"CLOB":          "text",
		"NCLOB":         "text",
		"LONG":          "text",
		"BLOB":          "longblob",
		"RAW":           "binary",
		"LONG RAW":      "longblob",
		"DATE":          "datetime",
		"TIMESTAMP":     "datetime",
		"INTERVAL YEAR": "varchar(100)",
		"INTERVAL DAY":  "varchar(100)",
		"XMLTYPE":       "text",
		"ROWID":         "char(18)",
		"UROWID":        "varchar(4000)",
	}

	// OracleToPostgreSQL Oracle'dan PostgreSQL'e veri tipi dönüşümleri
	OracleToPostgreSQL = map[string]string{
		"NUMBER":        "numeric",
		"NUMBER(1)":     "boolean",
		"NUMBER(3)":     "smallint",
		"NUMBER(5)":     "smallint",
		"NUMBER(10)":    "integer",
		"NUMBER(19)":    "bigint",
		"BINARY_FLOAT":  "real",
		"BINARY_DOUBLE": "double precision",
		"FLOAT":         "double precision",
		"CHAR":          "char",
		"NCHAR":         "char",
		"VARCHAR2":      "varchar",
		"NVARCHAR2":     "varchar",
		"CLOB":          "text",
		"NCLOB":         "text",
		"LONG":          "text",
		"BLOB":          "bytea",
		"RAW":           "bytea",
		"LONG RAW":      "bytea",
		"DATE":          "timestamp",
		"TIMESTAMP":     "timestamp",
		"INTERVAL YEAR": "interval",
		"INTERVAL DAY":  "interval",
		"XMLTYPE":       "xml",
		"ROWID":         "char(18)",
		"UROWID":        "varchar(4000)",
	}

	// OracleToSQLServer Oracle'dan SQL Server'a veri tipi dönüşümleri
	OracleToSQLServer = map[string]string{
		"NUMBER":        "decimal",
		"NUMBER(1)":     "bit",
		"NUMBER(3)":     "tinyint",
		"NUMBER(5)":     "smallint",
		"NUMBER(10)":    "int",
		"NUMBER(19)":    "bigint",
		"BINARY_FLOAT":  "real",
		"BINARY_DOUBLE": "float",
		"FLOAT":         "float",
		"CHAR":          "char",
		"NCHAR":         "nchar",
		"VARCHAR2":      "varchar",
		"NVARCHAR2":     "nvarchar",
		"CLOB":          "varchar(max)",
		"NCLOB":         "nvarchar(max)",
		"LONG":          "varchar(max)",
		"BLOB":          "varbinary(max)",
		"RAW":           "varbinary",
		"LONG RAW":      "varbinary(max)",
		"DATE":          "datetime2",
		"TIMESTAMP":     "datetime2",
		"INTERVAL YEAR": "varchar(100)",
		"INTERVAL DAY":  "varchar(100)",
		"XMLTYPE":       "xml",
		"ROWID":         "char(18)",
		"UROWID":        "varchar(4000)",
	}

	// OracleToSQLite Oracle'dan SQLite'a veri tipi dönüşümleri
	OracleToSQLite = map[string]string{
		"NUMBER":        "REAL",
		"NUMBER(1)":     "INTEGER",
		"NUMBER(3)":     "INTEGER",
		"NUMBER(5)":     "INTEGER",
		"NUMBER(10)":    "INTEGER",
		"NUMBER(19)":    "INTEGER",
		"BINARY_FLOAT":  "REAL",
		"BINARY_DOUBLE": "REAL",
		"FLOAT":         "REAL",
		"CHAR":          "TEXT",
		"NCHAR":         "TEXT",
		"VARCHAR2":      "TEXT",
		"NVARCHAR2":     "TEXT",
		"CLOB":          "TEXT",
		"NCLOB":         "TEXT",
		"LONG":          "TEXT",
		"BLOB":          "BLOB",
		"RAW":           "BLOB",
		"LONG RAW":      "BLOB",
		"DATE":          "TEXT",
		"TIMESTAMP":     "TEXT",
		"INTERVAL YEAR": "TEXT",
		"INTERVAL DAY":  "TEXT",
		"XMLTYPE":       "TEXT",
		"ROWID":         "TEXT",
		"UROWID":        "TEXT",
	}
)