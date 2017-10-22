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
	alarmCommandGet      uint8 = 0x04
	alarmCommandReport         = 0x05
	alarmCommandSet            = 0x06
	alarmSupportedGet          = 0x07
	alarmSupportedReport       = 0x08
)

// Alarm Type
const (
	AlarmTypeSmoke           uint8 = 0x01
	AlarmTypeCarbonMonoxide        = 0x02
	AlarmTypeCarbonDioxide         = 0x03
	AlarmTypeHeat                  = 0x04
	AlarmTypeWater                 = 0x05
	AlarmTypeAccessControl         = 0x06
	AlarmTypeBurglar               = 0x07
	AlarmTypePowerManagement       = 0x08
	AlarmTypeSystem                = 0x09
	AlarmTypeEmergency             = 0x0a
	AlarmTypeClock                 = 0x0b
	AlarmTypeAppliance             = 0x0c
	AlarmTypeHomeHealth            = 0x0d
	AlarmTypeFirstSupported        = 0xff
)

// Alarm information
type Alarm struct {
	*Node
}

// GetAlarm returns a Alarm or nil object
func (node *Node) GetAlarm() *Alarm {
	node.mutex.Lock()
	defer node.mutex.Unlock()

	if node.supportsCommandClass(CommandClassAlarm) {
		return &Alarm{node}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

// Activate turns the alarm on
func (node *Alarm) Activate(alarmType uint8) error {
	return node.zwSendDataRequest(CommandClassAlarm,
		[]uint8{binarySwitchCommandSet, alarmType, 0xff})
}

// Deactivate turns the alarm off
func (node *Alarm) Deactivate(alarmType uint8) error {
	return node.zwSendDataRequest(CommandClassAlarm,
		[]uint8{binarySwitchCommandSet, alarmType, 0x00})
}

////////////////////////////////////////////////////////////////////////////////

// Get queries the node alarm status
func (node *Alarm) Get(alarmType uint8) (isActive bool, respAlarmType uint8, err error) {
	var response *ApplicationCommandData

	filter := func(response *ApplicationCommandData) bool {
		if len(response.Command.Data) < 1 {
			return false
		}
		return (alarmType == AlarmTypeFirstSupported) || (response.Command.Data[0] == alarmType)
	}

	if response, err = node.zwSendDataWaitForResponse(
		CommandClassAlarm, []uint8{alarmCommandGet, alarmType},
		alarmCommandReport, filter); err != nil {
		return
	}

	return node.ParseReport(response)
}

// IsReport checks if the report is a ParseReport
func (node *Alarm) IsReport(report *ApplicationCommandData) bool {
	return report.Command.ID == alarmCommandReport
}

// ParseReport of status
func (node *Alarm) ParseReport(report *ApplicationCommandData) (isActive bool, alarmType uint8, err error) {
	if report.Command.ClassID != CommandClassAlarm {
		err = fmt.Errorf("Bad Report Command Class ID: 0x%02x != 0x%02x",
			report.Command.ClassID, CommandClassAlarm)
		return
	}

	if report.Command.ID != alarmCommandReport {
		err = fmt.Errorf("Bad Report Command ID 0x%02x != 0x%02x",
			report.Command.ID, alarmCommandReport)
		return
	}

	data := report.Command.Data
	if len(data) < 2 {
		err = fmt.Errorf("Bad Report Data length %d < 2", len(data))
		return
	}

	alarmType = data[0]
	alarmLevel := data[1]

	// TODO: Better V2/Notification support
	if len(data) > 2 {
		if len(data) < 7 {
			err = fmt.Errorf("Bad Report Data length %d > 2 but %d < 7", len(data))
			return
		}

		alarmType = data[4]
		alarmLevel = data[3]
	}

	isActive = (alarmLevel != 0)
	return
}

////////////////////////////////////////////////////////////////////////////////

// GetSupportedAlarmTypes queries the alarm to get the list of supported alarm types
func (node *Alarm) GetSupportedAlarmTypes() (notificationOnly bool, alarmTypes []uint8, err error) {
	var response *ApplicationCommandData

	if response, err = node.zwSendDataWaitForResponse(
		CommandClassAlarm, []uint8{alarmSupportedGet}, alarmSupportedReport,
		nil); err != nil {
		return
	}

	data := response.Command.Data
	if len(data) < 1 {
		err = fmt.Errorf("Response too short %d < 1", len(data))
		return
	}

	notificationOnly = (data[0] & 0x80) == 0

	// Loop over bit mask
	if len(data) > 0 {
		alarmType := uint8(0)
		for _, b := range data[1:] {
			for i := uint32(0); i < 8; i++ {
				if (b & (1 << i)) != 0 {
					alarmTypes = append(alarmTypes, alarmType)
				}
				alarmType++
			}
		}
	}

	return
}
