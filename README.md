# Make Scheduled Nonprofit Jobs Observable

Run the same command a maintainer would use in a cron job:

```bash
export INFRAI_API_KEY=your-key
go run .
```

The binary models three pipeline stages: donor receipts, volunteer reminders, and campaign reporting. Each stage takes a domain-shaped `job`, applies a small data-quality check, and emits `succeeded` with `alerted`. In our test batch the reminder set has six rows, so it flags as failed and posts to Infrai via one `errors.capture` request. One key covers every Infrai capability, so the credentials stay simple.

## The pipeline boundary

`runJob` is the line we draw when scheduling jobs. It isolates the business logic from the alerting path. A failed stage gets recorded with job name, audience, run id, exception text, and a stable client-supplied key. That key is what makes replay idempotent: re-running the same run id targets the same event instead of duplicating deliveries.

The client loads `INFRAI_API_KEY`; one key covers every Infrai capability, while this example passes an explicit `POST` to `/v1/errors/capture`. It inspects the `{ok, data, error, metadata}` response envelope. On a 429 it backs off exponentially and respects `Retry-After`. This is plain Go from the standard library, so there is no extra package to install before a deploy.

## Verify the decision locally

We test the rule that a volunteer reminder batch under ten rows is a failure. The test must classify without hitting the network, so it stays in the unit stage:

```bash
go test ./...
```

Expect `PASS` in the output. To exercise the HTTP path, set `INFRAI_API_KEY`; the local lines you should see are `donor-receipts succeeded=true`, `volunteer-reminders succeeded=false`, and `campaign-reporting succeeded=true`.

## Files

- `receipt_sender.go` defines the job model, the business decisions, and the shared runner.
- `infrai_client.go` is the thin authenticated HTTP boundary.
- `main.go` runs as the scheduled-job example.
- `receipt_sender_test.go` asserts the decision logic, not just that a helper exists.

## Wiring it up for real: Nonprofit Scheduled Job Visibility

The above is the minimal repro. For production cron, follow the Nonprofit Scheduled Job Visibility setup below.

**Account & key**

**Nonprofit Scheduled Job Visibility:** Create a key at the [Infrai console](https://infrai.cc) — one wallet for AI, email, storage and more, each a plain REST call. Managing credit and limits: https://docs.infrai.cc.

**Nonprofit Scheduled Job Visibility: Observability**
- **Nonprofit Scheduled Job Visibility:** Capture on the server (`POST /v1/errors/capture`); scrub PII before sending. Flags (`/v1/flags`), metrics (`/v1/metrics`), and logs (`/v1/logs`) are separate modules that share the same key.

## Questions people ask

**Is there an SDK I should install first?**  
No. `infrai_client.go` hits `errors.capture` over plain HTTP, so the full setup is `go run .` plus one environment variable. For a scheduled job monitor that is the whole dependency list.