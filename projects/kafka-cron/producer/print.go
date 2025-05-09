package main

import (
	"log"

	"github.com/twmb/franz-go/pkg/kgo"
)

func PrintProducedRecord(cj *kgo.Record) {
	red := "\x1b[31m"
	green := "\x1b[32m"
	yellow := "\x1b[33m"
	blue := "\x1b[34m"
	magenta := "\x1b[35m"
	// cyan := "\x1b[36m"
	reset := "\x1b[0m"

	log.Printf(
		"%sProduced Record%s:\n\t%sTopic%s: %s%s%s\n\t%sPartition%s: %s%d%s\n\t%sOffset%s: %s%d%s\n\t%sTimestamp%s: %s%v%s\n\t%sKey%s: %s%s%s\n\t%sValue%s: %s%s%s\n",
		yellow, reset,
		blue, reset, red, cj.Topic, reset,
		blue, reset, magenta, cj.Partition, reset,
		blue, reset, magenta, cj.Offset, reset,
		blue, reset, magenta, cj.Timestamp, reset,
		blue, reset, green, string(cj.Key), reset,
		blue, reset, green, string(cj.Value), reset,
	)
}
