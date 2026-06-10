package player

import (
	librespot "github.com/devgianlu/go-librespot"
	"log"
)

// stdLogger is a minimal librespot.Logger that forwards Info/Warn/Error to the
// standard logger (so the first-run OAuth2 login URL is visible) while dropping
// the very noisy Trace/Debug levels.
type stdLogger struct {
	prefix string
}

func newSpotifyLogger() *stdLogger { return &stdLogger{prefix: "[spotify] "} }

func (l *stdLogger) Tracef(string, ...interface{})     {}
func (l *stdLogger) Debugf(string, ...interface{})     {}
func (l *stdLogger) Infof(f string, a ...interface{})  { log.Printf(l.prefix+f, a...) }
func (l *stdLogger) Warnf(f string, a ...interface{})  { log.Printf(l.prefix+f, a...) }
func (l *stdLogger) Errorf(f string, a ...interface{}) { log.Printf(l.prefix+f, a...) }

func (l *stdLogger) Trace(...interface{})   {}
func (l *stdLogger) Debug(...interface{})   {}
func (l *stdLogger) Info(a ...interface{})  { log.Print(append([]interface{}{l.prefix}, a...)...) }
func (l *stdLogger) Warn(a ...interface{})  { log.Print(append([]interface{}{l.prefix}, a...)...) }
func (l *stdLogger) Error(a ...interface{}) { log.Print(append([]interface{}{l.prefix}, a...)...) }

func (l *stdLogger) WithField(string, interface{}) librespot.Logger { return l }
func (l *stdLogger) WithError(error) librespot.Logger               { return l }
