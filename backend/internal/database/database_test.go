package database

import (
	"os"
	"path/filepath"
	"testing"
)

// TestNewDatabase 验证数据库连接创建和连接池配置
func TestNewDatabase(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("NewDatabase failed: %v", err)
	}
	defer db.Close()

	// 验证连接可用
	if err := db.Ping(); err != nil {
		t.Fatalf("db.Ping failed: %v", err)
	}

	// 验证连接池限制
	if maxOpen := db.Stats().MaxOpenConnections; maxOpen != 1 {
		t.Fatalf("expected MaxOpenConnections=1, got %d", maxOpen)
	}
}

// TestRunMigrations_EmptyDir 验证空 migrations 目录不会报错
func TestRunMigrations_EmptyDir(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("NewDatabase failed: %v", err)
	}
	defer db.Close()

	migrationsDir := filepath.Join(t.TempDir(), "migrations")
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		t.Fatalf("failed to create migrations dir: %v", err)
	}

	if err := RunMigrations(db, migrationsDir); err != nil {
		t.Fatalf("RunMigrations with empty dir should not error, got: %v", err)
	}
}

// TestRunMigrationsWithFiles 验证有迁移文件时能正常执行
func TestRunMigrationsWithFiles(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("NewDatabase failed: %v", err)
	}
	defer db.Close()

	migrationsDir := filepath.Join(t.TempDir(), "migrations")
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		t.Fatalf("failed to create migrations dir: %v", err)
	}

	// 写入一个简单的 up 迁移文件
	upContent := []byte("CREATE TABLE IF NOT EXISTS _test_migration (\n    id INTEGER PRIMARY KEY AUTOINCREMENT,\n    name TEXT NOT NULL\n);\n")
	if err := os.WriteFile(filepath.Join(migrationsDir, "000001_test.up.sql"), upContent, 0644); err != nil {
		t.Fatalf("failed to write up migration: %v", err)
	}

	// 写入对应的 down 迁移文件
	downContent := []byte("DROP TABLE IF EXISTS _test_migration;\n")
	if err := os.WriteFile(filepath.Join(migrationsDir, "000001_test.down.sql"), downContent, 0644); err != nil {
		t.Fatalf("failed to write down migration: %v", err)
	}

	if err := RunMigrations(db, migrationsDir); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	// 验证表已创建
	var tableName string
	row := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='_test_migration'")
	if err := row.Scan(&tableName); err != nil {
		t.Fatalf("migration table not found: %v", err)
	}
	if tableName != "_test_migration" {
		t.Fatalf("expected _test_migration, got %s", tableName)
	}
}
