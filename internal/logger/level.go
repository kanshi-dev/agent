package logger

type Level int

const (
	DEBUG Level = iota
	INFO
	WARNING
	ERROR
)

func ParseLevel(s string) Level {
	switch s {
	case "debug":
		return DEBUG
	case "warn":
		return WARNING
	case "error":
		return ERROR
	default:
		return INFO
	}
}
