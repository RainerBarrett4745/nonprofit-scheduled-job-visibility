package main

import (
	"errors"
	"fmt"
)

type job struct {
	Name     string
	Audience string
	Records  int
}

type runResult struct {
	Job       string
	Succeeded bool
	Alerted   bool
}

func runJob(client *infraiClient, current job, runID string, process func(job) error) runResult {
	result := runResult{Job: current.Name}
	if err := process(current); err == nil {
		result.Succeeded = true
		return result
	} else {
		result.Alerted = client.captureFailure(current, runID, err) == nil
		return result
	}
}

func receiptJob(current job) error {
	if current.Records == 0 {
		return errors.New("no donor receipts were produced")
	}
	return nil
}

func reminderJob(current job) error {
	if current.Records < 10 {
		return fmt.Errorf("volunteer reminder batch has %d records", current.Records)
	}
	return nil
}

func campaignReportJob(current job) error {
	if current.Records < 100 {
		return fmt.Errorf("campaign report has %d rows", current.Records)
	}
	return nil
}
