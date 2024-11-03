package internal_test

import (
	"clean-duplicates/internal"
	"os"
	"testing"
)

func TestNtfy(t *testing.T) {
	os.Setenv("NTFY_URL", "http://192.168.1.95:8888")
	os.Setenv("NTFY_TOPIC", "test")
	internal.Ntfy("test", "test msg")
}
