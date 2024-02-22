package log

import (
	"context"
	"encoding"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"log/slog"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

const (
	CReset       = "\033[0m"
	CFaint       = "\033[2m"
	CRed         = "\033[91m"
	CBoldWhite   = "\033[1;97m"
	CBoldDefault = "\033[1;39m"
	CBoldYellow  = "\033[1;93m"
	CBoldRedBg   = "\033[1;97;101m"
	CBlueBg      = "\033[45m"
)

type PlainTextHandler struct {
	Level     slog.Level
	Writer    io.Writer
	withAttrs []byte
	mutex     *sync.Mutex
}

func NewPlainTextHandler(writer io.Writer, level slog.Level) *PlainTextHandler {
	return &PlainTextHandler{Writer: writer, Level: level, mutex: &sync.Mutex{}}
}

func (h *PlainTextHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.Level.Level()
}

func (h *PlainTextHandler) Handle(_ context.Context, record slog.Record) error {
	buf := GetBuffer()
	defer buf.Release()
	addTime(buf, record.Time)
	addLevel(buf, record.Level)
	addCaller(buf, record.PC)
	buf.AddStr(record.Message).AddChr(' ')
	if len(h.withAttrs) > 0 {
		buf.AddBytes(h.withAttrs)
	}
	record.Attrs(func(attr slog.Attr) bool {
		addAttr(buf, attr, "")
		return true
	})
	buf.AddChr('\n')
	h.mutex.Lock()
	defer h.mutex.Unlock()
	_, err := h.Writer.Write(*buf)
	return err
}

func (h *PlainTextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	buf := GetBuffer()
	defer buf.Release()
	h2 := *h
	for _, groupAttr := range attrs {
		addAttr(buf, groupAttr, "")
	}
	h2.withAttrs = make([]byte, len(*buf))
	copy(h2.withAttrs, *buf)
	return &h2
}

func (h *PlainTextHandler) WithGroup(name string) slog.Handler {
	panic("unimplemented")
}

func addTime(buf *PooledBuffer, timestamp time.Time) {
	truncatedTime := timestamp.Truncate(time.Millisecond)
	buf.AddStr(CBlueBg)
	*buf = truncatedTime.AppendFormat(*buf, "2006-01-02T15:04:05.000Z07:00")
	buf.AddStr(CReset).AddChr(' ')
}

func addLevel(buf *PooledBuffer, level slog.Level) {
	switch {
	case level < slog.LevelInfo:
		buf.AddStr(CBoldWhite)
		buf.AddStr("[DBG]")
	case level < slog.LevelWarn:
		buf.AddStr(CBoldDefault)
		buf.AddStr("[INF]")
	case level < slog.LevelError:
		buf.AddStr(CBoldYellow)
		buf.AddStr("[WAR]")
	default:
		buf.AddStr(CBoldRedBg)
		buf.AddStr("[ERR]")
	}
	buf.AddStr(CReset)
	buf.AddChr(' ')
}

func addCaller(buf *PooledBuffer, caller uintptr) {
	fs := runtime.CallersFrames([]uintptr{caller})
	f, _ := fs.Next()
	if f.Function != "" {
		dir, file := filepath.Split(f.File)
		buf.AddStr(CFaint).
			AddChr('[').
			AddStr(filepath.Join(filepath.Base(dir), file)).
			AddChr(':').
			AddStr(strconv.Itoa(f.Line)).
			AddChr(']').
			AddStr(CReset).
			AddChr(' ')
	}
}

func addAttr(buf *PooledBuffer, attr slog.Attr, prefix string) {
	attr.Value = attr.Value.Resolve()
	if attr.Value.Kind() == slog.KindGroup {
		if attr.Key != "" {
			prefix = attr.Key + "."
		}
		for _, groupAttr := range attr.Value.Group() {
			addAttr(buf, groupAttr, prefix)
		}
	} else {
		var quote = true
		if strings.HasPrefix(attr.Key, "@") {
			buf.AddChr('\n')
			quote = false
		}
		buf.AddStr(CFaint).AddStr(prefix).AddStr(attr.Key).AddChr('=').AddStr(CReset)
		addValue(buf, attr.Value, quote)
		buf.AddChr(' ')
	}
}

func addValue(buf *PooledBuffer, value slog.Value, quote bool) {
	switch value.Kind() {
	case slog.KindString:
		addString(buf, value.String(), quote)
	case slog.KindInt64:
		*buf = strconv.AppendInt(*buf, value.Int64(), 10)
	case slog.KindUint64:
		*buf = strconv.AppendUint(*buf, value.Uint64(), 10)
	case slog.KindFloat64:
		*buf = strconv.AppendFloat(*buf, value.Float64(), 'g', -1, 64)
	case slog.KindBool:
		*buf = strconv.AppendBool(*buf, value.Bool())
	case slog.KindDuration:
		addString(buf, value.Duration().String(), quote)
	case slog.KindTime:
		addString(buf, value.Time().String(), quote)
	case slog.KindAny:
		switch typedValue := value.Any().(type) {
		case error:
			addString(buf, typedValue.Error(), quote)
			appendStacktrace(buf, typedValue)
		case encoding.TextMarshaler:
			data, _ := typedValue.MarshalText()
			addString(buf, string(data), quote)
		default:
			addString(buf, fmt.Sprintf("%+v", value.Any()), quote)
		}
	}
}

func addString(buf *PooledBuffer, s string, quote bool) {
	if quote && needsQuoting(s) {
		*buf = strconv.AppendQuote(*buf, s)
	} else {
		buf.AddStr(s)
	}
}
func appendStacktrace(buf *PooledBuffer, err error) {
	type HasUnwrap interface {
		Cause() error
	}
	type HasStackTrace interface {
		StackTrace() errors.StackTrace
	}
	errorType := "error"
	for err != nil {
		if errSt, hasStackTrace := err.(HasStackTrace); hasStackTrace {
			buf.AddStr(CRed).AddChr('\n').AddStr(errorType).AddStr(": ").AddStr(err.Error())
			for _, frame := range errSt.StackTrace() {
				b, _ := frame.MarshalText()
				buf.AddStr("\n  at ").AddBytes(b)
			}
			buf.AddStr(CReset)
		}
		if errUn, hasUnwrap := err.(HasUnwrap); hasUnwrap {
			err = errUn.Cause()
			errorType = "cause"
		} else {
			break
		}
	}
}

func needsQuoting(s string) bool {
	if len(s) == 0 {
		return true
	}
	for _, r := range s {
		if unicode.IsSpace(r) || r == '"' || r == '=' || !unicode.IsPrint(r) {
			return true
		}
	}
	return false
}
