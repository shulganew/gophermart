package model

// 0-NEW, 1-PROCESSING, 2-INVALID, 3-PROCESSED, 4-REGISTERED.
type Status string

const (
	NEW        Status = "NEW"
	PROCESSING Status = "PROCESSING"
	INVALID    Status = "INVALID"
	PROCESSED  Status = "PROCESSED"
	REGISTERED Status = "REGISTERED"
)
