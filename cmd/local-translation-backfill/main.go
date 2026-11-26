package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

func main() {
	var options publishOptions
	flag.StringVar(&options.endpoint, "endpoint-url", "", "HTTP loopback DynamoDB Local endpoint")
	flag.StringVar(&options.table, "table", "", "local DynamoDB table (required)")
	flag.StringVar(&options.directory, "translations-dir", "", "reviewed translation sidecars (required)")
	flag.BoolVar(&options.apply, "apply", false, "conditionally add ContentEn; default is preview")
	flag.BoolVar(&options.dryRun, "dry-run", false, "preview without writing")
	flag.Parse()
	if flag.NArg() != 0 || (options.apply && options.dryRun) {
		fmt.Fprintln(os.Stderr, "choose only one of --apply or --dry-run; no positional arguments")
		os.Exit(2)
	}

	report, err := publishReviewedEnglish(context.Background(), options)
	if err == nil || len(report.Eligible)+len(report.Updated)+len(report.Unresolved) != 0 {
		if encodeErr := json.NewEncoder(os.Stdout).Encode(report); encodeErr != nil {
			fmt.Fprintln(os.Stderr, encodeErr)
			os.Exit(1)
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
