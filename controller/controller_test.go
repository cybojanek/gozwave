package controller

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
	"time"
)

// FIXME: mock out or parameterize
const testDevicePath = "/dev/tty.usbmodem1451"

func TestControllerOpenClose(t *testing.T) {
	controller := Controller{DevicePath: testDevicePath}

	if testing.Short() {
		t.Skipf("Skipping controller test")
	}

	if err := controller.Open(); err != nil {
		t.Errorf("Expected nil error: %v", err)
		t.FailNow()
	}

	defer func() {
		if err := controller.Close(); err != nil {
			t.Errorf("Expected nil error: %v", err)
		}
	}()
}

func TestController(t *testing.T) {
	controller := Controller{DevicePath: testDevicePath}

	if testing.Short() {
		t.Skipf("Skipping controller test")
	}

	if err := controller.Open(); err != nil {
		t.Errorf("Expected nil error: %v", err)
		t.FailNow()
	}

	defer func() {
		if err := controller.Close(); err != nil {
			t.Errorf("Expected nil error: %v", err)
		}
	}()

	requestPacket := message.SerialAPIGetInitDataRequest()
	for i := 0; i < 5; i++ {
		response, err := controller.BlockingRequest(requestPacket)
		if err != nil {
			t.Errorf("Expected nil error: %v", err)
		}
		t.Logf("Reponse: %v", response)

		time.Sleep(100 * time.Millisecond)
	}
}

func TestControllerMessage(t *testing.T) {
	controller := Controller{DevicePath: testDevicePath}

	if testing.Short() {
		t.Skipf("Skipping controller test")
	}

	if err := controller.Open(); err != nil {
		t.Errorf("Expected nil error: %v", err)
		t.FailNow()
	}

	defer func() {
		if err := controller.Close(); err != nil {
			t.Errorf("Expected nil error: %v", err)
		}
	}()

	if responsePacket, err := controller.BlockingRequest(
		message.SerialAPIGetInitDataRequest()); err != nil {
		t.Errorf("Expected nil error: %v", err)
	} else if responseMessage, err := message.SerialAPIGetInitDataResponse(
		responsePacket); err != nil {
		t.Errorf("Expected nil error: %v", err)
	} else {
		t.Logf("Response message: %v", responseMessage)
	}
}
