package logy

import (
	"bytes"
	"fmt"

	//"io"
	//"os"
	"runtime"
	"sort"

	//"strconv"
	//"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const (
	red    = 31
	yellow = 33
	blue   = 36
	gray   = 37
)

var baseTimestamp time.Time

func init() {
	baseTimestamp = time.Now()
}

// TextFormatter formats logs into text
type TextFormatter struct {
	ForceQuote bool

	DisableQuote bool

	DisableTimestamp bool

	FullTimestamp bool

	TimestampFormat string

	DisableSorting bool

	SortingFunc func([]string)

	DisableLevelTruncation bool

	PadLevelText bool

	// QuoteEmptyFields will wrap empty fields in quotes if true
	QuoteEmptyFields bool

	// Whether the logger's out is to a terminal
	isTerminal bool

	FieldMap FieldMap

	CallerPrettyfier func(*runtime.Frame) (function string, file string)

	terminalInitOnce sync.Once

	// The max length of the level text, generated dynamically on init
	levelTextMaxLength int
}

func (f *TextFormatter) init(entry *Entry) {
	// Get the max length of the level text
	for _, level := range AllLevels {
		levelTextLength := utf8.RuneCount([]byte(level.String()))
		if levelTextLength > f.levelTextMaxLength {
			f.levelTextMaxLength = levelTextLength
		}
	}
}

// Format renders a single log entry
func (f *TextFormatter) Format(entry *Entry) ([]byte, error) {
	//println("txtformat")
	data := make(Fields)
	//println(data)
	for k, v := range entry.Data {
		data[k] = v
	}
	prefixFieldClashes(data, f.FieldMap, entry.HasCaller())
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}

	var funcVal, fileVal string

	fixedKeys := make([]string, 0, 4+len(data))
	if !f.DisableTimestamp {
		fixedKeys = append(fixedKeys, f.FieldMap.resolve(FieldKeyTime))
	}
	fixedKeys = append(fixedKeys, f.FieldMap.resolve(FieldKeyLevel))
	if entry.Message != "" {
		fixedKeys = append(fixedKeys, f.FieldMap.resolve(FieldKeyMsg))
	}
	if entry.err != "" {
		fixedKeys = append(fixedKeys, f.FieldMap.resolve(FieldKeyLogrusError))
	}
	if entry.HasCaller() {
		if f.CallerPrettyfier != nil {
			funcVal, fileVal = f.CallerPrettyfier(entry.Caller)
		} else {
			funcVal = entry.Caller.Function
			fileVal = fmt.Sprintf("%s:%d", entry.Caller.File, entry.Caller.Line)
		}

		if funcVal != "" {
			fixedKeys = append(fixedKeys, f.FieldMap.resolve(FieldKeyFunc))
		}
		if fileVal != "" {
			fixedKeys = append(fixedKeys, f.FieldMap.resolve(FieldKeyFile))
		}
	}

	if !f.DisableSorting {
		if f.SortingFunc == nil {
			sort.Strings(keys)
			fixedKeys = append(fixedKeys, keys...)
		} else {
			//if !f.isColored() {
			//	fixedKeys = append(fixedKeys, keys...)
			//	f.SortingFunc(fixedKeys)
			//} else {
			f.SortingFunc(keys)
			//}
		}
	} else {
		fixedKeys = append(fixedKeys, keys...)
	}

	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	f.terminalInitOnce.Do(func() { f.init(entry) })

	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = defaultTimestampFormat
	}
	//if f.isColored() {
	//	f.printColored(b, entry, keys, data, timestampFormat)
	//} else {
	for _, key := range fixedKeys {
		var value interface{}
		switch {
		case key == f.FieldMap.resolve(FieldKeyTime):
			value = entry.Time.Format(timestampFormat)
		case key == f.FieldMap.resolve(FieldKeyLevel):
			value = entry.Level.String()
		case key == f.FieldMap.resolve(FieldKeyMsg):
			value = entry.Message
		case key == f.FieldMap.resolve(FieldKeyLogrusError):
			value = entry.err
		case key == f.FieldMap.resolve(FieldKeyFunc) && entry.HasCaller():
			value = funcVal
		case key == f.FieldMap.resolve(FieldKeyFile) && entry.HasCaller():
			value = fileVal
		default:
			value = data[key]
		}
		f.appendKeyValue(b, key, value)
	}
	//}

	b.WriteByte('\n')
	return b.Bytes(), nil
}

func (f *TextFormatter) needsQuoting(text string) bool {
	if f.ForceQuote {
		return true
	}
	if f.QuoteEmptyFields && len(text) == 0 {
		return true
	}
	if f.DisableQuote {
		return false
	}
	for _, ch := range text {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' || ch == '.' || ch == '_' || ch == '/' || ch == '@' || ch == '^' || ch == '+') {
			return true
		}
	}
	return false
}

func (f *TextFormatter) appendKeyValue(b *bytes.Buffer, key string, value interface{}) {
	if b.Len() > 0 {
		b.WriteByte(' ')
	}
	b.WriteString(key)
	b.WriteByte('=')
	f.appendValue(b, value)
}

func (f *TextFormatter) appendValue(b *bytes.Buffer, value interface{}) {
	stringVal, ok := value.(string)
	if !ok {
		stringVal = fmt.Sprint(value)
	}

	if !f.needsQuoting(stringVal) {

		b.WriteString(stringVal)

	} else {

		b.WriteString(fmt.Sprintf("%q", stringVal))
	}
}
