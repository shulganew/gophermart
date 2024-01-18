package model

import "fmt"

type Status int

const (
	NEW        Status = iota
	PROCESSING Status = iota
	INVALID    Status = iota
	PROCESSED  Status = iota
	REGISTERED Status = iota
)

func (s *Status) String() string {
	switch *s {
	case NEW:
		return "NEW"
	case PROCESSING:
		return "PROCESSING"
	case INVALID:
		return "INVALID"
	case PROCESSED:
		return "PROCESSED"
	case REGISTERED:
		return "REGISTERED"
	default:
		return fmt.Sprintf("NEW")
	}
}
func (s *Status) SetStatus(st string) {
	switch st {
	case "NEW":
		*s = 0
	case "PROCESSING":
		*s = 1
	case "INVALID":
		*s = 2
	case "PROCESSED":
		*s = 3
	case "REGISTERED":
		*s = 4
	default:
		*s = 5
	}
}
