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

	if message, err := api.SerialAPIGetInitData(); err != nil {
		t.Errorf("Expected nil error: %v", err)
	} else {
		t.Logf("SerialAPIGetInitData: %+v", message)
	}

	if message, err := api.SerialAPIGetCapabilities(); err != nil {
		t.Errorf("Expected nil error: %v", err)
	} else {
		t.Logf("SerialAPIGetCapabilities: %+v", message)
	}

	if message, err := api.ZWGetControllerCapabilities(); err != nil {
		t.Errorf("Expected nil error: %v", err)
	} else {
		t.Logf("ZWGetControllerCapabilities: %+v", message)
	}
}
