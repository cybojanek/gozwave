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
