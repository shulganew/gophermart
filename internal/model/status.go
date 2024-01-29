package model

type Status string

const (
	NEW        Status = "NEW"
	PROCESSING Status = "PROCESSING"
	INVALID    Status = "INVALID"
	PROCESSED  Status = "PROCESSED"
	REGISTERED Status = "REGISTERED"
)

func (s *Status) String() string {
	return string(*s)
}
