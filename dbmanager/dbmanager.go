package dbmanager

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var (
	// Database connection pool
	databases = make(map[string]*sql.DB)

	// Connection statuses
	connectionStatuses = map[string]bool{
		"sqlite":     false,
		"mysql":      false,
		"postgresql": false,
	}

	// Connection strings for Docker containers
	connectionStrings = map[string]string{
		"sqlite":     "./testdb.sqlite",
		"mysql":      "root:example@tcp(mysql:3306)/testdb",
		"postgresql": "postgres://postgres:example@postgres:5432/testdb?sslmode=disable",
	}
)

// InitDatabases initializes connections to all configured databases
func InitDatabases() error {
	// Check if we're running in a local environment (not Docker)
	if _, err := os.Stat("/.dockerenv"); os.IsNotExist(err) {
		// Running locally, use localhost
		connectionStrings["mysql"] = "root:example@tcp(localhost:3306)/testdb"
		connectionStrings["postgresql"] = "postgres://postgres:example@localhost:5432/testdb?sslmode=disable"
	}

	var lastError error

	// Initialize SQLite as it doesn't require a server
	if err := initSQLite(); err != nil {
		lastError = err
		fmt.Printf("SQLite initialization error: %v\n", err)
	} else {
		connectionStatuses["sqlite"] = true
	}

	// Try to connect to MySQL Docker container
	go connectWithRetry("mysql", "mysql", 5)

	// Try to connect to PostgreSQL Docker container
	go connectWithRetry("postgresql", "postgres", 5)

	return lastError
}

// GetDatabaseConnection returns the database connection for the specified dialect
func GetDatabaseConnection(dialect string) (*sql.DB, error) {
	db, ok := databases[dialect]
	if !ok {
		return nil, fmt.Errorf("no database connection available for %s", dialect)
	}

	// Test if the connection is still valid
	if err := db.Ping(); err != nil {
		// Try to reconnect
		connectionStatuses[dialect] = false
		connectWithRetry(dialect, dialectToDriver(dialect), 1)

		// Get the connection again
		db, ok = databases[dialect]
		if !ok {
			return nil, fmt.Errorf("failed to reconnect to %s", dialect)
		}
	}

	return db, nil
}

// GetConnectionStatuses returns the status of all database connections
func GetConnectionStatuses() map[string]bool {
	// Test all connections before returning statuses
	for dialect, db := range databases {
		if db != nil {
			if err := db.Ping(); err != nil {
				connectionStatuses[dialect] = false
			} else {
				connectionStatuses[dialect] = true
			}
		}
	}
	return connectionStatuses
}

// dialectToDriver converts a dialect name to the corresponding driver name
func dialectToDriver(dialect string) string {
	switch dialect {
	case "postgresql":
		return "postgres"
	default:
		return dialect
	}
}

