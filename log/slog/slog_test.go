package slog

import (
	"reflect"
	"testing"
)

func TestContext(t *testing.T) {
	logger := &Logger{}
	logger2 := logger.With("msg1", "hello")
	logger3 := logger2.With("msg2", "world")
	logger4 := logger3.(*Logger)

	if have, want := len(logger4.lctx), 4; have != want {
		t.Errorf("length: have: %v, want: %v", have, want)
	}
}

func TestExtractMsg(t *testing.T) {
	logs := []any{"foo", "bar", "msg", "wow", "hello", "world"}
	msg, msglessLogs := extractMsg(logs)

	if have, want := msg, "wow"; have != want {
		t.Errorf("msg: have: %v, want: %v", have, want)
	}

	if have, want := msglessLogs, []any{"foo", "bar", "hello", "world"}; !reflect.DeepEqual(have, want) {
		t.Errorf("msg: have: %v, want: %v", have, want)
	}
}
