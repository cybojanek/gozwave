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
	// "bytes"
	// "github.com/cybojanek/gozwave/packet"
	"testing"
)

func TestGetVersionRequest(t *testing.T) {
	GetVersionRequest()
}

func TestMemoryGetIDRequest(t *testing.T) {
	MemoryGetIDRequest()
}

func TestSerialAPIGetInitDataRequest(t *testing.T) {
	SerialAPIGetInitDataRequest()
}

func TestSerialAPIGetCapabilitiesRequest(t *testing.T) {
	SerialAPIGetCapabilitiesRequest()
}

func TestZWGetControllerCapabilitiesRequest(t *testing.T) {
	ZWGetControllerCapabilitiesRequest()
}

func TestZWGetNodeProtocolInfoRequest(t *testing.T) {
	for i := 0; i < 0xff; i++ {
		nodeID := uint8(i)
		p, err := ZWGetNodeProtocolInfoRequest(nodeID)
		if IsValidNodeID(nodeID) {
			if p == nil || err != nil {
				t.Errorf("Expected non nil packet and nil error: %v %v", p, err)
			}
		} else {
			if p != nil || err == nil {
				t.Errorf("Expected nil packet and non nil error: %v %v", p, err)
			}
		}
	}
}

func TestZWRequestNodeInfoRequest(t *testing.T) {
	for i := 0; i < 0xff; i++ {
		nodeID := uint8(i)
		p, err := ZWRequestNodeInfoRequest(nodeID)
		if IsValidNodeID(nodeID) {
			if p == nil || err != nil {
				t.Errorf("Expected non nil packet and nil error: %v %v", p, err)
			}
		} else {
			if p != nil || err == nil {
				t.Errorf("Expected nil packet and non nil error: %v %v", p, err)
			}
		}
	}
}

func TestZWSendDataRequest(t *testing.T) {
	// -3 for package and -5 for ZWSendDataRequest
	maxPayloadLength := 0xff - 3 - 4

	for i := 0; i < 0xff; i++ {
		nodeID := uint8(i)

		for b := 0; b < maxPayloadLength+1; b++ {
			payload := make([]uint8, b)

			// No callbackID
			p, err := ZWSendDataRequest(nodeID, 0, payload, 0, 0)
			if IsValidNodeID(nodeID) && b <= maxPayloadLength {
				if p == nil || err != nil {
					t.Errorf("Expected non nil packet and nil error: %v %v", p, err)
				}
			} else {
				if p != nil || err == nil {
					t.Errorf("Expected nil packet and non nil error: %v %v", p, err)
				}
			}

			// With callbackID
			p, err = ZWSendDataRequest(nodeID, 0, payload, 0, 1)
			if IsValidNodeID(nodeID) && b <= maxPayloadLength-1 {
				if p == nil || err != nil {
					t.Errorf("Expected non nil packet and nil error: %v %v", p, err)
				}
			} else {
				if p != nil || err == nil {
					t.Errorf("Expected nil packet and non nil error: %v %v", p, err)
				}
			}
		}
	}
}
