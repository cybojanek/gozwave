package api

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
	"github.com/cybojanek/gozwave/message"
	"testing"
)

// FIXME: mock out or parameterize
const testDevicePath = "/dev/tty.usbmodem1451"

func TestAPI(t *testing.T) {
	api := ZWAPI{DevicePath: testDevicePath}

	if testing.Short() {
		t.Skipf("Skipping API test")
	}

	if err := api.Open(); err != nil {
		t.Errorf("Expected nil error: %v", err)
		t.FailNow()
	}

	defer func() {
		if err := api.Close(); err != nil {
			t.Errorf("Expected nil error: %v", err)
		}
	}()

	var err error
	var serialAPIGetInitData *message.SerialAPIGetInitData
	var serialAPIGetCapabilities *message.SerialAPIGetCapabilities
	var zwGetControllerCapabilities *message.ZWGetControllerCapabilities

	serialAPIGetInitData, err = api.SerialAPIGetInitData()
	if err != nil {
		t.Errorf("Expected nil error: %v", err)
	} else {
		t.Logf("SerialAPIGetInitData: %+v", serialAPIGetInitData)
	}

	serialAPIGetCapabilities, err = api.SerialAPIGetCapabilities()
	if err != nil {
		t.Errorf("Expected nil error: %v", err)
	} else {
		t.Logf("SerialAPIGetCapabilities: %+v", serialAPIGetCapabilities)
	}

	zwGetControllerCapabilities, err = api.ZWGetControllerCapabilities()
	if err != nil {
		t.Errorf("Expected nil error: %v", err)
	} else {
		t.Logf("ZWGetControllerCapabilities: %+v", zwGetControllerCapabilities)
	}

	// Check existing nodes
	for _, nodeID := range serialAPIGetInitData.Nodes {
		if message, err := api.ZWGetNodeProtocolInfo(nodeID); err != nil {
			t.Errorf("Expected nil error: %v", err)
		} else {
			t.Logf("ZWGetNodeProtocolInfo: %d, %+v", nodeID, message)
		}
	}

	// Check non existant node
	if message, err := api.ZWGetNodeProtocolInfo(200); err != ErrNodeNotFound {
		t.Errorf("Expected non nil error ErrNodeNotFound: %v", err)
	} else {
		t.Logf("ZWGetNodeProtocolInfo: %d, %+v", 200, message)
	}
}
