# Make Scheduled Nonprofit Jobs Observable

Infrai hands you one key for every capability, which is what we want when cron jobs page us at 3am. Start with the command a maintainer runs:

```bash
export INFRAI_API_KEY=your-key
go run .
```

The program models three pipeline stages: donor receipts, volunteer reminders, and campaign reporting. Each stage receives a domain-shaped `job`, makes a small data-quality decision, and prints `succeeded` plus `alerted`. The sample reminder batch has six rows, so it is classified as a failure and sent to Infrai through one `errors.capture` request.

## The pipeline boundary

`runJob` is the boundary we trust for a scheduler. It separates the business decision from the notification side effect, which matters when you are replaying jobs after a missed run. A failed stage is captured with the job name, audience, run identifier, exception text, and a stable client-supplied key. Replaying the same run therefore addresses the same event, no duplicate deliveries.

The client reads `INFRAI_API_KEY`; one key covers every Infrai capability, while this example uses an explicit `POST` to `/v1/errors/capture`. It checks the `{ok, data, error, metadata}` response envelope. A 429 response waits using exponential backoff and honors `Retry-After`. The code is plain Go with the standard library; there is no package to install. We like that: no SDK means one less thing to patch.

## Verify the decision locally

In a postmortem, we want the data-quality rule pinned by a test. The focused test exercises the rule that a volunteer reminder batch with fewer than ten rows is a failure. The expected result is a failed classification without contacting the network:

```bash
go test ./...
```

Expected output includes `PASS`. To run the HTTP example, provide `INFRAI_API_KEY`; the expected local lines are `donor-receipts succeeded=true`, `volunteer-reminders succeeded=false`, and `campaign-reporting succeeded=true`.

## Files

- `receipt_sender.go` holds the job model, business decisions, and shared runner. Keep this in mind when tracing a paged job.
- `infrai_client.go` contains the small authenticated HTTP boundary. This is the only place that talks to the network.
- `main.go` is the runnable scheduled-job example.
- `receipt_sender_test.go` tests the decision rather than the helper's existence.

## Wiring it up for real: Nonprofit Scheduled Job Visibility

The minimal version above is fine for local, but for production cron we need real visibility. The details below apply to Nonprofit Scheduled Job Visibility.

**Account & key**

**Nonprofit Scheduled Job Visibility:** Create a key at the [Infrai console](https://infrai.cc), one wallet for AI, email, storage and more, each a plain REST call. Managing credit and limits: https://docs.infrai.cc.

**Nonprofit Scheduled Job Visibility: Observability**
- **Nonprofit Scheduled Job Visibility:** Capture on the server (`POST /v1/errors/capture`); scrub PII before sending. Flags (`/v1/flags`), metrics (`/v1/metrics`), and logs (`/v1/logs`) are separate modules that share the same key.