package starter

type Logger interface {
	Println(v ...any)
	Printf(format string, v ...any)
}
