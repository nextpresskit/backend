package migrate

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	_ "github.com/lib/pq"
)

const (
	upMarker   = "-- +migrate Up"
	downMarker = "-- +migrate Down"
)

type Migration struct {
	Version int64
	Name    string
	Up      string
	Down    string
}

type Runner struct {
	db     *sql.DB
	dir    string
	parsed []Migration
}

func NewRunner(dsn string, dir string) (*Runner, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}
	r := &Runner{db: db, dir: dir}
	if err := r.parse(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return r, nil
}

func (r *Runner) Close() error {
	return r.db.Close()
}

func (r *Runner) parse() error {
	ents, err := os.ReadDir(r.dir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}
	re := regexp.MustCompile(`^(\d+)_(.+)\.sql$`)
	var list []Migration
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		m := re.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		var ver int64
		if _, err := fmt.Sscanf(m[1], "%d", &ver); err != nil {
			continue
		}
		path := filepath.Join(r.dir, e.Name())
		raw, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		up, down := splitUpDown(string(raw))
		list = append(list, Migration{Version: ver, Name: m[2], Up: up, Down: down})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Version < list[j].Version })
	r.parsed = list
	return nil
}

func splitUpDown(s string) (up, down string) {
	s = strings.TrimSpace(s)
	idxUp := strings.Index(s, upMarker)
	idxDown := strings.Index(s, downMarker)
	if idxUp < 0 || idxDown < 0 {
		return "", ""
	}
	up = strings.TrimSpace(s[idxUp+len(upMarker) : idxDown])
	down = strings.TrimSpace(s[idxDown+len(downMarker):])
	return up, down
}

func (r *Runner) ensureSchemaMigrations() error {
	_, err := r.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT NOT NULL PRIMARY KEY,
			dirty BOOLEAN NOT NULL DEFAULT FALSE
		)
	`)
	return err
}

func (r *Runner) currentVersion() (int64, bool, error) {
	var ver int64
	var dirty bool
	err := r.db.QueryRow(`SELECT version, dirty FROM schema_migrations ORDER BY version DESC LIMIT 1`).Scan(&ver, &dirty)
	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return ver, dirty, nil
}

func (r *Runner) Up(steps int) error {
	if err := r.ensureSchemaMigrations(); err != nil {
		return err
	}
	current, dirty, err := r.currentVersion()
	if err != nil {
		return err
	}
	if dirty {
		return fmt.Errorf("migration %d is dirty; fix or force version", current)
	}
	var ran int
	for _, m := range r.parsed {
		if m.Version <= current {
			continue
		}
		if steps > 0 && ran >= steps {
			break
		}
		if err := r.runUp(m); err != nil {
			return fmt.Errorf("migrate up %d %s: %w", m.Version, m.Name, err)
		}
		ran++
	}
	if ran > 0 {
		log.Printf("Applied %d migration(s)", ran)
	}
	return nil
}

func (r *Runner) runUp(m Migration) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(m.Up); err != nil {
		return err
	}
	if _, err := tx.Exec(`INSERT INTO schema_migrations (version, dirty) VALUES ($1, FALSE)`, m.Version); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Runner) Down(steps int) error {
	if err := r.ensureSchemaMigrations(); err != nil {
		return err
	}
	current, dirty, err := r.currentVersion()
	if err != nil {
		return err
	}
	if dirty {
		return fmt.Errorf("migration %d is dirty; fix or force version", current)
	}
	if current == 0 {
		return nil
	}
	var last *Migration
	for i := len(r.parsed) - 1; i >= 0; i-- {
		if r.parsed[i].Version == current {
			last = &r.parsed[i]
			break
		}
	}
	if last == nil {
		return fmt.Errorf("no migration found for version %d", current)
	}
	if err := r.runDown(*last); err != nil {
		return fmt.Errorf("migrate down %d %s: %w", last.Version, last.Name, err)
	}
	log.Printf("Rolled back migration %d %s", last.Version, last.Name)
	if steps == 0 {
		return r.Down(0)
	}
	if steps > 1 {
		return r.Down(steps - 1)
	}
	return nil
}

func (r *Runner) runDown(m Migration) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(m.Down); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM schema_migrations WHERE version = $1`, m.Version); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Runner) Version() (int64, bool, error) {
	if err := r.ensureSchemaMigrations(); err != nil {
		return 0, false, err
	}
	return r.currentVersion()
}

func (r *Runner) Force(version int) error {
	if err := r.ensureSchemaMigrations(); err != nil {
		return err
	}
	_, err := r.db.Exec(`
		DELETE FROM schema_migrations;
		INSERT INTO schema_migrations (version, dirty) VALUES ($1, FALSE)
	`, version)
	return err
}

// Drop drops all tables in the public schema (including schema_migrations), so migrate-up can run from scratch.
func (r *Runner) Drop() error {
	_, err := r.db.Exec(`
		DO $$ DECLARE r RECORD;
		BEGIN
		  FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public')
		  LOOP
		    EXECUTE 'DROP TABLE IF EXISTS public.' || quote_ident(r.tablename) || ' CASCADE';
		  END LOOP;
		END $$;
	`)
	return err
}

