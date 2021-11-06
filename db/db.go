package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"time"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v4/log/zapadapter"
	"github.com/jackc/pgx/v4/pgxpool"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/starshine-sys/oodles/common"

	// pgx driver for migrations
	_ "github.com/jackc/pgx/v4/stdlib"
)

// DB ...
type DB struct {
	*pgxpool.Pool

	BotConfig common.BotConfig
	Config    Config
	Perms     PermissionConfig
	Overrides CommandOverrides
}

// New returns a new DB
func New(conf common.BotConfig) (*DB, error) {
	err := runMigrations(conf.Database)
	if err != nil {
		return nil, err
	}

	pgxconf, err := pgxpool.ParseConfig(conf.Database)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse config: %w", err)
	}
	pgxconf.ConnConfig.LogLevel = pgx.LogLevelWarn
	pgxconf.ConnConfig.Logger = zapadapter.NewLogger(common.Log.Desugar())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.ConnectConfig(ctx, pgxconf)
	if err != nil {
		return nil, fmt.Errorf("Unable to connect to database: %w", err)
	}

	db := &DB{
		Pool:      pool,
		BotConfig: conf,

		Config:    make(Config),
		Overrides: make(CommandOverrides),
	}

	if err = db.createConfig(); err != nil {
		return nil, err
	}

	if err = db.fetchConfig(); err != nil {
		return nil, err
	}

	return db, nil
}

//go:embed migrations
var fs embed.FS

func runMigrations(url string) (err error) {
	db, err := sql.Open("pgx", url)
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	migrations := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: fs,
		Root:       "migrations",
	}

	n, err := migrate.Exec(db, "postgres", migrations, migrate.Up)
	if err != nil {
		return err
	}

	if n != 0 {
		common.Log.Infof("Performed %v migrations!", n)
	}

	err = db.Close()
	return err
}
