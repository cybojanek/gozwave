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
	"bytes"
	"crypto/rand"
	"github.com/cybojanek/gozwave/packet"
	"testing"
)

// Parse bytes into a packet
func parsePacketBytes(t *testing.T, bytes []byte) *packet.Packet {
	parser := packet.Parser{}
	for i, x := range bytes {
		p, err := parser.Parse(x)
		if i != len(bytes)-1 {
			if p != nil || err != nil {
				t.Errorf("Expected nil packet and nil error: %v %v", p, err)
				t.FailNow()
			}
		} else {
			if p == nil || err != nil {
				t.Errorf("Expected non nil packet and nil error: %v %v", p, err)
				t.FailNow()
			}
			return p
		}
	}
	t.Errorf("Empty bytes")
	t.FailNow()
	return nil
}

func TestApplicationCommandResponse(t *testing.T) {
	p := &packet.Packet{Preamble: packet.PacketPreambleSOF,
		PacketType:  packet.PacketTypeResponse,
		MessageType: MessageTypeApplicationCommand}

	for i := 0; i < 0xff-3; i++ {
		p.Body = make([]uint8, i)
		rand.Read(p.Body)
		p.Update()

		if i < 3 {
			if command, err := ApplicationCommandResponse(p); command != nil || err == nil {
				t.Errorf("Expected nil command and non nil error: %v %v", command, err)
			}

			continue
		}

		p.Body[2] = uint8(i - 3)
		p.Update()
		if command, err := ApplicationCommandResponse(p); command == nil || err != nil {
			t.Errorf("Expected non nil command and nil error: %v %v", command, err)
			continue
		} else if !bytes.Equal(p.Body[3:], command.Body) {
			t.Errorf("Command body mismatch")
		}

		// Bad message type
		p.MessageType++
		p.Update()
		if command, err := ApplicationCommandResponse(p); command != nil || err == nil {
			t.Errorf("Expected nil command and non nil error: %v %v", command, err)
		}
		p.MessageType--

		// Check for bad length
		p.Body[2] = uint8(i - 4)
		p.Update()
		if command, err := ApplicationCommandResponse(p); command != nil || err == nil {
			t.Errorf("Expected nil command and non nil error: %v %v", command, err)
		}

		p.Body[2] = uint8(i - 2)
		p.Update()
		if command, err := ApplicationCommandResponse(p); command != nil || err == nil {
			t.Errorf("Expected nil command and non nil error: %v %v", command, err)
		}
	}
}

func TestGetVersionResponse(t *testing.T) {
	p := &packet.Packet{Preamble: packet.PacketPreambleSOF,
		PacketType:  packet.PacketTypeResponse,
		MessageType: MessageTypeGetVersion,
		Body:        []uint8{0x68, 0x65, 0x6c, 0x6c, 0x6f, 0xf9}}
	p.Update()

	version, err := GetVersionResponse(p)
	if version == nil || err != nil {
		t.Errorf("Expected non nil version and nil error: %v %v", version, err)
		t.FailNow()
	}

	if version.Info != "hello" || version.LibraryType != 0xf9 {
		t.Errorf("Bad Info/LibraryType: %v %v", version.Info, version.LibraryType)
	}

	// Bad MessageType
	p.MessageType++
	p.Update()
	if version, err := GetVersionResponse(p); version != nil || err == nil {
		t.Errorf("Expected nil version and non nil error: %v %v", version, err)
	}
	p.MessageType--

	// Body too short
	for i := 0; i < 2; i++ {
		p.Body = make([]uint8, i)
		p.Update()
		if version, err := GetVersionResponse(p); version != nil || err == nil {
			t.Errorf("Expected nil version and non nil error: %v %v", version, err)
		}
	}
}

