package chatmix

import (
	"fmt"
	"strings"

	"github.com/jfreymuth/pulse"
	"github.com/jfreymuth/pulse/proto"
)

type VirtualSink struct {
	*pulse.Sink
	moduleID uint32
}

func (v *VirtualSink) Destroy() error {
	pactl, err := pulse.NewClient()
	if err != nil {
		return err
	}
	defer pactl.Close()
	return pactl.RawRequest(&proto.UnloadModule{ModuleIndex: v.moduleID}, nil)
}

func (v *VirtualSink) SetVolume(percentage float64) error {
	pactl, err := pulse.NewClient()
	if err != nil {
		return err
	}
	defer pactl.Close()

	channelVolumes := make([]uint32, len(v.Channels()))
	for i := range channelVolumes {
		channelVolumes[i] = uint32((65536 / 100) * percentage)
	}

	return pactl.RawRequest(&proto.SetSinkVolume{
		SinkIndex:      v.SinkIndex(),
		ChannelVolumes: proto.ChannelVolumes(channelVolumes),
	}, nil)
}

func CreateVirtualSink(originalSink *pulse.Sink, name string) (*VirtualSink, error) {
	pactl, err := pulse.NewClient()
	if err != nil {
		return nil, err
	}
	defer pactl.Close()

	sinkId := "gg-chatmix-" + strings.ToLower(strings.ReplaceAll(name, " ", "_"))
	sinkName := "GG ChatMix - " + name

	var chatMixModule proto.LoadModuleReply
	if err := pactl.RawRequest(&proto.LoadModule{
		Name: "module-combine-sink",
		Args: fmt.Sprintf("sink_name=%s slaves=%s sink_properties='device.description=\"%s\"'", sinkId, originalSink.ID(), sinkName),
	}, &chatMixModule); err != nil {
		return nil, err
	}

	chatMixSink, err := pactl.SinkByID(sinkId)
	if err != nil {
		pactl.RawRequest(&proto.UnloadModule{ModuleIndex: chatMixModule.ModuleIndex}, nil)
		return nil, err
	}

	virtualSink := &VirtualSink{
		moduleID: chatMixModule.ModuleIndex,
		Sink:     chatMixSink,
	}
	if err := virtualSink.SetVolume(100); err != nil {
		virtualSink.Destroy()
		return nil, err
	}

	return virtualSink, nil
}
