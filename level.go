package raven

// Level represents the severity of a log message
type Level byte

const (
    Transient Level = iota  // live progress updates
    Verbose                 // debug info
    Info                    // normal events
    Warning                 // unusual events
    Error                   // something went wrong

    levelMax
    levelMin Level = 0
)

func (l Level) String() string {
    switch l {
    case Transient:
        return "transient"
    case Verbose:
        return "verbose"
    case Info:
        return "info"
    case Warning:
        return "warning"
    case Error:
        return "error"
    }
    return "unknown"
}

// IsValid checks if the level is within the valid range
func (l Level) IsValid() bool {
    return l >= levelMin && l < levelMax
}