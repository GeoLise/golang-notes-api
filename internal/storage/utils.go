package storage

import (
	"strings"
	"time"
)

func getCacheKey(keys ...string) string {

	return strings.Join(keys, ":")

}

const DEFAULT_TTL = time.Second * 60 * 60 * 24
