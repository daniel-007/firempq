package conf

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/op/go-logging"
)

type PQueueConfigData struct {
	DefaultMessageTtl    int64
	DefaultDeliveryDelay int64
	DefaultLockTimeout   int64
	DefaultPopCountLimit int64
	ExpirationBatchSize  int64
	UnlockBatchSize      int64
	MaxPopWaitTimeout    int64
	MaxPopBatchSize      int64
	MaxLockTimeout       int64
	MaxDeliveryTimeout   int64
}

// DSQueueConfigData a config specific to a DSQueue
type DSQueueConfigData struct {
	DefaultMessageTtl    int64
	DefaultDeliveryDelay int64
	DefaultLockTimeout   int64
	DefaultPopCountLimit int64
	ExpirationBatchSize  int64
	UnlockBatchSize      int64
	MaxPopWaitTimeout    int64
	MaxPopBatchSize      int64
	MaxLockTimeout       int64
	MaxDeliveryTimeout   int64
}

// Config is a generic service config type.
type Config struct {
	LogLevel            logging.Level
	Port                int
	Interface           string
	DbFlushInterval     time.Duration
	DbBufferSize        int64
	DatabasePath        string
	PQueueConfig        PQueueConfigData
	DSQueueConfig       DSQueueConfigData
	UpdateInterval      time.Duration
	BinaryLogPath       string
	BinaryLogBufferSize int
	BinaryLogPageSize   uint64
	BinaryLogFrameSize  uint64
}

var CFG *Config
var CFG_PQ *PQueueConfigData
var CFG_DSQ *DSQueueConfigData

func init() {
	NewDefaultConfig()
}

func NewDefaultConfig() *Config {
	cfg := Config{
		LogLevel:            logging.INFO,
		Port:                9033,
		Interface:           "",
		DatabasePath:        "./",
		DbFlushInterval:     100,
		DbBufferSize:        10000,
		BinaryLogPath:       "./",
		BinaryLogBufferSize: 128,
		BinaryLogPageSize:   2 * 1024 * 1024 * 1025, // 2Gb
		PQueueConfig: PQueueConfigData{
			DefaultMessageTtl:    10 * 60 * 1000,
			DefaultDeliveryDelay: 0,
			DefaultLockTimeout:   60 * 1000,
			DefaultPopCountLimit: 0,
			ExpirationBatchSize:  1000,
			UnlockBatchSize:      1000,
			MaxPopWaitTimeout:    30000,
			MaxPopBatchSize:      10,
			MaxLockTimeout:       3600 * 1000,
			MaxDeliveryTimeout:   3600 * 1000 * 12,
		},
		DSQueueConfig: DSQueueConfigData{
			DefaultMessageTtl:    10 * 60 * 1000,
			DefaultDeliveryDelay: 0,
			DefaultLockTimeout:   60 * 1000,
			DefaultPopCountLimit: 0,
			ExpirationBatchSize:  1000,
			UnlockBatchSize:      1000,
			MaxPopWaitTimeout:    30000,
			MaxPopBatchSize:      10,
			MaxLockTimeout:       3600 * 1000,
			MaxDeliveryTimeout:   3600 * 1000 * 12,
		},
	}
	CFG = &cfg
	CFG_PQ = &(cfg.PQueueConfig)
	CFG_DSQ = &(cfg.DSQueueConfig)
	return &cfg
}

func getErrorLine(data []byte, byteOffset int64) (int64, int64, string) {
	var lineNum int64 = 1
	var lineOffset int64
	var lineData []byte
	for idx, b := range data {
		if b < 32 {
			if lineOffset > 0 {
				lineNum++
				lineOffset = 0
				lineData = make([]byte, 0, 32)
			}

		} else {
			lineOffset++
			lineData = append(lineData, b)
		}
		if int64(idx) == byteOffset {
			break
		}
	}
	return lineNum, lineOffset, string(lineData)
}

func formatTypeError(lineNum, lineOffset int64, lineText string, err *json.UnmarshalTypeError) string {
	return fmt.Sprintf(
		"Config error at line %d:%d. Unexpected data type '%s', should be '%s': '%s'",
		lineNum, lineOffset, err.Value, err.Type.String(), strings.TrimSpace(lineText))
}

// ReadConfig reads and decodes firempq_cfg.json file.
func ReadConfig() error {
	confData, err := ioutil.ReadFile("firempq_cfg.json")

	if err != nil {
		return err
	}

	decoder := json.NewDecoder(bytes.NewReader(confData))

	cfg := NewDefaultConfig()
	err = decoder.Decode(cfg)
	if err != nil {
		if e, ok := err.(*json.UnmarshalTypeError); ok {
			num, offset, str := getErrorLine(confData, e.Offset)
			err = errors.New(formatTypeError(num, offset, str, e))
		}
		return err
	}
	CFG = cfg
	CFG_PQ = &(cfg.PQueueConfig)
	CFG_DSQ = &(cfg.DSQueueConfig)
	return nil
}
