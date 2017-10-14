package node

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

const (
	clockSet    uint8 = 0x04
	clockGet          = 0x05
	clockReport       = 0x06
)

// Clock information
type Clock struct {
	*Node
}

// GetClock returns a Clock or nil object
func (node *Node) GetClock() *Clock {
	node.mutex.Lock()
	defer node.mutex.Unlock()

	if node.supportsCommandClass(CommandClassClock) {
		return &Clock{node}
	}

	return nil
}

// Get time of the clock in weekday, hour, and minute.
// weekday is in the range of [0, 7], where 0 is unknown, and [1, 7] is
// Monday through Sunday, hour is in the range [0, 23], and minute is in the
// range of [0, 59]
func (node *Clock) Get() (uint8, uint8, uint8, error) {
	// Issue request
	var response *applicationCommandData
	var err error
	if response, err = node.zwSendDataWaitForResponse(
		CommandClassClock, []uint8{clockGet}, clockReport); err != nil {
		return 0, 0, 0, err
	}

	data := response.Command.Data
	if len(data) != 2 {
		return 0, 0, 0, fmt.Errorf("Bad length: %d != %d", len(data), 2)
	}

	weekday := (data[0] >> 5) & 0x7
	hour := (data[0] & 0x1f)
	minute := data[1]

	// NOTE: Don't check weekday, because 0x7 guarantees [0, 7], and 0 is not
	// 		 really an error, since it means unknown

	if hour > 23 {
		return 0, 0, 0, fmt.Errorf("Bad hour %d > 23", hour)
	}
	if minute > 59 {
		return 0, 0, 0, fmt.Errorf("Bad minute %d > 59", minute)
	}

	return weekday, hour, minute, nil
}

// Set the time of the clock. Value values for weekday [1, 7], hour [0, 23],
// minute [0, 59]
func (node *Clock) Set(weekday uint8, hour uint8, minute uint8) error {
	// Unlike in Get, treat a 0 weekday as an error, because command will fail
	if weekday < 1 || weekday > 7 {
		return fmt.Errorf("Bad weekday not in range [1, 7]")
	}
	if hour > 23 {
		return fmt.Errorf("Bad hour not in range [0, 23]")
	}
	if minute > 59 {
		return fmt.Errorf("Bad minute not in range [0, 59]")
	}

	return node.zwSendDataRequest(CommandClassClock,
		[]uint8{clockSet, (weekday << 5) | (hour), minute})
}
