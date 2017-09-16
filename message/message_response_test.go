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

func TestSerialAPIGetInitDataResponse(t *testing.T) {
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

func TestSerialAPIGetCapabilitiesResponse(t *testing.T) {
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

	if message.Version != 0x2010 {
		t.Errorf("Expected Version: 0x%04x got: 0x%04x", 0x2010, message.Version)
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
