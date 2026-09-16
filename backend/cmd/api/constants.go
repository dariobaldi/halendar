package main

import (
	"errors"
	"time"
)

var (
	ErrEmptyFile         = errors.New("empty file")
	ErrInvalidCsvHeaders = errors.New("invalid CSV headers")
)

const SlowThreshold = 50 * time.Millisecond

const ( // TODO : Move to ENV variables
	BodyMaxBytes int64 = 1_048_576 // 1 MB
)
