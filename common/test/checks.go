package test

import (
	"errors"
	"testing"
)

func NoError(t *testing.T, err error, message ...string) {
	t.Helper()
	if err != nil {
		t.Error(err, message)
	}
}

func HasError(t *testing.T, err error, message ...string) {
	t.Helper()
	if err == nil {
		t.Error(err, message)
	}
}

func ErrorIs(t *testing.T, err, target error, msg ...string) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Fatal(msg)
	}
}

func ErrorIsF(t *testing.T, err, target error, format string, msg ...string) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Fatalf(format, msg)
	}
}

func ErrorIsNot(t *testing.T, err, target error, msg ...string) {
	t.Helper()
	if errors.Is(err, target) {
		t.Fatal(msg)
	}
}

func ErrorIsNotf(t *testing.T, err, target error, format string, msg ...string) {
	t.Helper()
	if errors.Is(err, target) {
		t.Fatalf(format, msg)
	}
}
