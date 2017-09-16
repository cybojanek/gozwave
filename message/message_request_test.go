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

func TestSerialAPIGetInitDataRequest(t *testing.T) {
	if p := SerialAPIGetInitDataRequest(); p == nil {
		t.Logf("Expected non nil packet")
	}
}

func TestSerialAPIGetCapabilitiesRequest(t *testing.T) {
	if p := SerialAPIGetCapabilitiesRequest(); p == nil {
		t.Logf("Expected non nil packet")
	}
}

func TestZWGetControllerCapabilitiesRequest(t *testing.T) {
	if p := ZWGetControllerCapabilitiesRequest(); p == nil {
		t.Logf("Expected non nil packet")
	}
}

func TestZWGetNodeProtocolInfoRequest(t *testing.T) {
	for i := 0; i < 0xff; i++ {
		nodeID := uint8(i)
		p, err := ZWGetNodeProtocolInfoRequest(nodeID)
		if IsValidNodeID(nodeID) {
			if p == nil || err != nil {
				t.Logf("Expected non nil packet and nil error: %v %v", p, err)
			}
		} else {
			if p != nil || err == nil {
				t.Logf("Expected nil packet and non nil error: %v %v", p, err)
			}
		}
	}
}
