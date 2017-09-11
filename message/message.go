// Package message converts Packets to higher level ZWave application messages
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
)

// Basic Command
const (
	BasicCommandSet    uint8 = 0x01
	BasicCommandGet          = 0x02
	BasicCommandReport       = 0x03
)

// Command Class
const (
	CommandClassBasic                 uint8 = 0x20
	ComamndClassControllerReplication       = 0x21
	CommandClassApplicationStatus           = 0x22
	CommandClassHail                        = 0x23
)

// Message Type
const (
	MessageTypeNone                        uint8 = 0x00
	MessageTypeSerialAPIGetInitdata              = 0x02
	MessageTypeApplicationCommandHandler         = 0x04
	MessageTypeZWGetControllerCapabilities       = 0x05
	MessageTypeSerialAPIGetCapabilities          = 0x07
	MessageTypeZWSendData                        = 0x13
	MessageTypeZWApplicationUpdate               = 0x49
)

// Transmit Option
const (
	TransmitOptionACK       uint8 = 0x01
	TransmitOptionLowPower        = 0x02
	TransmitOptionAutoRoute       = 0x04
	TransmitOptionNoRoute         = 0x10
	TransmitOptionExplore         = 0x20
)

// Transmit Complete
const (
	TransmitCompleteOK      uint8 = 0x00
	TransmitCompleteNoAck         = 0x01
	TransmitCompleteFail          = 0x02
	TransmitCompleteNotIdle       = 0x03
	TransmitCompleteNoRoute       = 0x04
)

// SerialAPIGetInitData information
type SerialAPIGetInitData struct {
	Version      uint8
	Capabilities struct {
		Secondary    bool
		StaticUpdate bool
	}
	Nodes []uint8 // List of nodes on the network
}

// SerialAPIGetCapabilities information
type SerialAPIGetCapabilities struct {
	Version      uint16
	Manufacturer uint16
	Product      struct {
		Type uint16
		ID   uint16
	}
	MessageTypes []uint8 // List of supported Message Types
}

// ZWGetControllerCapabilities information
type ZWGetControllerCapabilities struct {
	Secondary                      bool
	NonStandardHomeID              bool
	StaticUpdateControllerIDServer bool
	WasPrimary                     bool
	StaticUpdateController         bool
}

// ZWSendData information
type ZWSendData struct {
	NodeID          uint8
	CommandClass    uint8
	Payload         []uint8
	TransmitOptions struct {
		ACK       bool
		LowPower  bool
		AutoRoute bool
		NoRoute   bool
		Explore   bool
	}
	CallbackID uint8
}

// SetTransmitOptionsMask takes the options byte and sets the individual bool
// flags of ZWSendData.TransmitOptions
func (message *ZWSendData) SetTransmitOptionsMask(options uint8) error {
	fullMask := (TransmitOptionACK | TransmitOptionLowPower |
		TransmitOptionAutoRoute | TransmitOptionNoRoute | TransmitOptionExplore)

	if (fullMask & options) != options {
		return fmt.Errorf("Options contains unknown bits: 0x%02x", options)
	}

	message.TransmitOptions.ACK = (options & TransmitOptionACK) != 0
	message.TransmitOptions.LowPower = (options & TransmitOptionLowPower) != 0
	message.TransmitOptions.AutoRoute = (options & TransmitOptionAutoRoute) != 0
	message.TransmitOptions.NoRoute = (options & TransmitOptionNoRoute) != 0
	message.TransmitOptions.Explore = (options & TransmitOptionExplore) != 0

	return nil
}