func TestMemoryGetIDResponse(t *testing.T) {
	p := &packet.Packet{Preamble: packet.PacketPreambleSOF,
		PacketType:  packet.PacketTypeResponse,
		MessageType: MessageTypeMemoryGetID,
		Body:        []uint8{0x12, 0x34, 0x56, 0x78, 0x9a}}
	p.Update()

	id, err := MemoryGetIDResponse(p)
	if id == nil || err != nil {
		t.Errorf("Expected non nil id and nil error: %v %v", id, err)
		t.FailNow()
	}

	if id.HomeID != 0x12345678 || id.NodeID != 0x9a {
		t.Errorf("Bad HomeID/NodeID: %v %v", id.HomeID, id.NodeID)
	}

	// Bad MessageType
	p.MessageType++
	p.Update()
	if id, err := MemoryGetIDResponse(p); id != nil || err == nil {
		t.Errorf("Expected nil id and non nil error: %v %v", id, err)
	}
	p.MessageType--

	// Bad body length
	for i := 0; i < 32; i++ {
		if i == 5 {
			continue
		}

		p.Body = make([]uint8, i)
		p.Update()
		if id, err := MemoryGetIDResponse(p); id != nil || err == nil {
			t.Errorf("Expected nil id and non nil error: %v %v", id, err)
		}
	}
}

func TestSerialAPIGetCapabilitiesResponse(t *testing.T) {
	// Payload taken from serial dump.
	packetBytes := []uint8{0x01, 0x2b, 0x01, 0x07,
		0x10, 0x20, 0x35, 0x86, 0x19, 0xa7, 0x87, 0x23,
		0x07, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0xa7, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x81, 0x05, 0x00, 0x40,
		0x2d}
	packet := parsePacketBytes(t, packetBytes)
	message, err := SerialAPIGetCapabilitiesResponse(packet)
	if message == nil || err != nil {
		t.Errorf("Expected non nil message and nil error: %v %v", message, err)
		t.FailNow()
	}

	if message.Application.Version != 0x10 {
		t.Errorf("Expected Version: 0x%02x got: 0x%02x", 0x10, message.Application.Version)
	}

	if message.Application.Revision != 0x20 {
		t.Errorf("Expected Revision: 0x%02x got: 0x%02x", 0x20, message.Application.Revision)
	}

	if message.Manufacturer != 0x3586 {
		t.Errorf("Expected Version: 0x%04x got: 0x%04x", 0x3586, message.Manufacturer)
	}

	if message.Product.Type != 0x19a7 || message.Product.ID != 0x8723 {
		t.Errorf("Expcted Product: 0x%04x 0x%04x got: 0x%04x 0x%04x",
			0x19a7, 0x8723, message.Product.Type, message.Product.ID)
	}

	expectedMessageTypes := []byte{1, 2, 3, 10, 97, 98, 99, 102, 104, 225, 232,
		233, 235, 255}
	if !bytes.Equal(expectedMessageTypes, message.MessageTypes) {
		t.Errorf("Expected MessageTypes: %v got: %v",
			expectedMessageTypes, message.MessageTypes)
	}

	// Bad MessageType 256
	packet.Body[len(packet.Body)-1] |= 0x80
	message, err = SerialAPIGetCapabilitiesResponse(packet)
	if message != nil || err == nil {
		t.Errorf("Expected nil message and non nil error: %v %v", message, err)
	}
	packet.Body[len(packet.Body)-1] &= (0xff ^ 0x80)

	// Bad MessageType
	packet.MessageType++
	message, err = SerialAPIGetCapabilitiesResponse(packet)
	if message != nil || err == nil {
		t.Errorf("Expected nil message and non nil error: %v %v", message, err)
	}
	packet.MessageType--

	// Bad BodyLength
	packet.Body = packet.Body[0 : len(packet.Body)-1]
	message, err = SerialAPIGetCapabilitiesResponse(packet)
	if message != nil || err == nil {
		t.Errorf("Expected nil message and non nil error: %v %v", message, err)
	}
	packet.Body = packet.Body[0 : len(packet.Body)+1]
}

