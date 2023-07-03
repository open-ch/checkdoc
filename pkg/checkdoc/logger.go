package checkdoc

// Logger wraps the the main components of a logger like logrus and allows us
// to avoid coupling the package to a specific logging library.
type Logger interface {
	Debugf(format string, args ...any)

	Infof(format string, args ...any)

	Warnf(format string, args ...any)

	Errorf(format string, args ...any)

	Fatalf(format string, args ...any)

	Panicf(format string, args ...any)

	// WithFields(keyValues Fields) Logger
}
