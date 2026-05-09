package database

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"time"

	// modernc.org/sqlite — pure Go SQLite driver, no CGO required
	_ "modernc.org/sqlite"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// NewDatabase 创建 SQLite 数据库连接。
// 使用 modernc.org/sqlite（纯 Go 实现，无需 CGO）。
// 连接池限制为单写者，启用 WAL 模式和 busy_timeout。
func NewDatabase(path string) (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s?cache=shared&_journal_mode=WAL&_busy_timeout=5000", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("database: open failed: %w", err)
	}

	// SQLite 单写者限制
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("database: ping failed: %w", err)
	}

	return db, nil
}

// RunMigrations 使用 golang-migrate 执行数据库迁移。
// migrationsPath 指向包含 .up.sql / .down.sql 文件的目录。
// 空目录或不含待执行迁移时不返回错误。
func RunMigrations(db *sql.DB, migrationsPath string) error {
	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return fmt.Errorf("database: migration driver failed: %w", err)
	}

	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return fmt.Errorf("database: resolve migrations path failed: %w", err)
	}

	// file: + path 通过 URL.Opaque 处理，兼容 Windows（C:/path）和 Unix（/path）
	sourceURL := "file:" + filepath.ToSlash(absPath)

	m, err := migrate.NewWithDatabaseInstance(sourceURL, "sqlite", driver)
	if err != nil {
		return fmt.Errorf("database: migration instance failed: %w", err)
	}

	if err := m.Up(); err != nil {
		// 空目录（无迁移文件）或所有迁移已执行时视为成功
		if errors.Is(err, fs.ErrNotExist) || err == migrate.ErrNoChange {
			return nil
		}
		return fmt.Errorf("database: migration up failed: %w", err)
	}

	return nil
}
