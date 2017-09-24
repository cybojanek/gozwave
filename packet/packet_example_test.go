package packet_test

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

func ExamplePacket_Bytes() {
	p := packet.Packet{Preamble: packet.PacketPreambleACK}
	if b, err := p.Bytes(); err != nil {
		fmt.Printf("Failed to encode: %v\n", err)
	} else {
		fmt.Printf("Bytes: %v\n", b)
	}
	// Output: Bytes: [6]
}

func ExamplePacket_Update() {
	p := packet.Packet{Preamble: packet.PacketPreambleSOF,
		PacketType:  packet.PacketTypeRequest,
		MessageType: 0x02}
	if err := p.Update(); err != nil {
		fmt.Printf("Failed to update: %v\n", err)
	}
	fmt.Printf("Packet: %+v\n", p)
	// Output: Packet: {Preamble:1 Length:3 PacketType:0 MessageType:2 Body:[] Checksum:254}
}

func ExampleParser_Parse() {
	parser := packet.Parser{}
	data := []byte{0x01, 0x05, 0x00, 0x78, 0x65, 0xd3, 0x34, 0x06, 0x23, 0x15}
	for _, x := range data {
		if packet, err := parser.Parse(x); err != nil {
			fmt.Printf("Failed to parse: %v\n", err)
		} else if packet != nil {
			fmt.Printf("Got Packet: %+v\n", packet)
		}
	}
	// Output: Got Packet: {Preamble: 0x01 Length: 0x05 PacketType: 0x00 MessageType: 0x78 Body: 0x65 0xd3 Checksum: 0x34}
	// Got Packet: {Preamble: 0x06 Length: 0x00 PacketType: 0x00 MessageType: 0x00 Body:  Checksum: 0x00}
	// Failed to parse: Bad preamble: 35
	// Got Packet: {Preamble: 0x15 Length: 0x00 PacketType: 0x00 MessageType: 0x00 Body:  Checksum: 0x00}
}
