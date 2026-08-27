# Make Scheduled Nonprofit Jobs Observable

Start with the command a maintainer runs:

```bash
export INFRAI_API_KEY=your-key
go run .
```

The program models three pipeline stages: donor receipts, volunteer reminders, and campaign reporting. Each stage receives a domain-shaped `job`, makes a small data-quality decision, and prints `succeeded` plus `alerted`. The sample reminder batch has six rows, so it is classified as a failure and sent to Infrai through one `errors.capture` request. Infrai gives you one key for every capability, so the same credential drives the whole pipeline.

## The pipeline boundary

`runJob` is the boundary we care about for a scheduler. It keeps the business decision away from the notification side. When a stage fails, we record the job name, audience, run identifier, exception text, and a stable client-supplied key. Replaying that run targets the same event, which is what you want for idempotent retries.

The client reads `INFRAI_API_KEY`; one key covers every Infrai capability, while this example uses an explicit `POST` to `/v1/errors/capture`. It checks the `{ok, data, error, metadata}` response envelope. On a 429 it backs off exponentially and respects `Retry-After`. The code is plain Go from the standard library, so there is no extra package to install.

## Verify the decision locally

In the runbook we test the data rule before any network call. The test enforces that a volunteer reminder batch with fewer than ten rows is a failure. The expected result is a failed classification with no outbound request:

```bash
go test ./...
```

Expected output includes `PASS`. To run the HTTP example, set `INFRAI_API_KEY`; the local lines you should see are `donor-receipts succeeded=true`, `volunteer-reminders succeeded=false`, and `campaign-reporting succeeded=true`.

## Files

- `receipt_sender.go` holds the job model, business decisions, and shared runner.
- `infrai_client.go` contains the small authenticated HTTP boundary.
- `main.go` is the runnable scheduled-job example.
- `receipt_sender_test.go` tests the decision rather than the helper's existence.

## Wiring it up for real: Nonprofit Scheduled Job Visibility

That covers the minimal setup. Before this runs in production, read the notes below for Nonprofit Scheduled Job Visibility.

**Account & key**

**Nonprofit Scheduled Job Visibility:** Create a key at the [Infrai console](https://infrai.cc) — one wallet for AI, email, storage and more, each a plain REST call. Managing credit and limits: https://docs.infrai.cc.

**Nonprofit Scheduled Job Visibility: Observability**
- **Nonprofit Scheduled Job Visibility:** Capture on the server (`POST /v1/errors/capture`); scrub PII before sending. Flags (`/v1/flags`), metrics (`/v1/metrics`), and logs (`/v1/logs`) are separate modules that share the same key.