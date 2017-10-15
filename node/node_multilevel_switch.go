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
	"github.com/cybojanek/gozwave/message"
	"time"
)

const (
	multiLevelSwitchCommandSet              uint8 = 0x01
	multiLevelSwitchCommandGet                    = 0x02
	multiLevelSwitchCommandReport                 = 0x03
	multiLevelSwitchCommandStartLevelChange       = 0x04
	multiLevelSwitchCommandStopLevelChange        = 0x05
)

// MultiLevelSwitch information
type MultiLevelSwitch struct {
	*Node
}

// GetMultiLevelSwitch returns a MultiLevelSwitch or nil object
func (node *Node) GetMultiLevelSwitch() *MultiLevelSwitch {
	node.mutex.Lock()
	defer node.mutex.Unlock()

	if node.supportsCommandClass(CommandClassMultiLevelSwitch) {
		return &MultiLevelSwitch{node}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

// On turns the switch on to the most recent non-zero level
func (node *MultiLevelSwitch) On() error {
	return node.zwSendDataRequest(CommandClassMultiLevelSwitch,
		[]uint8{multiLevelSwitchCommandSet, 0xff})
}

// Off turns the switch off
func (node *MultiLevelSwitch) Off() error {
	return node.zwSendDataRequest(CommandClassMultiLevelSwitch,
		[]uint8{multiLevelSwitchCommandSet, 0x00})
}

// IsOn queries the switch to check current status
func (node *MultiLevelSwitch) IsOn() (bool, error) {
	if response, err := node.zwSendDataWaitForResponse(
		CommandClassMultiLevelSwitch, []uint8{multiLevelSwitchCommandGet},
		multiLevelSwitchCommandReport); err != nil {
		return false, err
	} else if len(response.Command.Data) != 1 {
		return false, fmt.Errorf("Bad response")
	} else {
		return response.Command.Data[0] != 0, nil
	}
}

////////////////////////////////////////////////////////////////////////////////

// Get queries the switch to check current value
func (node *MultiLevelSwitch) Get() (uint8, error) {
	if response, err := node.zwSendDataWaitForResponse(
		CommandClassMultiLevelSwitch, []uint8{multiLevelSwitchCommandGet},
		multiLevelSwitchCommandReport); err != nil {
		return 0, err
	} else if len(response.Command.Data) != 1 {
		return 0, fmt.Errorf("Bad response")
	} else {
		return response.Command.Data[0], nil
	}
}

// Set sets the level to the requested value, which must be in the range
// of [0, 99] or 0xff, where 255 is the most recent non-zero level
func (node *MultiLevelSwitch) Set(value uint8) error {
	if value > 99 && value < 0xff {
		return fmt.Errorf("Value must be in range [0, 99] or 255")
	}
	return node.zwSendDataRequest(CommandClassMultiLevelSwitch,
		[]uint8{multiLevelSwitchCommandSet, value})
}

////////////////////////////////////////////////////////////////////////////////

// Start a level change
func (node *MultiLevelSwitch) Start(up bool, ignoreStart bool, start uint8) error {
	if start > 99 && start < 0xff {
		return fmt.Errorf("Start must be in range [0, 99] or 255")
	}
	flags := uint8(0)
	if up {
		flags |= (1 << 6)
	}
	if ignoreStart {
		flags |= (1 << 5)
	}
	return node.zwSendDataRequest(CommandClassMultiLevelSwitch,
		[]uint8{multiLevelSwitchCommandStartLevelChange, flags, start})
}

// Stop an ongoing level change
func (node *MultiLevelSwitch) Stop() error {
	return node.zwSendDataRequest(CommandClassMultiLevelSwitch,
		[]uint8{multiLevelSwitchCommandStopLevelChange})
}

////////////////////////////////////////////////////////////////////////////////

// SetV2 sets the level to the requested value, which must be in the range
// of [0, 99] or 0xff, where 255 is the most recent non-zero level, and duration
// must be either [0, 127] seconds or [1, 127] minutes
func (node *MultiLevelSwitch) SetV2(value uint8, duration time.Duration) error {
	if value > 99 && value < 0xff {
		return fmt.Errorf("Value must be in range [0, 99] or 255")
	}

	// Calculate duration
	var durationByte uint8
	var err error
	if durationByte, err = message.EncodeDuration(duration); err != nil {
		return err
	}
	return node.zwSendDataRequest(CommandClassMultiLevelSwitch,
		[]uint8{multiLevelSwitchCommandStartLevelChange, value, durationByte})
}

// StartV2 a level change with a duration to the requested value, which must be
// in the range of [0, 99] or 0xff, where 255 is the most recent non-zero level,
// and duration must be either [0, 127] seconds or [1, 127] minutes
func (node *MultiLevelSwitch) StartV2(up bool, ignoreStart bool, start uint8, duration time.Duration) error {
	if start > 99 && start < 0xff {
		return fmt.Errorf("Start must be in range [0, 99] or 255")
	}
	flags := uint8(0)
	if up {
		flags |= (1 << 6)
	}
	if ignoreStart {
		flags |= (1 << 5)
	}

	// Calculate duration
	var durationByte uint8
	var err error
	if durationByte, err = message.EncodeDuration(duration); err != nil {
		return err
	}
	return node.zwSendDataRequest(CommandClassMultiLevelSwitch,
		[]uint8{multiLevelSwitchCommandStartLevelChange, flags, start, durationByte})
}
