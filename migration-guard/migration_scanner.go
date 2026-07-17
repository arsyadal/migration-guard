package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type MigrationPair struct {
	UpPath   string
	DownPath string
	BaseName string
}

func ScanDirectory(dir string) ([]MigrationPair, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	pairMap := make(map[string]*MigrationPair)
	var keys []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}

		isUp := strings.HasSuffix(name, ".up.sql")
		isDown := strings.HasSuffix(name, ".down.sql")
		if !isUp && !isDown {
			continue
		}

		baseName := ""
		if isUp {
			baseName = strings.TrimSuffix(name, ".up.sql")
		} else {
			baseName = strings.TrimSuffix(name, ".down.sql")
		}

		pair, exists := pairMap[baseName]
		if !exists {
			pair = &MigrationPair{BaseName: baseName}
			pairMap[baseName] = pair
			keys = append(keys, baseName)
		}

		fullPath := filepath.Join(dir, name)
		if isUp {
			pair.UpPath = fullPath
		} else {
			pair.DownPath = fullPath
		}
	}

	sort.Strings(keys)
	var pairs []MigrationPair
	for _, k := range keys {
		pairs = append(pairs, *pairMap[k])
	}

	return pairs, nil
}
