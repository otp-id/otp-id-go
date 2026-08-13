package otpid

// Channel is the OTP delivery channel accepted by the V3 API.
type Channel string

const (
	ChannelWhatsApp        Channel = "whatsapp"
	ChannelSMS             Channel = "sms"
	ChannelVoice           Channel = "voice"
	ChannelEmail           Channel = "email"
	ChannelMisscall        Channel = "misscall"
	ChannelWhatsAppInbound Channel = "whatsapp_inbound"
)
