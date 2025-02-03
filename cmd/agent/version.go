//go:build version

package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"time"
)

// init() выполнится только если мы собираем с тегом "version".
func init() {
	// Получаем дату сборки (UTC).
	buildDate = time.Now().UTC().Format(time.RFC3339)

	// Попробуем достать короткий хеш коммита из Git.
	commit, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		fmt.Println("Failed to retrieve git commit: ", err)
		buildCommit = "unknown"
	} else {
		buildCommit = string(bytes.TrimSpace(commit))
	}

	// Пытаемся получить последний тег.
	tag, err := exec.Command("git", "describe", "--tags", "--abbrev=0").Output()
	if err != nil {
		fmt.Println("Failed to retrieve git tag: ", err)
		buildVersion = "dev"
	} else {
		buildVersion = string(bytes.TrimSpace(tag))
	}
}
