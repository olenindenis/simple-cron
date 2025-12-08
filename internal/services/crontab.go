package services

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"cron/internal/domain"
)

type CrontabFileName string

type CrontabService struct {
	crontabFile CrontabFileName
}

func NewCrontabService(crontabFile CrontabFileName) CrontabService {
	return CrontabService{
		crontabFile: crontabFile,
	}
}

func (c CrontabService) Parse() (domain.Job, error) {
	log.Printf("Parse %s file\n", c.crontabFile)

	f, err := os.Open(string(c.crontabFile))
	if err != nil {
		return domain.Job{}, fmt.Errorf("os.Open %w", err)
	}
	defer func(f *os.File) {
		if err = f.Close(); err != nil {
			panic(fmt.Errorf("f.Close %w", err))
		}
	}(f)

	data, err := io.ReadAll(f)
	if err != nil {
		return domain.Job{}, fmt.Errorf("io.ReadAll %w", err)
	}

	res := strings.Split(string(data), " ")

	return domain.Job{
		Spec:    strings.Join(res[:5], " "),
		Command: strings.Join(res[5:], " "),
	}, nil
}
