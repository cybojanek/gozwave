// ZWave Packet parsing and serialization
package packet

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
	"errors"
	"fmt"
)

// Packet preamble
const (
	PacketPreambleSOF uint8 = 0x01
	PacketPreambleACK       = 0x06
	PacketPreambleNAK       = 0x15
	PacketPreambleCAN       = 0x18
)

// Packet type
const (
	PacketTypeRequest  uint8 = 0x00
	PacketTypeResponse uint8 = 0x01
)

// Packet
type Packet struct {
	Preamble    uint8
	Length      uint8
	PacketType  uint8
	MessageType uint8
	Body        []uint8 // Maximum length is 252 bytes
	Checksum    uint8
}

// A PacketParser instance
type PacketParser struct {
	state  packetParseState
	packet *Packet
}

// Internal PacketParser state
type packetParseState int

const (
	packetParseStatePreamble packetParseState = iota
	packetParseStateLength
	packetParseStatePacketType
	packetParseStateMessageType
	packetParseStateBody
	packetParseStateChecksum
)

// Get String representation of a packet
func (packet *Packet) String() string {
	return fmt.Sprintf("%+v", *packet)
}

// Serialize Packet to []byte representation. Runs Update() before returning
func (packet *Packet) Bytes() ([]byte, error) {
	if err := packet.Update(); err != nil {
		return nil, err
	}

	bytes := []byte{}

	bytes = append(bytes, packet.Preamble)
	if packet.Preamble == PacketPreambleSOF {
		bytes = append(bytes, packet.Length)
		bytes = append(bytes, packet.PacketType)
		bytes = append(bytes, packet.MessageType)
		for _, b := range packet.Body {
			bytes = append(bytes, b)
		}
		bytes = append(bytes, packet.Checksum)
	}

	return bytes, nil
}

// Set the length and checksum of the packet based on the other fields.
// Errors if body is too long.
func (packet *Packet) Update() error {
	// These don't have a length nor checksum
	switch packet.Preamble {
	case PacketPreambleACK, PacketPreambleNAK, PacketPreambleCAN:
		return nil
	}

	if len(packet.Body) > 0xff-3 {
		return errors.New(fmt.Sprintf("Packet Body is too long: %d > %d",
			len(packet.Body), 0xff-3))
	}

	// Minimum length
	packet.Length = 3
	// Add body
	packet.Length += uint8(len(packet.Body))

	// Reset to 0xff
	packet.Checksum = 0xff
	// Preamble is not part of checksum
	packet.Checksum ^= packet.Length
	packet.Checksum ^= packet.PacketType
	packet.Checksum ^= packet.MessageType
	for _, x := range packet.Body {
		packet.Checksum ^= x
	}

	return nil
}

// Parse a byte. If finished parsing a Packet, then Packet return is non nil.
// Resets internal state on error and should eventually again return a valid
// packet.
func (parser *PacketParser) Parse(b uint8) (*Packet, error) {
	var p *Packet
	var e error

	switch parser.state {

	case packetParseStatePreamble:
		switch b {

		case PacketPreambleACK, PacketPreambleNAK, PacketPreambleCAN:
			return &Packet{Preamble: b}, nil

		case PacketPreambleSOF:
			parser.packet = &Packet{Preamble: b}
			parser.state = packetParseStateLength

		default:
			e = errors.New(fmt.Sprintf("Bad preamble: %d", b))
			goto reset
		}

	case packetParseStateLength:
		if b < 3 {
			e = errors.New(fmt.Sprintf("Bad length: %d", b))
			goto reset
		}
		parser.packet.Length = b
		parser.state = packetParseStatePacketType

	case packetParseStatePacketType:
		if b != PacketTypeRequest && b != PacketTypeResponse {
			e = errors.New(fmt.Sprintf("Bad PacketType: %d", b))
			goto reset
		}

		parser.packet.PacketType = b
		parser.state = packetParseStateMessageType

	case packetParseStateMessageType:
		parser.packet.MessageType = b

		if parser.packet.Length == 3 {
			// Get checksum, because message type counts towards length
			parser.state = packetParseStateChecksum
		} else {
			// Get body!
			parser.state = packetParseStateBody
		}

	case packetParseStateBody:
		parser.packet.Body = append(parser.packet.Body, b)

		// Subtract 3 for: packet type, message type, checksum
		if len(parser.packet.Body) == int(parser.packet.Length)-3 {
			parser.state = packetParseStateChecksum
		}

	case packetParseStateChecksum:
		// NOTE: this should be unreachable
		if err := parser.packet.Update(); err != nil {
			e = errors.New(fmt.Sprintf("Failed to compute checksum: %v %v",
				parser.packet, err))
			goto reset
		}

		if parser.packet.Checksum == b {
			p = parser.packet
		} else {
			e = errors.New(fmt.Sprintf("Failed to validate checksum: %v", parser.packet))
		}
		goto reset

	default:
		e = errors.New(fmt.Sprintf("Invalid internal state: %d", parser.state))
		goto reset
	}

	return nil, nil

reset:
	parser.state = packetParseStatePreamble
	parser.packet = nil

	return p, e
}
