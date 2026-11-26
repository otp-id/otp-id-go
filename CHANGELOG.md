# Changelog

All notable changes to this project are documented in this file. The format
follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the
project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

## [0.2.0] - 2026-11-26

### Added

- `Failure` struct (`Code`, `Message`) and `FailureCode*` constants
  (`NUMBER_NOT_ON_WHATSAPP`, `TOO_FREQUENT`, `CHANNEL_UNAVAILABLE`,
  `PROVIDER_UNAVAILABLE`, `DELIVERY_FAILED`), exposed as `Failure` on both
  `OrderResult` and `StatusResult`. Set only when `Status == "failed"`.
- `UPGRADING.md` with a v0.1.x to v0.2.0 migration guide.

### Changed

- `StatusResult.Verification` (`GET /v3/otp/{otp_id}`) is now also
  populated for pending, not-expired `whatsapp_inbound` transactions
  (previously misscall only), so polling UIs can recover the `wa.me` link.
- Billing for `whatsapp_inbound` (server-side change, documented in the SDK):
  transactions are now charged only when they become `verified`. They stay
  `pending` while waiting for the user's WhatsApp message and can end as
  `expired` (TTL passed, not charged). `OrderResult.Price` is the amount that
  will be charged on verification, and `LastBalance` is unchanged at request
  time. A message sent from a different number than the request's
  `Destination` is rejected and the transaction stays `pending`. Doc-only in
  the SDK; `Status` stays a plain `string`.

## [0.1.1] - 2026-08-17

### Added

- `examples/` — runnable examples for every channel (`whatsapp`, `sms`,
  `voice`, `email`, `misscall`, `whatsapp-inbound`, `send`), mirroring the
  cURL examples in the API docs.

### Changed

- Language floor lowered to **Go 1.15** (`any` → `interface{}`, `io.ReadAll`
  → `ioutil.ReadAll`) so merchants on older toolchains can `go get` the
  module. CI matrix now includes 1.15, 1.21, and stable.
- README quickstart now sets `Brand` in `OrderParams`.

## [0.1.0] - 2026-08-14

### Added

- `Client` with `RequestOTP`, `SendOTP`, `VerifyOTP`, `OTPStatus`,
  `Account`, and `CreateTopup` covering the full OTP.ID V3 API.
- `VerifyWebhookSignature` and `ParseVerifiedEvent` for the `otp.verified`
  webhook (HMAC-SHA256, constant-time comparison, ±5 minute replay window).
- Typed `APIError` with the complete V3 error code set.
- Zero-dependency implementation (standard library only), Go 1.21+.

[Unreleased]: https://github.com/otp-id/otp-id-go/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/otp-id/otp-id-go/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/otp-id/otp-id-go/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/otp-id/otp-id-go/releases/tag/v0.1.0
