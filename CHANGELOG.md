# Changelog

All notable changes to this project are documented in this file. The format
follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the
project adheres to [Semantic Versioning](https://semver.org/).

## [0.1.0] - 2026-08-14

### Added

- `Client` with `RequestOTP`, `SendOTP`, `VerifyOTP`, `OTPStatus`,
  `Account`, and `CreateTopup` covering the full OTP.ID V3 API.
- `VerifyWebhookSignature` and `ParseVerifiedEvent` for the `otp.verified`
  webhook (HMAC-SHA256, constant-time comparison, ±5 minute replay window).
- Typed `APIError` with the complete V3 error code set.
- Zero-dependency implementation (standard library only), Go 1.21+.
