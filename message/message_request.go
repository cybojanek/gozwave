package message

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
	"github.com/cybojanek/gozwave/packet"
)

// SerialAPIGetInitDataRequest creates a SerialAPIGetInitData request packet
func SerialAPIGetInitDataRequest() *packet.Packet {
	p := packet.Packet{Preamble: packet.PacketPreambleSOF,
		PacketType:  packet.PacketTypeRequest,
		MessageType: MessageTypeSerialAPIGetInitdata}

	if err := p.Update(); err != nil {
		panic(fmt.Sprintf("This should never fail: %v", err))
	}

	return &p
}

// SerialAPIGetCapabilitiesRequest creates a  SerialAPIGetCapabilities request
// packet
func SerialAPIGetCapabilitiesRequest() *packet.Packet {
	p := packet.Packet{Preamble: packet.PacketPreambleSOF,
		PacketType:  packet.PacketTypeRequest,
		MessageType: MessageTypeSerialAPIGetCapabilities}

	if err := p.Update(); err != nil {
		panic(fmt.Sprintf("This should never fail: %v", err))
	}

	return &p
}

// ZWGetControllerCapabilitiesRequest creates a ZWGetControllerCapabilities
// request packet
func ZWGetControllerCapabilitiesRequest() *packet.Packet {
	p := packet.Packet{Preamble: packet.PacketPreambleSOF,
		PacketType:  packet.PacketTypeRequest,
		MessageType: MessageTypeZWGetControllerCapabilities}

	if err := p.Update(); err != nil {
		panic(fmt.Sprintf("This should never fail: %v", err))
	}

	return &p
}

// ZWSendDataGetRequest creates a ZWSendData Get request packet
func ZWSendDataGetRequest(nodeID uint8, transmitOptions uint8, callbackID uint8) (*packet.Packet, error) {

	message := ZWSendData{}
	message.NodeID = nodeID
	message.CommandClass = CommandClassBasic
	message.Payload = []uint8{BasicCommandGet}
	if err := message.SetTransmitOptionsMask(transmitOptions); err != nil {
		return nil, err
	}
	message.CallbackID = callbackID

	return ZWSendDataToRequestPacket(&message)
}

// ZWSendDataSetRequest creates a ZWSendData Set request packet
func ZWSendDataSetRequest(nodeID uint8, value uint8, transmitOptions uint8, callbackID uint8) (*packet.Packet, error) {

	message := ZWSendData{}
	message.NodeID = nodeID
	message.CommandClass = CommandClassBasic
	message.Payload = []uint8{BasicCommandSet, value}
	if err := message.SetTransmitOptionsMask(transmitOptions); err != nil {
		return nil, err
	}
	message.CallbackID = callbackID

	return ZWSendDataToRequestPacket(&message)
}

// ZWSendDataToRequestPacket creates a ZWSendData request packet
func ZWSendDataToRequestPacket(message *ZWSendData) (*packet.Packet, error) {
	p := packet.Packet{}
	p.Preamble = packet.PacketPreambleSOF
	p.PacketType = packet.PacketTypeRequest
	p.MessageType = MessageTypeZWSendData

	transmitOptions := uint8(0)
	if message.TransmitOptions.ACK {
		transmitOptions |= TransmitOptionACK
	}

	if message.TransmitOptions.LowPower {
		transmitOptions |= TransmitOptionLowPower
	}

	if message.TransmitOptions.AutoRoute {
		transmitOptions |= TransmitOptionAutoRoute
	}

	if message.TransmitOptions.NoRoute {
		transmitOptions |= TransmitOptionNoRoute
	}

	if message.TransmitOptions.Explore {
		transmitOptions |= TransmitOptionExplore
	}

	// Body: | NODE_ID | LENGTH_OF_PAYLOAD  + 1 | COMMAND_CLASS |
	//       | PAYLOAD | TRANSMIT_OPTIONS | CALLBACK_ID |
	data := []uint8{message.NodeID, 1 + uint8(len(message.Payload)),
		message.CommandClass}
	data = append(data, message.Payload...)
	if message.CallbackID != 0 {
		data = append(data, message.CallbackID)
	}
	p.Body = data

	if err := p.Update(); err != nil {
		return nil, err
	}

	return &p, nil
}
