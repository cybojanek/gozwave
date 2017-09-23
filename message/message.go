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

// Message Type
const (
	MessageTypeNone                        uint8 = 0x00
	MessageTypeSerialAPIGetInitData              = 0x02
	MessageTypeApplicationCommandHandler         = 0x04
	MessageTypeZWGetControllerCapabilities       = 0x05
	MessageTypeSerialAPIGetCapabilities          = 0x07
	MessageTypeZWSendData                        = 0x13
	MessageTypeGetVersion                        = 0x15
	MessageTypeMemoryGetID                       = 0x20
	MessageTypeZWGetNodeProtocolInfo             = 0x41
	MessageTypeZWApplicationUpdate               = 0x49
	MessageTypeZWRequestNodeInfo                 = 0x60
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
	TransmitCompleteNoACK         = 0x01
	TransmitCompleteFail          = 0x02
	TransmitCompleteNotIdle       = 0x03
	TransmitCompleteNoRoute       = 0x04
)

// Library Type
const (
	LibraryTypeControllerStatic uint8 = 0x01
	LibraryTypeController             = 0x02
	LibraryTypeSlaveEnhanced          = 0x03
	LibraryTypeSlave                  = 0x04
	LibraryTypeInstaller              = 0x05
	LibraryTypeSlaveRouting           = 0x06
	LibraryTypeControllerBridge       = 0x07
	LibraryTypeDUT                    = 0x08
	LibraryTypeAVRemote               = 0x0a
	LibraryTypeAVDevice               = 0x0b
)

// ZWApplicationUpdate Status meaning
const (
	ZWApplicationUpdateStateSUCID         uint8 = 0x10
	ZWApplicationUpdateStateDeleteDone          = 0x20
	ZWApplicationUpdateStateNewIDAssigned       = 0x40
	ZWApplicationUpdateStateRoutePending        = 0x80
	ZWApplicationUpdateStateRequestFailed       = 0x81
	ZWApplicationUpdateStateRequestDone         = 0x82
	ZWApplicationUpdateStateReceived            = 0x84
)

// ApplicationCommandHandler information
type ApplicationCommandHandler struct {
	Status uint8
	NodeID uint8
	Body   []uint8
}

// GetVersion information
type GetVersion struct {
	Info        string
	LibraryType uint8
}

// MemoryGetID informationZ
type MemoryGetID struct {
	HomeID uint32
	NodeID uint8
}

// SerialAPIGetInitData information
type SerialAPIGetInitData struct {
	Version      uint8
	Capabilities struct {
		Slave        bool
		TimerSupport bool
		Secondary    bool
		StaticUpdate bool
	}
	Nodes []uint8
}

// SerialAPIGetCapabilities information
type SerialAPIGetCapabilities struct {
	Application struct {
		Version  uint8
		Revision uint8
	}
	Manufacturer uint16
	Product      struct {
		Type uint16
		ID   uint16
	}
	MessageTypes []uint8
}

// ZWApplicationUpdate information
type ZWApplicationUpdate struct {
	Status uint8
	NodeID uint8
	Body   []uint8
}

// ZWGetControllerCapabilities information
type ZWGetControllerCapabilities struct {
	Secondary                      bool
	NonStandardHomeID              bool
	StaticUpdateControllerIDServer bool
	WasPrimary                     bool
	StaticUpdateController         bool
}

// ZWGetNodeProtocolInfo information
type ZWGetNodeProtocolInfo struct {
	Capabilities struct {
		Listening bool
	}
	DeviceClass struct {
		Basic    uint8
		Generic  uint8
		Specific uint8
	}
}

// ZWRequestNodeInfo information
type ZWRequestNodeInfo struct {
	Status uint8
}

// ZWSendData information
type ZWSendData struct {
	CallbackID uint8
	Status     uint8
}

// IsValidNodeID checks if the nodeID is in the valid range of nodes
func IsValidNodeID(nodeID uint8) bool {
	return nodeID > 0 && nodeID < 233
}
