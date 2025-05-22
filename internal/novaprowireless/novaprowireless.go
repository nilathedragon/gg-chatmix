package novaprowireless

import (
	"errors"
	"strings"

	"github.com/google/gousb"
	"github.com/jfreymuth/pulse"
	"github.com/nilathedragon/gg-chatmix/internal/chatmix"
)

const (
	VendorID  = gousb.ID(0x1038)
	ProductID = gousb.ID(0x12E0)

	ConfigNr  = 1
	Interface = 4
	Alt       = 0

	EndpointTX = 4
	EndpointRX = int(0x84)

	// Name of the Sink this DAC creates
	Sink = "SteelSeries_Arctis_Nova_Pro_Wireless"

	// TX/RX mode for packet
	ToStation   = 6
	FromStation = 7

	// Commands
	CommandSonarIcon      = 141
	CommandChatMixControl = 73
	CommandChatMix        = 69
)

type novaProWireless struct {
	ctx    *gousb.Context
	device *gousb.Device
	config *gousb.Config
	inter  *gousb.Interface

	rx *gousb.InEndpoint
	tx *gousb.OutEndpoint
}

func (n *novaProWireless) Close() error {
	n.inter.Close()
	return errors.Join(n.config.Close(), n.device.Close(), n.ctx.Close())
}

func (n *novaProWireless) GetSink() (*pulse.Sink, error) {
	pactl, err := pulse.NewClient()
	if err != nil {
		return nil, err
	}
	defer pactl.Close()

	sinks, err := pactl.ListSinks()
	if err != nil {
		return nil, err
	}

	for _, sink := range sinks {
		if strings.Contains(sink.ID(), Sink) {
			return sink, nil
		}
	}
	return nil, errors.New("sink not found")
}

func (n *novaProWireless) ReadChatMix() chan chatmix.ChatMixPacket {
	c := make(chan chatmix.ChatMixPacket)
	go func() {
		defer close(c)
		for {
			bytes := make([]byte, 64)
			bytesRead, err := n.rx.Read(bytes)
			if err != nil {
				break
			}

			// Packet Size is 64, everything else we cannot parse
			if bytesRead != 64 || bytes[0] != FromStation {
				continue
			}

			// Skip packets that are not chat mix related
			if bytes[1] != CommandChatMix {
				continue
			}

			c <- chatmix.ChatMixPacket{
				Chat: float64(bytes[3]),
				Game: float64(bytes[2]),
			}
		}
	}()
	return c
}

func (n *novaProWireless) SetChatmixEnabled(enable bool) error {
	enableVal := byte(0)
	if enable {
		enableVal = byte(1)
	}
	_, err := n.tx.Write([]byte{ToStation, CommandChatMixControl, enableVal})
	return err
}

func (n *novaProWireless) SetSonarIconEnabled(enable bool) error {
	enableVal := byte(0)
	if enable {
		enableVal = byte(1)
	}
	_, err := n.tx.Write([]byte{ToStation, CommandSonarIcon, enableVal})
	return err
}

func Open() (chatmix.ChatmixDAC, error) {
	ctx := gousb.NewContext()
	device, err := ctx.OpenDeviceWithVIDPID(VendorID, ProductID)
	if err != nil {
		return nil, err
	}
	device.SetAutoDetachLibOnly(true)

	config, err := device.Config(ConfigNr)
	if err != nil {
		return nil, err
	}

	inter, err := config.Interface(Interface, Alt)
	if err != nil {
		return nil, err
	}

	rxEndpoint, err := inter.InEndpoint(EndpointRX)
	if err != nil {
		return nil, err
	}

	txEndpoint, err := inter.OutEndpoint(EndpointTX)
	if err != nil {
		return nil, err
	}

	return &novaProWireless{
		ctx:    ctx,
		device: device,
		config: config,
		inter:  inter,

		rx: rxEndpoint,
		tx: txEndpoint,
	}, nil
}
