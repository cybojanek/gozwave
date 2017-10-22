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
	"testing"
	"time"
)

func TestIsValidNodeID(t *testing.T) {
	if IsValidNodeID(0) {
		t.Errorf("Expected 0 to be an invalid node ID")
	}

	for i := uint8(1); i <= 232; i++ {
		if !IsValidNodeID(i) {
			t.Errorf("Expected node %d to be valid", i)
		}
	}

	for i := 233; i <= 255; i++ {
		if IsValidNodeID(uint8(i)) {
			t.Errorf("Expected node %d to be invalid", i)
		}
	}
}

func TestEncodeDecodeDuration(t *testing.T) {
	// Encode seconds
	for i := uint8(0); i < 128; i++ {
		if value, err := EncodeDuration(time.Second * time.Duration(i)); err != nil || value != i {
			t.Errorf("Failed encoding for %d seconds: %d %v", i, value, err)
		}

		if DecodeDuration(i) != (time.Second * time.Duration(i)) {
			t.Errorf("Failed to decode byte: 0x%02x", i)
		}
	}

	// Encode minutes
	for i := uint8(3); i < 128; i++ {
		if value, err := EncodeDuration(time.Minute * time.Duration(i)); err != nil || value != i+(0x80-1) {
			t.Errorf("Failed encoding for %d minutes: %d %v", i, value, err)
		}

		if DecodeDuration(i+(0x80-1)) != (time.Minute * time.Duration(i)) {
			t.Errorf("Failed to decode byte: 0x%02x", i)
		}
	}
}

func TestDecodeFloat(t *testing.T) {
	type testCase struct {
		binary    []uint8
		precision uint8
		result    float32
	}

	cases := []testCase{
		{binary: []uint8{0}, precision: 0, result: 0.0},
		{binary: []uint8{0}, precision: 1, result: 0.0},
		{binary: []uint8{0}, precision: 2, result: 0.0},
		{binary: []uint8{23}, precision: 0, result: 23.0},
		{binary: []uint8{23}, precision: 1, result: 2.3},
		{binary: []uint8{23}, precision: 2, result: 0.23},
		{binary: []uint8{23}, precision: 3, result: 0.023},
		{binary: []uint8{252}, precision: 0, result: -4.0},
		{binary: []uint8{252}, precision: 2, result: -0.04},
		{binary: []uint8{127, 255}, precision: 0, result: 32767.0},
		{binary: []uint8{127, 255}, precision: 3, result: 32.767},
		{binary: []uint8{255, 255}, precision: 0, result: -1.0},
		{binary: []uint8{255, 255}, precision: 1, result: -0.1},
		{binary: []uint8{255, 23}, precision: 0, result: -233.0},
		{binary: []uint8{255, 23}, precision: 2, result: -2.33},
		{binary: []uint8{127, 255, 255, 203}, precision: 0, result: 2147483647.0},
		{binary: []uint8{255, 255, 255, 203}, precision: 0, result: -53.0},
	}

	for i, test := range cases {
		if value, err := DecodeFloat(test.binary, test.precision); err != nil || value != test.result {
			t.Errorf("Failed case %d, expected %v, got %v %v", i, test.result, value, err)
		}
	}

	if _, err := DecodeFloat([]uint8{0xff, 0xff, 0xff}, 0); err == nil {
		t.Errorf("Decoding should have failed")
	}
}
