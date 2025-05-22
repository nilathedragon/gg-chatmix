package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jfreymuth/pulse"
	"github.com/nilathedragon/gg-chatmix/internal/chatmix"
	"github.com/nilathedragon/gg-chatmix/internal/novaprowireless"
)

func main() {
	pactl, err := pulse.NewClient()
	if err != nil {
		panic(err)
	}
	defer pactl.Close()

	dac, err := novaprowireless.Open()
	if err != nil {
		panic(err)
	}
	defer dac.Close()

	// Make sure we disable all the features again when GG Chatmix exits
	defer dac.SetSonarIconEnabled(false)
	defer dac.SetChatmixEnabled(false)

	sink, err := dac.GetSink()
	if err != nil {
		panic(err)
	}

	chatMixSink, err := chatmix.CreateVirtualSink(sink, "Chat")
	if err != nil {
		panic(err)
	}
	defer chatMixSink.Destroy()

	gameMixSink, err := chatmix.CreateVirtualSink(sink, "Game")
	if err != nil {
		panic(err)
	}
	defer gameMixSink.Destroy()

	// Now that we are ready to mix, we can configure the DAC to allow
	// the user to configure ChatMix via the headset!
	dac.SetSonarIconEnabled(true)
	dac.SetChatmixEnabled(true)

	// Make sure we handle all shutdown signals
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	chatMix := dac.ReadChatMix()

	fmt.Println("GG ChatMix is now running")
	for {
		select {
		case packet := <-chatMix:
			if err := chatMixSink.SetVolume(packet.Chat); err != nil {
				fmt.Println("Error while setting Chat volume ", err)
			}
			if err := gameMixSink.SetVolume(packet.Game); err != nil {
				fmt.Println("Error while setting Game volume ", err)
			}
			break
		case <-sigc:
			fmt.Println("Shutdown signal received, exiting")
			return
		}
	}
}
