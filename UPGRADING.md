# Upgrading

## From v0.1.x to v0.2.0

```bash
go get github.com/otp-id/otp-id-go@v0.2.0
```

**No breaking changes.** v0.2.0 only adds things: new fields on existing
result structs, new constants, and documentation of server behavior. Every
exported identifier from v0.1.x still exists with the same signature, so
existing code compiles and behaves the same. The Go floor is still 1.15.

There is one server-side billing change for `whatsapp_inbound` (described
below). It needs no code change, but if you poll `OTPStatus` you should
handle the `expired` status.

### Delivery failure reason: `Failure`

`OrderResult` and `StatusResult` now have a `Failure *Failure` field. It is
set only when `Status == "failed"` (OTP.ID accepted the request, then
delivery failed) and is `nil` otherwise. A failed delivery is still **not** an
`*APIError`: the call returns `(result, nil)`.

```go
res, err := client.RequestOTP(ctx, otpid.OrderParams{
	Channel:     otpid.ChannelWhatsApp,
	Destination: "6281234567890",
})
if err != nil {
	return err // *otpid.APIError, network error, ...
}
if res.Status == "failed" && res.Failure != nil {
	switch res.Failure.Code {
	case otpid.FailureCodeNumberNotOnWhatsApp:
		// offer SMS instead
	case otpid.FailureCodeTooFrequent:
		// ask the user to wait before retrying
	case otpid.FailureCodeChannelUnavailable, otpid.FailureCodeProviderUnavailable:
		// temporary, try another channel or retry later
	default: // otpid.FailureCodeDeliveryFailed or a code added later
		log.Printf("delivery failed: %s (%s)", res.Failure.Code, res.Failure.Message)
	}
}
```

`Failure.Message` is a human-readable Indonesian description meant for logs.
Branch on `Failure.Code`, and treat codes you don't recognize as a generic
failure, because the server may add new ones.

### `OTPStatus` returns `Verification` for pending WhatsApp Inbound

`StatusResult.Verification` used to be set only for misscall. It is now also
set for `whatsapp_inbound` transactions that are `pending` and not expired,
with `WaNumber`, `Message`, `WaLink` and `ExpiresAt` filled in. A polling UI
that reloads (or a second device) can recover the `wa.me` link from the status
call instead of storing the original `RequestOTP` response:

```go
st, err := client.OTPStatus(ctx, otpID)
if err != nil {
	return err
}
if st.Status == "pending" && st.Verification != nil && st.Verification.WaLink != "" {
	showLink(st.Verification.WaLink) // valid until st.Verification.ExpiresAt
}
```

### WhatsApp Inbound billing: pay only when verified

This is a server-side change, so it applies to every SDK version:

- A `whatsapp_inbound` transaction is charged **only when it becomes
  `verified`**.
- It stays `pending` while waiting for the user's WhatsApp message.
- If `ExpiresAt` passes without a match, it ends as `expired`. Expired
  transactions are **not charged**, and `expired` is terminal.
- `OrderResult.Price` is the amount that will be charged on verification.
  `OrderResult.LastBalance` is unchanged at request time.
- If you set `Destination`, a message from a different number is rejected and
  the transaction stays `pending`.

If your polling loop only stopped on `verified` or `failed`, it would keep
polling an expired transaction. Stop on `expired` too:

```go
func waitForInbound(ctx context.Context, client *otpid.Client, otpID string) (*otpid.StatusResult, error) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for {
		st, err := client.OTPStatus(ctx, otpID)
		if err != nil {
			return nil, err
		}
		switch st.Status {
		case "verified":
			return st, nil // charged st.Price
		case "expired":
			return st, nil // TTL passed without a match, not charged
		case "failed":
			return st, nil // st.Failure explains why
		}
		// still "pending": keep waiting
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}
```

The `otp.verified` webhook (`ParseVerifiedEvent`) is unchanged and is still
the recommended way to detect completion. Polling is the fallback.
