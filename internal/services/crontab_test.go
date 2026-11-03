package services

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCrontabServiceParse(t *testing.T) {
	service := NewCrontabService("../../crontab")
	job, err := service.Parse()
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if want := "* * * * *"; job.Spec != want {
		t.Errorf("Spec = %q, want %q", job.Spec, want)
	}
	if want := "php artisan schedule:run >> /dev/null 2>&1"; job.Command != want {
		t.Errorf("Command = %q, want %q", job.Command, want)
	}
}

func TestCrontabServiceParse_TrailingNewline(t *testing.T) {
	path := filepath.Join(t.TempDir(), "crontab")
	if err := os.WriteFile(path, []byte("* * * * * echo hi\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	job, err := NewCrontabService(path).Parse()
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if want := "echo hi"; job.Command != want {
		t.Errorf("Command = %q, want %q", job.Command, want)
	}
}

func TestCrontabServiceParse_TooFewFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "crontab")
	if err := os.WriteFile(path, []byte("* * * * *"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := NewCrontabService(path).Parse(); err == nil {
		t.Fatal("expected error for line without command, got nil")
	}
}
