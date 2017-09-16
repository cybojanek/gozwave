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
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/cybojanek/gozwave/packet"
)

// SerialAPIGetCapabilitiesResponse parses a SerialAPIGetCapabilities response
// packet
func SerialAPIGetCapabilitiesResponse(p *packet.Packet) (*SerialAPIGetCapabilities, error) {
	if p.MessageType != MessageTypeSerialAPIGetCapabilities {
		return nil, fmt.Errorf("Bad MessageType: %d", p.MessageType)
	}

	if len(p.Body) != 40 {
		return nil, fmt.Errorf("Bad Body length: %d", len(p.Body))
	}

	message := SerialAPIGetCapabilities{}
	// FIXME: why are these supposed to be the same endianness?
	message.Version = binary.LittleEndian.Uint16(p.Body[0:2])
	message.Manufacturer = binary.BigEndian.Uint16(p.Body[2:4])
	message.Product.Type = binary.BigEndian.Uint16(p.Body[4:6])
	message.Product.ID = binary.BigEndian.Uint16(p.Body[6:8])

	for i, x := range p.Body[8:] {
		for b := uint8(0); b < 8; b++ {
			if (x & (1 << b)) != 0 {
				if i == 31 && b == 7 {
					return nil, errors.New("Unexpected supported MessageType 256")
				}
				message.MessageTypes = append(message.MessageTypes,
					1+(uint8(i)*8)+b)
			}
		}
	}

	return &message, nil
}

// SerialAPIGetInitDataResponse parses a SerialAPIGetInitData response packet
func SerialAPIGetInitDataResponse(p *packet.Packet) (*SerialAPIGetInitData, error) {
	if p.MessageType != MessageTypeSerialAPIGetInitData {
		return nil, fmt.Errorf("Bad MessageType: %d", p.MessageType)
	}

	if len(p.Body) != 34 {
		return nil, fmt.Errorf("Bad Body length: %d", len(p.Body))
	}

	message := SerialAPIGetInitData{}
	message.Version = p.Body[0]

	capabilities := p.Body[1]
	message.Capabilities.Secondary = (capabilities & 0x4) != 0
	message.Capabilities.StaticUpdate = (capabilities & 0x8) != 0

	// Should be 29 for 29 * 8 = 232 bits / node ids
	if p.Body[2] != 29 {
		return nil, fmt.Errorf("Bad bitmap byte length: %d", p.Body[2])
	}

	for i, x := range p.Body[3 : 3+29] {
		for b := uint8(0); b < 8; b++ {
			if (x & (1 << b)) != 0 {
				message.Nodes = append(message.Nodes, 1+uint8(i)*8+b)
			}
		}
	}

	// FIXME: what are these used for: packet.Body[32:34]

	return &message, nil
}

// ZWGetControllerCapabilitiesResponse parses a ZWGetControllerCapabilities
// response packet
func ZWGetControllerCapabilitiesResponse(p *packet.Packet) (*ZWGetControllerCapabilities, error) {
	if p.MessageType != MessageTypeZWGetControllerCapabilities {
		return nil, fmt.Errorf("Bad MessageType: %d", p.MessageType)
	}

	if len(p.Body) != 1 {
		return nil, fmt.Errorf("Bad Body length: %d", len(p.Body))
	}

	message := ZWGetControllerCapabilities{}
	capabilities := p.Body[0]
	message.Secondary = (capabilities & 0x1) != 0
	message.NonStandardHomeID = (capabilities & 0x2) != 0
	message.StaticUpdateControllerIDServer = (capabilities & 0x4) != 0
	message.WasPrimary = (capabilities & 0x8) != 0
	message.StaticUpdateController = (capabilities & 0x10) != 0

	return &message, nil
}

// ZWGetNodeProtocolInfoResponse parses a ZWGetNodeProtocolInfo response packet
func ZWGetNodeProtocolInfoResponse(p *packet.Packet) (*ZWGetNodeProtocolInfo, error) {
	if p.MessageType != MessageTypeZWGetNodeProtocolInfo {
		return nil, fmt.Errorf("Bad MessageType: %d", p.MessageType)
	}

	if len(p.Body) != 6 {
		return nil, fmt.Errorf("Bad Body length: %d", len(p.Body))
	}

	message := ZWGetNodeProtocolInfo{}

	message.Capabilities.Listening = (p.Body[0] & 0x80) != 0

	message.DeviceClass.Basic = p.Body[3]
	message.DeviceClass.Generic = p.Body[4]
	message.DeviceClass.Specific = p.Body[5]

	return &message, nil
}
