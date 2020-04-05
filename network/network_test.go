package network

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
	"time"
)

// FIXME: mock out or parameterize
const testDevicePath = "/dev/tty.usbmodem1451"

func TestApiOpenClose(t *testing.T) {
	net := Network{DevicePath: testDevicePath}

	if testing.Short() {
		t.Skipf("Skipping API test")
	}

	if err := net.Open(); err != nil {
		t.Errorf("Expected nil error: %v", err)
		t.FailNow()
	}

	defer func() {
		if err := net.Close(); err != nil {
			t.Errorf("Expected nil error: %v", err)
		}
	}()

	if err := net.Close(); err != nil {
		t.Errorf("Expected nil error: %v", err)
	}

	if err := net.Open(); err != nil {
		t.Errorf("Expected nil error: %v", err)
	}

	if err := net.Open(); err != nil {
		t.Errorf("Expected nil error: %v", err)
	}

	if err := net.Close(); err != nil {
		t.Errorf("Expected nil error: %v", err)
	}
}

func TestAPI(t *testing.T) {
	net := Network{DevicePath: testDevicePath}
	net.DebugLogging = false

	if testing.Short() {
		t.Skipf("Skipping API test")
	}

	if err := net.Open(); err != nil {
		t.Errorf("Expected nil error: %v", err)
		t.FailNow()
	}

	defer func() {
		if err := net.Close(); err != nil {
			t.Errorf("Expected nil error: %v", err)
		}
	}()

	var err error

	err = net.Initialize()
	if err != nil {
		t.Errorf("Expected nil error: %v", err)
		t.FailNow()
	}

	nodes := net.GetNodes()
	for _, node := range nodes {
		t.Logf("Loading node: %d", node.ID)
		if _, err := node.Load(nil); err != nil {
			t.Errorf("Node: %d Expected nil error: %v", node.ID, err)
		}
	}

	s := net.GetNode(14)
	bs := s.GetBinarySwitch()
	if bs != nil {
		if err := bs.On(); err != nil {
			t.Logf("Failed to turn on node: %v", err)
		}
		if isOn, err := bs.IsOn(); err != nil {
			t.Logf("Failed to check node on: %v", err)
		} else if !isOn {
			t.Logf("Expected switch to be on")
		}

		time.Sleep(time.Second * 1)

		if err := bs.Off(); err != nil {
			t.Logf("Failed to turn off node: %v", err)
		}
		if isOn, err := bs.IsOn(); err != nil {
			t.Logf("Failed to check node off: %v", err)
		} else if isOn {
			t.Logf("Expected switch to be off")
		}
	} else {
		t.Logf("Node %d is not a switch", s.ID)
	}

	for _, node := range nodes {
		t.Logf("Node: %+v", node)
	}
}
