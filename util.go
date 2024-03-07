package main

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