func TestSerialAPIGetInitDataResponse(t *testing.T) {
	// Payload taken from serial dump.
	packetBytes := []uint8{0x01, 0x25, 0x01, 0x02,
		0x15, 0x23, 0x1d,
		0x07, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0xa7, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x81,
		0x05, 0x00,
		0xd4}
	packet := parsePacketBytes(t, packetBytes)
	message, err := SerialAPIGetInitDataResponse(packet)
	if message == nil || err != nil {
		t.Errorf("Expected non nil message and nil error: %v %v", message, err)
		t.FailNow()
	}

	if message.Version != 0x15 {
		t.Errorf("Expected Version: %d got: %d", 0x15, message.Version)
	}

	expectedNodes := []uint8{1, 2, 3, 10, 97, 98, 99, 102, 104, 225, 232}
	if !bytes.Equal(expectedNodes, message.Nodes) {
		t.Errorf("Expected Nodes: %v got: %v", expectedNodes, message.Nodes)
	}

	if message.Capabilities.Secondary {
		t.Errorf("Expected not Secondary")
	}

	if message.Capabilities.StaticUpdate {
		t.Errorf("Expected not StaticUpdate")
	}

	// Capabilities
	packet.Body[1] = 0x4
	message, err = SerialAPIGetInitDataResponse(packet)
	if message == nil || err != nil {
		t.Errorf("Expected non nil message and nil error: %v %v", message, err)
		t.FailNow()
	}
	if !message.Capabilities.Secondary {
		t.Errorf("Expected Secondary")
	}

	if message.Capabilities.StaticUpdate {
		t.Errorf("Expected not StaticUpdate")
	}

	packet.Body[1] = 0x8
	message, err = SerialAPIGetInitDataResponse(packet)
	if message == nil || err != nil {
		t.Errorf("Expected non nil message and nil error: %v %v", message, err)
		t.FailNow()
	}
	if message.Capabilities.Secondary {
		t.Errorf("Expected not Secondary")
	}

	if !message.Capabilities.StaticUpdate {
		t.Errorf("Expected StaticUpdate")
	}

	packet.Body[1] = 0xc
	message, err = SerialAPIGetInitDataResponse(packet)
	if message == nil || err != nil {
		t.Errorf("Expected non nil message and nil error: %v %v", message, err)
		t.FailNow()
	}
	if !message.Capabilities.Secondary {
		t.Errorf("Expected Secondary")
	}

	if !message.Capabilities.StaticUpdate {
		t.Errorf("Expected StaticUpdate")
	}

	// Bad Body length
	packet.Body = packet.Body[0 : len(packet.Body)-1]
	message, err = SerialAPIGetInitDataResponse(packet)
	if message != nil || err == nil {
		t.Errorf("Expected nil message and non nil error: %v %v", message, err)
	}
	packet.Body = packet.Body[0 : len(packet.Body)+1]

	// Bad Bitmap length
	packet.Body[2]++
	message, err = SerialAPIGetInitDataResponse(packet)
	if message != nil || err == nil {
		t.Errorf("Expected nil message and non nil error: %v %v", message, err)
	}
	packet.Body[2]--

	// Bad MessageType
	packet.MessageType++
	message, err = SerialAPIGetInitDataResponse(packet)
	if message != nil || err == nil {
		t.Errorf("Expected nil message and non nil error: %v %v", message, err)
	}
	packet.MessageType--
}

func TestZWApplicationUpdateResponse(t *testing.T) {
	p := &packet.Packet{Preamble: packet.PacketPreambleSOF,
		PacketType:  packet.PacketTypeResponse,
		MessageType: MessageTypeZWApplicationUpdate}

	for i := 0; i < 0xff-3; i++ {
		p.Body = make([]uint8, i)
		rand.Read(p.Body)
		p.Update()

		if i < 3 {
			if update, err := ZWApplicationUpdateResponse(p); update != nil || err == nil {
				t.Errorf("Expected nil update and non nil error: %v %v", update, err)
			}

			continue
		}

		p.Body[2] = uint8(i - 3)
		p.Update()
		if update, err := ZWApplicationUpdateResponse(p); update == nil || err != nil {
			t.Errorf("Expected non nil update and nil error: %v %v", update, err)
			continue
		} else if !bytes.Equal(p.Body[3:], update.Body) {
			t.Errorf("Command body mismatch")
		} else if update.Status != p.Body[0] {
			t.Errorf("Status mismatch")
		} else if update.NodeID != p.Body[1] {
			t.Errorf("NodeID mismatch")
		}

		// Bad message type
		p.MessageType++
		p.Update()
		if update, err := ZWApplicationUpdateResponse(p); update != nil || err == nil {
			t.Errorf("Expected nil update and non nil error: %v %v", update, err)
		}
		p.MessageType--

		// Check for bad length
		p.Body[2] = uint8(i - 4)
		p.Update()
		if update, err := ZWApplicationUpdateResponse(p); update != nil || err == nil {
			t.Errorf("Expected nil update and non nil error: %v %v", update, err)
		}

		p.Body[2] = uint8(i - 2)
		p.Update()
		if update, err := ZWApplicationUpdateResponse(p); update != nil || err == nil {
			t.Errorf("Expected nil update and non nil error: %v %v", update, err)
		}
	}
}

