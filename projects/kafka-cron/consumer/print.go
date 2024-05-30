package main

import (
	"log"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

const (
	red     = "\x1b[31m"
	green   = "\x1b[32m"
	yellow  = "\x1b[33m"
	blue    = "\x1b[34m"
	magenta = "\x1b[35m"
	cyan    = "\x1b[36m"
	reset   = "\x1b[0m"
)

func PrintConsumedRecord(cj *kgo.Record) {
	log.Printf(
		"%sConsumed Record%s:\n\t%sTopic%s: %s%s%s\n\t%sPartition%s: %s%d%s\n\t%sOffset%s: %s%d%s\n\t%sTimestamp%s: %s%v%s\n\t%sKey%s: %s%s%s\n\t%sValue%s: %s%s%s\n",
		yellow, reset,
		blue, reset, red, cj.Topic, reset,
		blue, reset, magenta, cj.Partition, reset,
		blue, reset, magenta, cj.Offset, reset,
		blue, reset, magenta, cj.Timestamp, reset,
		blue, reset, green, string(cj.Key), reset,
		blue, reset, green, string(cj.Value), reset,
	)
}

func PrintCommandOutput(output []byte, executionDuration float64) {
	log.Printf("%sCommand Executed%s:\n\tExecution Duration:%s%v%s\n\tOutput:%s%s%s\n",
		yellow, reset,
		magenta, time.Duration(executionDuration*float64(time.Second)), reset,
		green, string(output), reset)
}

func PrintRetryProducedRecord(cj *kgo.Record) {
	log.Printf(
		"%sProduced Retry Record%s:\n\t%sTopic%s: %s%s%s\n\t%sPartition%s: %s%d%s\n\t%sOffset%s: %s%d%s\n\t%sTimestamp%s: %s%v%s\n\t%sKey%s: %s%s%s\n\t%sValue%s: %s%s%s\n",
		yellow, reset,
		blue, reset, red, cj.Topic, reset,
		blue, reset, magenta, cj.Partition, reset,
		blue, reset, magenta, cj.Offset, reset,
		blue, reset, magenta, cj.Timestamp, reset,
		blue, reset, green, string(cj.Key), reset,
		blue, reset, green, string(cj.Value), reset,
	)
}
