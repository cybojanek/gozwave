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
	basicCommandSet    uint8 = 0x01
	basicCommandGet          = 0x02
	basicCommandReport       = 0x03
)

// Basic information
type Basic struct {
	*Node
}

// GetBasic returns a Basic or nil object
func (node *Node) GetBasic() *Basic {
	node.mutex.Lock()
	defer node.mutex.Unlock()

	if node.supportsCommandClass(CommandClassBasic) {
		return &Basic{node}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

// Set the value
func (node *Basic) Set(value uint8) error {
	return node.zwSendDataRequest(CommandClassBasic,
		[]uint8{basicCommandSet, value})
}

// Get the value
func (node *Basic) Get() (uint8, error) {
	var response *ApplicationCommandData
	var err error

	if response, err = node.zwSendDataWaitForResponse(
		CommandClassBasic, []uint8{basicCommandGet},
		basicCommandReport, nil); err != nil {
		return 0, err
	}

	return node.ParseReport(response)
}

// IsReport checks if the report is a ParseReport
func (node *Basic) IsReport(report *ApplicationCommandData) bool {
	return report.Command.ID == basicCommandReport && len(report.Command.Data) == 1
}

// ParseReport of status
func (node *Basic) ParseReport(report *ApplicationCommandData) (uint8, error) {
	if report.Command.ClassID != CommandClassBasic {
		return 0, fmt.Errorf("Bad Report Command Class ID: 0x%02x != 0x%02x",
			report.Command.ClassID, CommandClassBasic)
	}

	if report.Command.ID != basicCommandReport {
		return 0, fmt.Errorf("Bad Report Command ID 0x%02x != 0x%02x",
			report.Command.ID, basicCommandReport)
	}

	data := report.Command.Data
	if len(data) != 1 {
		return 0, fmt.Errorf("Bad Report Data length %d != 1", len(data))
	}

	return data[0], nil
}

////////////////////////////////////////////////////////////////////////////////

// GetV2 the value
func (node *Basic) GetV2() (currentValue uint8, targetValue uint8, duration time.Duration, err error) {
	var response *ApplicationCommandData

	if response, err = node.zwSendDataWaitForResponse(
		CommandClassBasic, []uint8{basicCommandGet},
		basicCommandReport, nil); err != nil {
		return
	}

	return node.ParseReportV2(response)
}

// IsReportV2 checks if the report is a ParseReportV2
func (node *Basic) IsReportV2(report *ApplicationCommandData) bool {
	return report.Command.ID == basicCommandReport && len(report.Command.Data) == 3
}

// ParseReportV2 of status
func (node *Basic) ParseReportV2(report *ApplicationCommandData) (currentValue uint8, targetValue uint8, duration time.Duration, err error) {
	if report.Command.ClassID != CommandClassBasic {
		err = fmt.Errorf("Bad Report Command Class ID: 0x%02x != 0x%02x",
			report.Command.ClassID, CommandClassBasic)
		return
	}

	if report.Command.ID != basicCommandReport {
		err = fmt.Errorf("Bad Report Command ID 0x%02x != 0x%02x",
			report.Command.ID, basicCommandReport)
		return
	}

	data := report.Command.Data
	if len(data) != 3 {
		err = fmt.Errorf("Bad Report Data length %d != 3", len(data))
		return
	}

	currentValue = data[0]
	targetValue = data[1]
	duration = message.DecodeDuration(data[2])

	return
}