func TestZWGetControllerCapabilitiesResponse(t *testing.T) {
	packetBytes := []uint8{0x1, 0x4, 0x1, 0x5, 0xf, 0xf0}
	packet := parsePacketBytes(t, packetBytes)

	message, err := ZWGetControllerCapabilitiesResponse(packet)
	if message == nil || err != nil {
		t.Errorf("Expected non nil message and nil error: %v %v", message, err)
		t.FailNow()
	}

	// Test capabilities
	compatibilityBytes := []uint8{0x0, 0x1, 0x2, 0x4, 0x8, 0x10, 0x12, 0x14, 0x18, 0x6}
	compatibilities := [][5]bool{{false, false, false, false, false},
		{true, false, false, false, false},
		{false, true, false, false, false},
		{false, false, true, false, false},
		{false, false, false, true, false},
		{false, false, false, false, true},
		{false, true, false, false, true},
		{false, false, true, false, true},
		{false, false, false, true, true},
		{false, true, true, false, false}}

	for i, b := range compatibilityBytes {
		packet.Body[0] = b
		message, err = ZWGetControllerCapabilitiesResponse(packet)
		if message == nil || err != nil {
			t.Errorf("Expected non nil message and nil error: %v %v", message, err)
			t.FailNow()
		}

		expectedBools := compatibilities[i]
		actualBools := [5]bool{message.Secondary, message.NonStandardHomeID,
			message.StaticUpdateControllerIDServer, message.WasPrimary,
			message.StaticUpdateController}
		if expectedBools != actualBools {
			t.Errorf("Expected Bools: %v got: %v", expectedBools, actualBools)
		}
	}

	// Bad MessageType
	packet.MessageType++
	message, err = ZWGetControllerCapabilitiesResponse(packet)
	if message != nil || err == nil {
		t.Errorf("Expected nil message and non nil error: %v %v", message, err)
	}
	packet.MessageType--

	// Bad BodyLength
	packet.Body = packet.Body[0 : len(packet.Body)-1]
	message, err = ZWGetControllerCapabilitiesResponse(packet)
	if message != nil || err == nil {
		t.Errorf("Expected nil message and non nil error: %v %v", message, err)
	}
	packet.Body = packet.Body[0 : len(packet.Body)+1]
}

