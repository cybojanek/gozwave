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

// GetVersionRequest creates a GetVersion request packet
func GetVersionRequest() *packet.Packet {
	p := packet.Packet{Preamble: packet.PacketPreambleSOF,
		PacketType:  packet.PacketTypeRequest,
		MessageType: MessageTypeGetVersion}

	if err := p.Update(); err != nil {
		panic(fmt.Sprintf("This should never fail: %v", err))
	}

	return &p
}

// MemoryGetIDRequest creates a MemoryGetID request packet
func MemoryGetIDRequest() *packet.Packet {
	p := packet.Packet{Preamble: packet.PacketPreambleSOF,
		PacketType:  packet.PacketTypeRequest,
		MessageType: MessageTypeMemoryGetID}

	if err := p.Update(); err != nil {
		panic(fmt.Sprintf("This should never fail: %v", err))
	}

	return &p
}

// SerialAPIGetInitDataRequest creates a SerialAPIGetInitData request packet
func SerialAPIGetInitDataRequest() *packet.Packet {
	p := packet.Packet{Preamble: packet.PacketPreambleSOF,
		PacketType:  packet.PacketTypeRequest,
		MessageType: MessageTypeSerialAPIGetInitData}

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

// ZWGetNodeProtocolInfoRequest creates a ZWGetNodeProtocolInfo
// request packet
func ZWGetNodeProtocolInfoRequest(nodeID uint8) (*packet.Packet, error) {
	if !IsValidNodeID(nodeID) {
		return nil, fmt.Errorf("Invalid nodeID: 0x%02x", nodeID)
	}

	p := packet.Packet{Preamble: packet.PacketPreambleSOF,
		PacketType:  packet.PacketTypeRequest,
		MessageType: MessageTypeZWGetNodeProtocolInfo,
		Body:        []uint8{nodeID}}

	if err := p.Update(); err != nil {
		panic(fmt.Sprintf("This should never fail: %v", err))
	}

	return &p, nil
}

// ZWRequestNodeInfoRequest creates a MessageTypeZWRequestNodeInfo
// request packet
func ZWRequestNodeInfoRequest(nodeID uint8) (*packet.Packet, error) {
	if !IsValidNodeID(nodeID) {
		return nil, fmt.Errorf("Invalid nodeID: 0x%02x", nodeID)
	}

	p := packet.Packet{Preamble: packet.PacketPreambleSOF,
		PacketType:  packet.PacketTypeRequest,
		MessageType: MessageTypeZWRequestNodeInfo,
		Body:        []uint8{nodeID}}

	if err := p.Update(); err != nil {
		panic(fmt.Sprintf("This should never fail: %v", err))
	}

	return &p, nil
}

// ZWSendDataRequest creates a ZWSendData request packet
func ZWSendDataRequest(nodeID uint8, commandClass uint8, payload []uint8,
	transmitOptions uint8, callbackID uint8) (*packet.Packet, error) {

	if !IsValidNodeID(nodeID) {
		return nil, fmt.Errorf("Invalid nodeID: 0x%02x", nodeID)
	}

	p := packet.Packet{}
	p.Preamble = packet.PacketPreambleSOF
	p.PacketType = packet.PacketTypeRequest
	p.MessageType = MessageTypeZWSendData

	// Body: | NODE_ID | LENGTH_OF_PAYLOAD + 1 | COMMAND_CLASS |
	//       | PAYLOAD | TRANSMIT_OPTIONS | CALLBACK_ID |
	data := []uint8{nodeID, 1 + uint8(len(payload)), commandClass}
	data = append(data, payload...)
	data = append(data, transmitOptions)
	// TODO: is this accurate?
	if callbackID != 0 {
		data = append(data, callbackID)
	}
	p.Body = data

	if err := p.Update(); err != nil {
		return nil, err
	}

	return &p, nil
}