// initSQLite initializes the SQLite database
func initSQLite() error {
	db, err := sql.Open("sqlite3", connectionStrings["sqlite"])
	if err != nil {
		return err
	}

	// Create a test table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS test_data (
		id INTEGER PRIMARY KEY,
		name TEXT,
		value INTEGER
	)`)
	if err != nil {
		return err
	}

	// Insert some test data
	_, err = db.Exec(`DELETE FROM test_data`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`INSERT INTO test_data (id, name, value) VALUES 
		(1, 'Item 1', 100),
		(2, 'Item 2', 200),
		(3, 'Item 3', 300),
		(4, 'Item 4', 400),
		(5, 'Item 5', 500),
		(6, 'Item 6', 600),
		(7, 'Item 7', 700),
		(8, 'Item 8', 800),
		(9, 'Item 9', 900),
		(10, 'Item 10', 1000)
	`)
	if err != nil {
		return err
	}

	databases["sqlite"] = db
	fmt.Println("SQLite database initialized successfully")
	return nil
}

// connectWithRetry attempts to connect to a database with retries
func connectWithRetry(dialect string, driver string, maxRetries int) {
	for i := 0; i < maxRetries; i++ {
		fmt.Printf("Attempting to connect to %s (attempt %d/%d)\n", dialect, i+1, maxRetries)

		if connected := tryConnect(dialect, driver); connected {
			break
		}

		// Wait before retrying
		time.Sleep(5 * time.Second)
	}
}

// tryConnect attempts to connect to a database
func tryConnect(dialect string, driver string) bool {
	db, err := sql.Open(driver, connectionStrings[dialect])
	if err != nil {
		fmt.Printf("Failed to open %s connection: %v\n", dialect, err)
		return false
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		fmt.Printf("Failed to ping %s database: %v\n", dialect, err)
		return false
	}

	// Apply safety settings for the database
	if err := SetSafeDatabaseDefaults(db, dialect); err != nil {
		fmt.Printf("Warning: Failed to set safe defaults for %s: %v\n", dialect, err)
	}

	// Apply transaction limits
	if err := ApplyTransactionLimits(db, dialect); err != nil {
		fmt.Printf("Warning: Failed to set transaction limits for %s: %v\n", dialect, err)
	}

	// Connection successful, initialize the database
	err = initDatabase(db, dialect)
	if err != nil {
		fmt.Printf("Failed to initialize %s database: %v\n", dialect, err)
		return false
	}

	// Set connection pool limits to prevent resource exhaustion
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(30 * time.Minute)

	// Store the connection
	databases[dialect] = db
	connectionStatuses[dialect] = true
	fmt.Printf("%s database connected and initialized successfully\n", dialect)
	return true
}

// initDatabase initializes database schema and sample data
func initDatabase(db *sql.DB, dialect string) error {
	switch dialect {
	case "mysql":
		return initMySQLDatabase(db)
	case "postgresql":
		return initPostgreSQLDatabase(db)
	default:
		return nil
	}
}

// initMySQLDatabase initializes MySQL database with sample data
func initMySQLDatabase(db *sql.DB) error {
	// Create tables
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS products (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		description TEXT,
		price DECIMAL(10,2) NOT NULL,
		category VARCHAR(50),
		stock INT DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return err
	}

	// Check if we need to insert sample data
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM products").Scan(&count)
	if err != nil || count == 0 {
		// Insert sample products
		_, err = db.Exec(`INSERT INTO products (name, description, price, category, stock) VALUES
			('Laptop', 'High-performance laptop with SSD', 899.99, 'Electronics', 45),
			('Smartphone', 'Latest model with dual camera', 699.99, 'Electronics', 120),
			('Coffee Maker', 'Premium coffee machine', 89.99, 'Kitchen', 30),
			('Headphones', 'Noise cancelling wireless headphones', 199.99, 'Audio', 75),
			('Monitor', '27-inch 4K monitor', 349.99, 'Computer Accessories', 25),
			('Office Chair', 'Ergonomic office chair', 249.99, 'Furniture', 15),
			('Tablet', '10-inch tablet with stylus', 429.99, 'Electronics', 35),
			('Smart Watch', 'Fitness tracking smart watch', 159.99, 'Wearables', 50),
			('Desk', 'Modern computer desk', 179.99, 'Furniture', 10),
			('Keyboard', 'Mechanical gaming keyboard', 129.99, 'Computer Accessories', 40),
			('Mouse', 'Wireless gaming mouse', 59.99, 'Computer Accessories', 60),
			('Speakers', 'Bluetooth speakers', 79.99, 'Audio', 45),
			('External SSD', '1TB portable SSD drive', 149.99, 'Storage', 30),
			('Webcam', 'HD webcam for video conferencing', 69.99, 'Computer Accessories', 25),
			('Printer', 'Color laser printer', 299.99, 'Office Equipment', 12)
		`)
		if err != nil {
			return err
		}
	}

	return nil
}

// initPostgreSQLDatabase initializes PostgreSQL database with sample data
func initPostgreSQLDatabase(db *sql.DB) error {
	// Create tables
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS customers (
		id SERIAL PRIMARY KEY,
		first_name VARCHAR(50) NOT NULL,
		last_name VARCHAR(50) NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL,
		phone VARCHAR(20),
		country VARCHAR(50),
		city VARCHAR(50),
		address TEXT,
		postal_code VARCHAR(20),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return err
	}

	// Check if we need to insert sample data
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM customers").Scan(&count)
	if err != nil || count == 0 {
		// Insert sample customers
		_, err = db.Exec(`INSERT INTO customers (first_name, last_name, email, phone, country, city, address, postal_code) VALUES
			('John', 'Doe', 'john.doe@example.com', '555-123-4567', 'USA', 'New York', '123 Broadway St', '10001'),
			('Jane', 'Smith', 'jane.smith@example.com', '555-987-6543', 'USA', 'Los Angeles', '456 Hollywood Blvd', '90028'),
			('Robert', 'Johnson', 'robert.j@example.com', '555-234-5678', 'USA', 'Chicago', '789 Michigan Ave', '60601'),
			('Emily', 'Williams', 'emily.w@example.com', '555-345-6789', 'Canada', 'Toronto', '567 Yonge St', 'M4Y 1Z2'),
			('Michael', 'Brown', 'michael.b@example.com', '555-456-7890', 'UK', 'London', '234 Oxford St', 'W1D 1BS'),
			('Sarah', 'Davis', 'sarah.d@example.com', '555-567-8901', 'Australia', 'Sydney', '890 George St', '2000'),
			('David', 'Miller', 'david.m@example.com', '555-678-9012', 'Germany', 'Berlin', '123 Unter den Linden', '10117'),
			('Jennifer', 'Wilson', 'jennifer.w@example.com', '555-789-0123', 'France', 'Paris', '456 Champs-Élysées', '75008'),
			('James', 'Taylor', 'james.t@example.com', '555-890-1234', 'Japan', 'Tokyo', '789 Shibuya', '150-0002'),
			('Lisa', 'Anderson', 'lisa.a@example.com', '555-901-2345', 'Italy', 'Rome', '890 Via del Corso', '00186'),
			('Thomas', 'Jackson', 'thomas.j@example.com', '555-012-3456', 'Spain', 'Madrid', '123 Gran Via', '28013'),
			('Patricia', 'White', 'patricia.w@example.com', '555-123-4567', 'Brazil', 'Rio de Janeiro', '456 Copacabana', '22070'),
			('Richard', 'Harris', 'richard.h@example.com', '555-234-5678', 'USA', 'San Francisco', '789 Market St', '94103'),
			('Elizabeth', 'Clark', 'elizabeth.c@example.com', '555-345-6789', 'USA', 'Boston', '890 Newbury St', '02115')
		`)
		if err != nil {
			return err
		}
	}

	return nil
}
