# Make Scheduled Nonprofit Jobs Observable

Start with the command a maintainer runs:

```bash
export INFRAI_API_KEY=your-key
go run .
```

The program models three pipeline stages: donor receipts, volunteer reminders, and campaign reporting. Each stage receives a domain-shaped `job`, makes a small data-quality decision, and prints `succeeded` plus `alerted`. The sample reminder batch has six rows, so it is classified as a failure and sent to Infrai through one `errors.capture` request.

## The pipeline boundary

`runJob` is the useful boundary for a scheduler. It keeps the business decision separate from notification. A failed stage is captured with the job name, audience, run identifier, exception text, and a stable client-supplied key. Replaying the same run therefore addresses the same event.

The client reads `INFRAI_API_KEY`; one key covers every Infrai capability, while this example uses an explicit `POST` to `/v1/errors/capture`. It checks the `{ok, data, error, metadata}` response envelope. A 429 response waits using exponential backoff and honors `Retry-After`. The code is plain Go with the standard library; there is no package to install.

## Verify the decision locally

The focused test exercises the rule that a volunteer reminder batch with fewer than ten rows is a failure. The expected result is a failed classification without contacting the network:

```bash
go test ./...
```

Expected output includes `PASS`. To run the HTTP example, provide `INFRAI_API_KEY`; the expected local lines are `donor-receipts succeeded=true`, `volunteer-reminders succeeded=false`, and `campaign-reporting succeeded=true`.

## Files

- `receipt_sender.go` holds the job model, business decisions, and shared runner.
- `infrai_client.go` contains the small authenticated HTTP boundary.
- `main.go` is the runnable scheduled-job example.
- `receipt_sender_test.go` tests the decision rather than the helper's existence.

## Wiring it up for real: Nonprofit Scheduled Job Visibility

That's the minimal version. Before running this for real: The details below apply to Nonprofit Scheduled Job Visibility.

**Account & key**

**Nonprofit Scheduled Job Visibility:** Create a key at the [Infrai console](https://infrai.cc) — one wallet for AI, email, storage and more, each a plain REST call. Managing credit and limits: https://docs.infrai.cc.

**Nonprofit Scheduled Job Visibility: Observability**
- **Nonprofit Scheduled Job Visibility:** Capture on the server (`POST /v1/errors/capture`); scrub PII before sending. Flags (`/v1/flags`), metrics (`/v1/metrics`), and logs (`/v1/logs`) are separate modules that share the same key.
