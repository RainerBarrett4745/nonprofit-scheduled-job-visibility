package main

import (
	"fmt"
	"time"
)

func main() {
	client, err := newInfraiClient()
	if err != nil {
		fmt.Println(err)
		return
	}
	runID := time.Now().UTC().Format("20060102T150405Z")
	jobs := []struct {
		definition job
		process    func(job) error
	}{
		{job{Name: "donor-receipts", Audience: "donors", Records: 250}, receiptJob},
		{job{Name: "volunteer-reminders", Audience: "volunteers", Records: 6}, reminderJob},
		{job{Name: "campaign-reporting", Audience: "campaign-team", Records: 120}, campaignReportJob},
	}
	for _, scheduled := range jobs {
		result := runJob(client, scheduled.definition, runID, scheduled.process)
		fmt.Printf("%s succeeded=%t alerted=%t\n", result.Job, result.Succeeded, result.Alerted)
	}
}

// The copied call shape is infrai.errors.capture: POST /v1/errors/capture.
