package controller_test

/*
Copyright (C) 2017 Jan Kasiak

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	"fmt"
	"github.com/cybojanek/gozwave/controller"
	"github.com/cybojanek/gozwave/packet"
)

func Example() {
	// Create controller
	con := controller.SerialController{DevicePath: "/dev/tty.usbmodem1451",
		DebugLogging: true}

	// Optionally register a callback channel for receiving all asynchronous
	// messages, i.e. switch reports. This can be done at any time, even after
	// open. To unregister, call SetCallbackChannel with nil.
	callbackChannel := make(chan *packet.Packet, 1)
	go func() {
		for {
			packet := <-callbackChannel
			fmt.Printf("Got packet: %v\n", packet)
		}
	}()
	con.SetCallbackChannel(callbackChannel)

	// Open controller
	if err := con.Open(); err != nil {
		fmt.Printf("Failed to open controller: %v", err)
		return
	}

	// Create a request
	requestPacket := packet.Packet{Preamble: packet.PacketPreambleSOF,
		PacketType: packet.PacketTypeRequest, MessageType: 0x07}

	// Issue and wait for response
	responsePacket, err := con.DoRequest(&requestPacket)
	if err != nil {
		fmt.Printf("Failed to process request: %v", err)
	} else {
		fmt.Printf("Got reponse packet: %v\n", responsePacket)
	}

	// Do more work ... from many go routines

	// Close controller
	if err := con.Close(); err != nil {
		fmt.Printf("Failed to open controller: %v", err)
		return
	}
}
