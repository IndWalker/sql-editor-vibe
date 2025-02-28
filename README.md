# SQL Playground Server

A Go-based SQL playground that allows testing SQL queries against SQLite, MySQL, and PostgreSQL databases.
Almost entirely vibecoded with Github Copilot.

## Features

- Web interface for executing SQL queries
- Support for SQLite, MySQL, and PostgreSQL dialects
- Query validation
- Results displayed in a tabular format
- Database connection status monitoring

## Prerequisites

- Docker and Docker Compose

## Getting Started

1. Clone the repository
2. Run the application with Docker Compose:

```bash
docker-compose up
```

3. Access the web interface at `http://localhost:8080`

## Database Information

### SQLite
- Location: Local file `testdb.sqlite`
- Sample table: `test_data`

### MySQL
- Host: localhost:3306
- Username: root
- Password: example
- Database: testdb
- Sample table: `products`

### PostgreSQL
- Host: localhost:5432
- Username: postgres
- Password: example
- Database: testdb
- Sample table: `customers`

## Example Queries

### SQLite
```sql
SELECT * FROM test_data WHERE value > 300
```

### MySQL
```sql
SELECT * FROM products WHERE category = 'Electronics' ORDER BY price DESC
```

### PostgreSQL
```sql
SELECT * FROM customers WHERE country = 'USA'
```
