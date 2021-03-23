package errors

import (
	"errors"
	stdlib_errors "errors"
	"fmt"
	"io"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		err  string
		want error
	}{
		{"", fmt.Errorf("")},
		{"foo", fmt.Errorf("foo")},
		{"foo", New("foo")},
		{"string with format specifiers: %v", errors.New("string with format specifiers: %v")},
	}

	for _, tt := range tests {
		got := New(tt.err)
		if got.Error() != tt.want.Error() {
			t.Errorf("New.Error(): got: %q, want %q", got, tt.want)
		}
	}
}

func TestWrapNil(t *testing.T) {
	got := Wrap(nil, "no error")
	if got != nil {
		t.Errorf("Wrap(nil, \"no error\"): got %#v, expected nil", got)
	}
}

func TestWrap(t *testing.T) {
	tests := []struct {
		err     error
		message string
		want    string
	}{
		{io.EOF, "read error", "read error: EOF"},
		{Wrap(io.EOF, "read error"), "client error", "client error: read error: EOF"},
	}

	for _, tt := range tests {
		got := Wrap(tt.err, tt.message).Error()
		if got != tt.want {
			t.Errorf("Wrap(%v, %q): got: %v, want %v", tt.err, tt.message, got, tt.want)
		}
	}
}

type nilError struct{}

func (nilError) Error() string { return "nil error" }

func TestCause(t *testing.T) {
	x := New("error")
	tests := []struct {
		err  error
		want error
	}{{
		// nil error is nil
		err:  nil,
		want: nil,
	}, {
		// explicit nil error is nil
		err:  (error)(nil),
		want: nil,
	}, {
		// typed nil is nil
		err:  (*nilError)(nil),
		want: (*nilError)(nil),
	}, {
		// uncaused error is unaffected
		err:  io.EOF,
		want: io.EOF,
	}, {
		// caused error returns cause
		err:  Wrap(io.EOF, "ignored"),
		want: io.EOF,
	}, {
		err:  x, // return from errors.New
		want: x,
	}, {
		WithMessage(nil, "whoops"),
		nil,
	}, {
		WithMessage(io.EOF, "whoops"),
		io.EOF,
	}, {
		WithStack(nil),
		nil,
	}, {
		WithStack(io.EOF),
		io.EOF,
	}}

	for i, tt := range tests {
		got := Cause(tt.err)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("test %d: got %#v, want %#v", i+1, got, tt.want)
		}
	}
}

func TestWithStackNil(t *testing.T) {
	got := WithStack(nil)
	if got != nil {
		t.Errorf("WithStack(nil): got %#v, expected nil", got)
	}
}

func TestWithStack(t *testing.T) {
	tests := []struct {
		err  error
		want string
	}{
		{io.EOF, "EOF"},
		{WithStack(io.EOF), "EOF"},
	}

	for _, tt := range tests {
		got := WithStack(tt.err).Error()
		if got != tt.want {
			t.Errorf("WithStack(%v): got: %v, want %v", tt.err, got, tt.want)
		}
	}
}

func TestWithMessageNil(t *testing.T) {
	got := WithMessage(nil, "no error")
	if got != nil {
		t.Errorf("WithMessage(nil, \"no error\"): got %#v, expected nil", got)
	}
}

func TestWithMessage(t *testing.T) {
	tests := []struct {
		err     error
		message string
		want    string
	}{
		{io.EOF, "read error", "read error: EOF"},
		{WithMessage(io.EOF, "read error"), "client error", "client error: read error: EOF"},
	}

	for _, tt := range tests {
		got := WithMessage(tt.err, tt.message).Error()
		if got != tt.want {
			t.Errorf("WithMessage(%v, %q): got: %q, want %q", tt.err, tt.message, got, tt.want)
		}
	}
}

// errors.New, etc values are not expected to be compared by value
// but the change in errors#27 made them incomparable. Assert that
// various kinds of errors have a functional equality operator, even
// if the result of that equality is always false.
func TestErrorEquality(t *testing.T) {
	vals := []error{
		nil,
		io.EOF,
		errors.New("EOF"),
		New("EOF"),
		Wrap(io.EOF, "EOF"),
		WithMessage(nil, "whoops"),
		WithMessage(io.EOF, "whoops"),
		WithStack(io.EOF),
		WithStack(nil),
	}

	for i := range vals {
		for j := range vals {
			_ = vals[i] == vals[j] // mustn't panic
		}
	}
}

func TestIs(t *testing.T) {
	sentinelError := stdlib_errors.New("sentinel error")
	wrap := Wrap(sentinelError, "wrap error")
	if !Is(wrap, sentinelError) {
		t.Errorf("Expected that '%v' error and the '%v' error should be of the same type", sentinelError, wrap)
	}
}

func TestAs(t *testing.T) {
	type myError struct {
		error
	}
	err := &myError{stdlib_errors.New("error")}
	wrap := Wrap(err, "wrap error")

	var tt *myError
	if !As(wrap, &tt) {
		t.Errorf("Expected that '%v' error and the '%v' error should be of the same type", err, wrap)
	}
}

func TestWithFields(t *testing.T) {
	tests := []struct {
		err  error
		want string
	}{
		{New("EOF", Int("a", 1), String("b", "hello")), "EOF"},
		{WithMessage(New("EOF", Int("a", 1), String("b", "hello")), "yo", Bool("c", true), Float64("d", 1.1)), "yo: EOF"},
	}

	for _, tt := range tests {
		got := tt.err.Error()
		if got != tt.want {
			t.Errorf("WithFields: got: %v, want %v", got, tt.want)
		}
	}
}

func TestStorytime(t *testing.T) {
	storytime1 := []field{Int("a", 1), String("b", "hello")}
	storytime2 := []field{Bool("c", true), Float64("a", 1.1)}

	tests := []struct {
		err  error
		want int
	}{
		{err: io.EOF, want: 0},
		{err: New("EOF", storytime1...), want: 2},
		{err: Wrap(New("EOF", storytime1...), "wrapped", storytime2...), want: 3},
	}

	for _, tt := range tests {
		got := len(Storytime(tt.err))
		if got != tt.want {
			t.Errorf("Storytime(%v): got: %v, want %v", tt.err, got, tt.want)
		}
	}
}
