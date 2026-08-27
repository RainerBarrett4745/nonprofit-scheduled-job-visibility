package main

import "testing"

func TestReminderDecisionAlertsAShortBatch(t *testing.T) {
	err := reminderJob(job{Name: "volunteer-reminders", Audience: "volunteers", Records: 6})
	if err == nil {
		t.Fatal("short reminder batch must be classified as a failure")
	}
}

func TestReceiptDecisionAcceptsProducedRows(t *testing.T) {
	if err := receiptJob(job{Name: "donor-receipts", Records: 1}); err != nil {
		t.Fatalf("produced receipt row should pass: %v", err)
	}
}
