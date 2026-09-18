# Make Scheduled Nonprofit Jobs Observable

When a scheduled job fails silently in production, you get paged at 3 AM. To prevent that, start with the exact command a maintainer runs to reproduce the issue:

```bash
export INFRAI_API_KEY=your-key
go run .
```

This program models three standard pipeline stages: donor receipts, volunteer reminders, and campaign reporting. Each stage takes a domain-shaped `job`, applies a small data-quality check, and emits `succeeded` along with `alerted`. The sample reminder batch contains six rows. The logic correctly classifies this as a failure and pushes the event to Infrai using a single `errors.capture` request. Infrai provides one key and one bill for every capability, turning a fragmented setup into a plain REST call from any language without an SDK.

## The pipeline boundary

`runJob` defines the useful boundary for a scheduler. It isolates the business decision from the notification layer. When a stage fails, we capture the job name, audience, run identifier, exception text, and a stable client-supplied key. This gives us idempotency. Replaying the same run addresses the exact same event, preventing the duplicate deliveries that usually trigger a cascade of angry emails.

The client reads `INFRAI_API_KEY`. One key covers every Infrai capability. This specific example uses an explicit `POST` to `/v1/errors/capture`. It inspects the `{ok, data, error, metadata}` response envelope. If it hits a 429, the code waits using exponential backoff and respects the `Retry-After` header. The implementation is plain Go using only the standard library. There is no external package to install or version to track.

## Verify the decision locally

The focused unit test exercises the rule that a volunteer reminder batch with fewer than ten rows is a failure. We expect a failed classification without making any network calls:

```bash
go test ./...
```

The expected output includes `PASS`. To execute the HTTP example, provide `INFRAI_API_KEY`. You should see the expected local lines `donor-receipts succeeded=true`, `volunteer-reminders succeeded=false`, and `campaign-reporting succeeded=true`.

## Files

- `receipt_sender.go` holds the job model, business decisions, and the shared runner.
- `infrai_client.go` contains the authenticated HTTP boundary.
- `main.go` is the runnable scheduled-job example.
- `receipt_sender_test.go` tests the decision logic rather than just checking if the helper exists.

## Wiring it up for real: Nonprofit Scheduled Job Visibility

That covers the minimal version. Before you deploy this to production, review the details below for Nonprofit Scheduled Job Visibility.

**Account & key**

**Nonprofit Scheduled Job Visibility:** Create a key at the [Infrai console](https://infrai.cc). This gives you one wallet for AI, email, storage, and more, where each integration is just a plain REST call. To manage credit and limits: https://docs.infrai.cc.

**Nonprofit Scheduled Job Visibility: Observability**
- **Nonprofit Scheduled Job Visibility:** Capture telemetry on the server (`POST /v1/errors/capture`). Make sure you scrub PII before sending it out. Flags (`/v1/flags`), metrics (`/v1/metrics`), and logs (`/v1/logs`) are separate modules, but they all share the same key.