func TestZWGetNodeProtocolInfoResponse(t *testing.T) {
	p := &packet.Packet{Preamble: packet.PacketPreambleSOF,
		PacketType:  packet.PacketTypeResponse,
		MessageType: MessageTypeZWGetNodeProtocolInfo,
		Body:        []uint8{0x80, 0xff, 0xff, 0x12, 0x34, 0x56}}
	p.Update()

	// Parse A
	message, err := ZWGetNodeProtocolInfoResponse(p)
	if message == nil || err != nil {
		t.Errorf("Expected non nil message and nil error: %v %v", message, err)
		t.FailNow()
	}

	if message.DeviceClass.Basic != 0x12 || message.DeviceClass.Generic != 0x34 || message.DeviceClass.Specific != 0x56 {
		t.Errorf("Bad Basic/Generic: %v %v %v", message.DeviceClass.Basic,
			message.DeviceClass.Generic, message.DeviceClass.Specific)
	}

	if !message.Capabilities.Listening {
		t.Errorf("Bad Listening")
	}

	// Parse B
	p.Body = []uint8{0x7f, 0xff, 0xff, 0x78, 0x9a, 0xbc}
	p.Update()

	message, err = ZWGetNodeProtocolInfoResponse(p)
	if message == nil || err != nil {
		t.Errorf("Expected non nil message and nil error: %v %v", message, err)
		t.FailNow()
	}

	if message.DeviceClass.Basic != 0x78 || message.DeviceClass.Generic != 0x9a || message.DeviceClass.Specific != 0xbc {
		t.Errorf("Bad Basic/Generic: %v %v %v", message.DeviceClass.Basic,
			message.DeviceClass.Generic, message.DeviceClass.Specific)
	}

	if message.Capabilities.Listening {
		t.Errorf("Bad Listening")
	}

	// Bad MessageType
	p.MessageType++
	p.Update()
	if message, err = ZWGetNodeProtocolInfoResponse(p); message != nil || err == nil {
		t.Errorf("Expected nil message and non nil error: %v %v", message, err)
	}
	p.MessageType--

	// Bad body length
	for i := 0; i < 32; i++ {
		if i == 6 {
			continue
		}

		p.Body = make([]uint8, i)
		p.Update()
		if message, err := ZWGetNodeProtocolInfoResponse(p); message != nil || err == nil {
			t.Errorf("Expected nil message and non nil error: %v %v", message, err)
		}
	}
}

func TestZWRequestNodeInfoResponse(t *testing.T) {
	p := &packet.Packet{Preamble: packet.PacketPreambleSOF,
		PacketType:  packet.PacketTypeResponse,
		MessageType: MessageTypeZWRequestNodeInfo,
		Body:        []uint8{0x56}}
	p.Update()

	// Parse A
	message, err := ZWRequestNodeInfoResponse(p)
	if message == nil || err != nil {
		t.Errorf("Expected non nil message and nil error: %v %v", message, err)
		t.FailNow()
	}

	if message.Status != 0x56 {
		t.Errorf("Bad Status: %v", message.Status)
	}

	// Bad MessageType
	p.MessageType++
	p.Update()
	if message, err = ZWRequestNodeInfoResponse(p); message != nil || err == nil {
		t.Errorf("Expected nil message and non nil error: %v %v", message, err)
	}
	p.MessageType--

	// Bad body length
	for i := 0; i < 32; i++ {
		if i == 1 {
			continue
		}

		p.Body = make([]uint8, i)
		p.Update()
		if message, err := ZWRequestNodeInfoResponse(p); message != nil || err == nil {
			t.Errorf("Expected nil message and non nil error: %v %v", message, err)
		}
	}
}

func TestZWSendDataResponse(t *testing.T) {
	p := &packet.Packet{Preamble: packet.PacketPreambleSOF,
		PacketType:  packet.PacketTypeResponse,
		MessageType: MessageTypeZWSendData,
		Body:        []uint8{0x12, 0x34, 0x56, 0x78}}
	p.Update()

	// Parse A
	message, err := ZWSendDataResponse(p)
	if message == nil || err != nil {
		t.Errorf("Expected non nil message and nil error: %v %v", message, err)
		t.FailNow()
	}

	if message.CallbackID != 0x12 || message.Status != 0x34 || message.TransmitTime != 0x5678 {
		t.Errorf("Bad CallbackID/Status/TransmitTime: %v %v %v", message.CallbackID,
			message.Status, message.TransmitTime)
	}

	// Bad MessageType
	p.MessageType++
	p.Update()
	if message, err = ZWSendDataResponse(p); message != nil || err == nil {
		t.Errorf("Expected nil message and non nil error: %v %v", message, err)
	}
	p.MessageType--

	// Bad body length
	for i := 0; i < 32; i++ {
		if i == 4 {
			continue
		}

		p.Body = make([]uint8, i)
		p.Update()
		if message, err := ZWSendDataResponse(p); message != nil || err == nil {
			t.Errorf("Expected nil message and non nil error: %v %v", message, err)
		}
	}
}
