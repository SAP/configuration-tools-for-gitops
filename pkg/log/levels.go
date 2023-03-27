package log

import (
	"fmt"
	"strconv"
)

// type SLevel string

// Level holds the logging level. It can be set either via integers [-1,,5]
// or via the associated string names: Debug(-1), Info(0), Warn(1), Error(2),
// DPanic(3), Panic(4), Fatal(5).
// Default: Info(0)
type Level struct {
	number   int
	readible string
}

const (
	sDebug  = "Debug"
	sInfo   = "Info"
	sWarn   = "Warn"
	sError  = "Error"
	sDPanic = "DPanic"
	sPanic  = "Panic"
	sFatal  = "Fatal"
)

var (
	debug  = Level{-1, "Debug"}
	info   = Level{0, "Info"}
	warn   = Level{1, "Warn"}
	errors = Level{2, "Error"}
	dpanic = Level{3, "DPanic"}
	panics = Level{4, "Panic"}
	fatal  = Level{5, "Fatal"}

	allLevels = []Level{debug, info, warn, errors, dpanic, panics, fatal}
)

// New creates a new logging Level from a string input.
//
//	Allowed inputs: [Debug, Info, Warn, Error, DPanic, Panic, Fatal].
//	All other inputs will result in log Level Debug.
func New(level string) Level {
	switch level {
	case debug.readible:
		return debug
	case info.readible:
		return info
	case warn.readible:
		return warn
	case errors.readible:
		return errors
	case dpanic.readible:
		return dpanic
	case panics.readible:
		return panics
	case fatal.readible:
		return fatal
	default:
		return debug
	}
}

func Debug() Level {
	return debug
}
func Info() Level {
	return info
}
func Warn() Level {
	return warn
}
func Error() Level {
	return errors
}
func DPanic() Level {
	return dpanic
}
func Panic() Level {
	return panics
}
func Fatal() Level {
	return fatal
}

// // NewFromInt creates a new logging Level.
// //
// //	Available levels: [Debug(-1), Info(0), Warn(1), Error(2), DPanic(3), Panic(4), Fatal(5)].
// //	Inputs larger than 5 will result in Level Fatal(5).
// //	Inputs smaller than -1 will result in Level Debug(-1).
func NewFromInt(i int) Level {
	if i <= debug.number {
		return debug
	}
	switch i {
	case info.number:
		return info
	case warn.number:
		return warn
	case errors.number:
		return errors
	case dpanic.number:
		return dpanic
	case panics.number:
		return panics
	case fatal.number:
		return fatal
	default:
		return fatal
	}
}

func (l Level) Is(input Level) bool {
	return l.number == input.number
}

func (l Level) Compare(i Level) int {
	return l.number - i.number
}

// AllLevels returns a printed map of all possible logging inputs
// (keys and values are valid inputs)
func (l Level) AllLevels() string {
	allLvls := []string{}
	for _, l := range allLevels {
		allLvls = append(allLvls, fmt.Sprintf("%s:%v", l.readible, l.number))
	}
	return fmt.Sprintf("map%v", allLvls)
}

// Set validates a Level input and overwrites the objects Level
func (l *Level) Set(v string) error {
	var lvl Level
	ilvl, err := strconv.Atoi(v)
	if err == nil {
		lvl = NewFromInt(ilvl)
	} else {
		lvl = New(v)
		if lvl.Is(Debug()) && v != sDebug {
			return fmt.Errorf(
				"illegal Level \"%v\" found - Level must be either a key or value from %+v",
				v, l.AllLevels(),
			)
		}
	}
	l.number = lvl.number
	l.readible = lvl.readible
	return nil
}

// Type returns the type of the Level
func (l Level) Type() string {
	return "level"
}

// HumanReadable returns the Logging Level in plain string form
func (l Level) String() string {
	return l.readible
}

func (l Level) AsInt() int {
	return l.number
}

// String returns the Logging Level in string form with integer form in parentheses
func (l Level) Print() string {
	return fmt.Sprintf("%s(%v)", l.readible, l.number)
}
