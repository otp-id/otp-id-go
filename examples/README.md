# Examples

Runnable examples for every OTP.ID channel, mirroring the cURL examples in
the [API docs](https://docs.otp.id). Each example is a self-contained
`main` package: it sends an OTP and completes the matching verification
flow for that channel.

## Setup

Every example reads your credentials and target from environment variables:

| Variable | Required | Meaning |
| --- | --- | --- |
| `OTPID_API_KEY` | always | Your merchant API key (`Bearer` token) |
| `OTPID_DESTINATION` | all except `whatsapp-inbound` | Destination phone number (digits only, e.g. `6281234567890`) or email address for the `email` example |

## Run

```bash
export OTPID_API_KEY=your_api_key
export OTPID_DESTINATION=6281234567890

go run ./examples/whatsapp          # server-generated code via WhatsApp, then verify
go run ./examples/sms               # server-generated code via SMS, then verify
go run ./examples/voice             # spoken 4-digit code via phone call, then verify
OTPID_DESTINATION=user@example.com \
go run ./examples/email             # server-generated code via email, then verify
go run ./examples/misscall          # missed call — user completes the caller's number
go run ./examples/whatsapp-inbound  # user sends a WhatsApp message; polls status until verified
go run ./examples/send              # bring your own code (SendOTP) via SMS, then verify
```

The code-based examples (`whatsapp`, `sms`, `voice`, `email`, `misscall`,
`send`) prompt on stdin for the code the user received and call
`VerifyOTP`. The `whatsapp-inbound` example has no code at all — it prints
the `wa.me` deep link and polls `OTPStatus` until the transaction becomes
`verified` (or times out).

> Every run creates a real transaction and charges credits on your
> account. Use a number you control.
