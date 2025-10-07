package jsonlog

//Creating a custom logger, mostly a thin wrapper around io.Writer with helpers
//iota creates log levels, sync.Mutex to do the writing
import (
	"encoding/json"
	"io"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

// creating custom logging which is kinda overkill for this
type Level int8

// init constats which rep a severity level. Use Iota keyword
const (
	LevelInfo  Level = iota // Has value 0
	LevelError              //Balue 1
	LevelFatal              //Value is 2
	LevelOff                //Value of 3
)

// return human string for sev level
func (l Level) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return ""
	}
}

// custom logger type for our output destination to write log to
// Min sev holds
type Logger struct {
	out      io.Writer
	minLevel Level
	mu       sync.Mutex
}

// new logger instance to write log entries at or above Min Sev
func New(out io.Writer, minLevel Level) *Logger {
	return &Logger{
		out:      out,
		minLevel: minLevel,
	}
}

// declare helper method for writing log entries at diff levels
// all accept a map as second param with props
func (l *Logger) PrintInfo(message string, properties map[string]string) {
	l.print(LevelInfo, message, properties)
}

func (l *Logger) PrintError(err error, properties map[string]string) {
	l.print(LevelError, err.Error(), properties)
}
func (l *Logger) PrintFatal(err error, properties map[string]string) {
	l.print(LevelFatal, err.Error(), properties)
	os.Exit(1) //for stuff at fata level, will term app too
}

// Prtin is a internal method for writing the log entry
func (l *Logger) print(level Level, message string, properties map[string]string) (int, error) {
	//if sev level is below min sev for logger, then return with no further action
	if level < l.minLevel {
		return 0, nil
	}
	//create anon struct holding data for the log entry
	aux := struct {
		Level      string            `json:"level"`
		Time       string            `json:"time"`
		Message    string            `json:"message"`
		Properties map[string]string `json:"properties,omitempty"`
		Trace      string            `json:"trace,omitempty"`
	}{
		Level:      level.String(),
		Time:       time.Now().UTC().Format(time.RFC3339),
		Message:    message,
		Properties: properties,
	}
	//stack trace for stuff at error and FATAL levels
	if level >= LevelError {
		aux.Trace = string(debug.Stack())
	}
	//declare line var for holding log entry text
	var line []byte
	//marshal anon struct to JSON and store in line var
	//if problem, display that err instead
	line, err := json.Marshal(aux)
	if err != nil {
		line = []byte(LevelError.String() + ": unable to marshal log message:" + err.Error())
	}
	//lock mutex so no two writes to output can happen concurrently.
	//If we dont do this, possible two or more log entries could be comingled
	l.mu.Lock()
	defer l.mu.Unlock()
	//write log entry by new line
	return l.out.Write(append(line, '\n'))
}

// create a write() method on logger type to allow it to be in the
// io.writer interface to write log entry at ERROR level with no extra props
func (l *Logger) Write(message []byte) (n int, err error) {
	return l.print(LevelError, string(message), nil)
}
