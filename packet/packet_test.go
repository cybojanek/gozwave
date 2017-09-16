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
	"bytes"
	"fmt"
	"testing"
)

func TestPacket(t *testing.T) {
	parser := Parser{}

	// Test single byte packets
	singlePacketBytes := []uint8{PacketPreambleACK, PacketPreambleNAK, PacketPreambleCAN}
	for _, b := range singlePacketBytes {
		p, err := parser.Parse(b)
		if err != nil {
			t.Errorf("Expected nil error for %d: %v", b, err)
			continue
		}
		if p == nil {
			t.Errorf("Expected non nil packet for %d", b)
		}
		if err := p.Update(); err != nil {
			t.Errorf("Expected nil error for %d: %v", b, err)
		} else if p.Checksum != 0 || p.Length != 0 {
			t.Errorf("Expected 0 checksum and length: %d %d", p.Checksum, p.Length)
		}

		expectedBytes := []uint8{b}
		if actualBytes, err := p.Bytes(); err != nil {
			t.Errorf("Expected nil error for %d: %v", b, err)
		} else if !bytes.Equal(expectedBytes, actualBytes) {
			t.Errorf("Expected %v and got %v", expectedBytes, actualBytes)
		}
	}

	// Test bad preamble handling
	if p, err := parser.Parse(0x23); p != nil || err == nil {
		t.Errorf("Expected nil packet: %v and non nil error: %v", p, err)
	}

	// Test bad lengths
	badLengths := []uint8{0, 1, 2}
	for _, b := range badLengths {
		if p, err := parser.Parse(0x1); p != nil || err != nil {
			t.Errorf("Expected nil packet: %v and nil error: %v", p, err)
		}
		if p, err := parser.Parse(b); p != nil || err == nil {
			t.Errorf("Expected nil packet: %v and non nil error: %v", p, err)
		}
	}

	// Test bad PacketTypes
	for i := 2; i < 256; i++ {
		message := []byte("\x01\x03")
		message = append(message, uint8(i))

		for i, b := range message {
			p, err := parser.Parse(b)
			if i != len(message)-1 {
				if p != nil || err != nil {
					t.Errorf("Expected nil packet and nil error: %v %v", p, err)
					t.FailNow()
				}
			} else {
				if p != nil || err == nil {
					t.Errorf("Expected nil packet and non nil error: %v %v", p, err)
					t.FailNow()
				}
			}
		}
	}

	// Test bad checksum
	badChecksum := []byte("\x01\x04\x01\x02\x03\xff")
	for i, b := range badChecksum {
		p, err := parser.Parse(b)
		if i != len(badChecksum)-1 {
			if p != nil || err != nil {
				t.Errorf("Expected nil packet and nil error: %v %v", p, err)
				t.FailNow()
			}
		} else {
			if p != nil || err == nil {
				t.Errorf("Expected nil packet and non nil error: %v %v", p, err)
				t.FailNow()
			}
		}
	}

	messages := [][]byte{}
	packets := []Packet{}

	// Test 3 byte length, which skips body
	messages = append(messages, []byte("\x01\x03\x00\x78\x84"))
	packets = append(packets, Packet{Preamble: 1, Length: 3, PacketType: 0,
		MessageType: 0x78, Checksum: 0x84})

	// Test SERIAL_API_GET_INIT_DATA example
	messages = append(messages, []byte(
		"\x01\x25\x01\x02\x05\x00\x1d\x07\x00\x00"+
			"\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"+
			"\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"+
			"\x00\x11\x00\x27\x00\x14\x05\x00\xe1"))
	packets = append(packets, Packet{Preamble: 0x1, Length: 37, PacketType: 1,
		MessageType: 2, Body: messages[1][4 : len(messages[1])-1],
		Checksum: 0xe1})

	for iter, message := range messages {

		for i, b := range message {
			p, err := parser.Parse(b)
			if i != len(message)-1 {
				if p != nil || err != nil {
					t.Errorf("Expected nil packet and nil error: %v %v", p, err)
					t.FailNow()
				}
			} else {
				if err != nil {
					t.Errorf("Expected nil error on end of bytes: %v", err)
				}
				if p == nil {
					t.Errorf("Expected non nil packet for end of bytes")
				}
				// Test to string
				if v := fmt.Sprintf("%s", p); len(v) == 0 {
					t.Errorf("Expected non zero object string")
				}

				expectedPacket := packets[iter]

				if p.Preamble != expectedPacket.Preamble {
					t.Errorf("Expected Preamble: %d got: %d",
						expectedPacket.Preamble, p.Preamble)
				}

				if p.Length != expectedPacket.Length {
					t.Errorf("Expected Length: %d got: %d",
						expectedPacket.Length, p.Length)
				}

				if p.PacketType != expectedPacket.PacketType {
					t.Errorf("Expected PacketType: %d got: %d",
						expectedPacket.PacketType, p.PacketType)
				}

				if p.MessageType != expectedPacket.MessageType {
					t.Errorf("Expected MessageType: %d got: %d",
						expectedPacket.MessageType, p.MessageType)
				}

				if !bytes.Equal(expectedPacket.Body, p.Body) {
					t.Errorf("Expected Body: %v got: %v",
						expectedPacket.Body, p.Body)
				}

				if p.Checksum != expectedPacket.Checksum {
					t.Errorf("Expected Checksum: %d got: %d",
						expectedPacket.Checksum, p.Checksum)
				}

				if err := p.Update(); err != nil {
					t.Errorf("Expected nil error Update for: %d", i)
				}

				if actualBytes, err := p.Bytes(); err != nil {
					t.Errorf("Expected nil error for %d: %v", b, err)
				} else if !bytes.Equal(message, actualBytes) {
					t.Errorf("Expected %v and got %v", message, actualBytes)
				}
			}
		}
	}

	// Test bad update error handling
	goodMessage := []byte("\x01\x03\x00\x78\x84")
	for i, b := range goodMessage {
		if i == len(goodMessage)-1 {
			// Corrupt packet to trigger update error
			for x := 0; x < 0xff; x++ {
				parser.packet.Body = append(parser.packet.Body, 0xff)
			}
		}

		p, err := parser.Parse(b)
		if i != len(goodMessage)-1 {
			if p != nil || err != nil {
				t.Errorf("Expected nil packet and nil error: %v %v", p, err)
				t.FailNow()
			}
		} else {
			if err == nil {
				t.Errorf("Expected error on end of bytes")
			}
			if p != nil {
				t.Errorf("Expected nil packet for end of bytes: %v", p)
			}
		}
	}

	// Test invalid internal state
	parser.state += 20
	if p, err := parser.Parse(PacketPreambleACK); p != nil || err == nil {
		t.Errorf("Expected nil packet: %v and non nil error: %v", p, err)
	}

	// Test packet body too long
	badPacket := Packet{Preamble: PacketPreambleSOF}
	for i := 0; i <= 0xff-3; i++ {
		badPacket.Body = append(badPacket.Body, 0x00)
	}

	if err := badPacket.Update(); err == nil {
		t.Errorf("Expected non nil error")
	}

	if bytes, err := badPacket.Bytes(); bytes != nil || err == nil {
		t.Errorf("Expected nil bytes: %v and non nil error: %v", bytes, err)
	}
}
