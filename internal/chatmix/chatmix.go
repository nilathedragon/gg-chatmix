package chatmix

import "github.com/jfreymuth/pulse"

// Defines the basic capabilities of a compatible DAC with software-based ChatMix functionality
type ChatmixDAC interface {
	// Closes all the open resources
	Close() error
	// Attempt to find the output sink for this DAC
	GetSink() (*pulse.Sink, error)
	// Start a go routine that continuously reads incoming data from the DAC and forwards only ChatMix volumes
	ReadChatMix() chan ChatMixPacket
	// Configure whether the DAC will allow the chatmix balance to be configured
	SetChatmixEnabled(bool) error
	// Control whether the Sonar icon will be shown on the DAC or not
	SetSonarIconEnabled(bool) error
}

// Data received from the DAC when the ChatMix is changed
type ChatMixPacket struct {
	Chat float64
	Game float64
}
