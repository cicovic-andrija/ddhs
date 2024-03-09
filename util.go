package main

import (
	"crypto/rand"
	"encoding/hex"
)

func RandHexString(n int) (string, error) {
	if n < 1 {
		return "", nil
	}
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func Paginate[T any](s []T, pageNum int, pageSize int) []T {
	if len(s) == 0 || pageNum < 0 || pageSize < 1 || pageNum*pageSize >= len(s) {
		return nil
	}
	start := pageNum * pageSize
	end := (pageNum + 1) * pageSize
	if end > len(s) {
		end = len(s)
	}
	return s[start:end]
}

func Filter[T any](s []T, predicate func(T) bool) []T {
	filtered := make([]T, 0, len(s))
	for _, e := range s {
		if predicate(e) {
			filtered = append(filtered, e)
		}
	}
	return filtered
}
