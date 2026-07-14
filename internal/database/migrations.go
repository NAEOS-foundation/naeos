package database

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var migrationPattern = regexp.MustCompile(`^(\d+)_(.+)\.(up|down)\.sql$`)

type MigrationFile struct {
	Version  int
	Name     string
	UpFile   string
	DownFile string
	UpSQL    string
	DownSQL  string
}

func LoadMigrations(dir string) ([]Migration, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read migration directory: %w", err)
	}

	files := make(map[int]*MigrationFile)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		matches := migrationPattern.FindStringSubmatch(entry.Name())
		if matches == nil {
			continue
		}

		version, err := strconv.Atoi(matches[1])
		if err != nil {
			continue
		}

		name := matches[2]
		direction := matches[3]

		mf, ok := files[version]
		if !ok {
			mf = &MigrationFile{Version: version, Name: name}
			files[version] = mf
		}

		fullPath := filepath.Join(dir, entry.Name())
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, fmt.Errorf("read migration file %s: %w", entry.Name(), err)
		}

		switch direction {
		case "up":
			mf.UpFile = entry.Name()
			mf.UpSQL = string(content)
		case "down":
			mf.DownFile = entry.Name()
			mf.DownSQL = string(content)
		}
	}

	var versions []int
	for v := range files {
		versions = append(versions, v)
	}
	sort.Ints(versions)

	var migrations []Migration
	for _, v := range versions {
		mf := files[v]
		migrations = append(migrations, Migration{
			Version: mf.Version,
			Name:    mf.Name,
			Up:      mf.UpSQL,
			Down:    mf.DownSQL,
		})
	}

	return migrations, nil
}

func MigrationChecksum(migrations []Migration) string {
	var sb strings.Builder
	for _, m := range migrations {
		sb.WriteString(fmt.Sprintf("%d:%s:%s:%s\n", m.Version, m.Name, m.Up, m.Down))
	}
	return fmt.Sprintf("%x", []byte(sb.String()))
}
