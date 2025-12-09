package services

import (
	"testing"
)

func TestCrontabServiceParse(t *testing.T) {
	service := NewCrontabService("../../crontab")
	job, err := service.Parse()
	if err != nil {
		t.Errorf(`Parse %v`, err)
	}

	expectedSpec := "* * * * *"
	expectedCommand := "php artisan schedule:run >> /dev/null 2>&1"

	if job.Command != expectedCommand {
		t.Errorf(`Parse %v`, job.Command)
	}

	if job.Spec != expectedSpec {
		t.Errorf(`Parse %v`, job.Spec)
	}
}
