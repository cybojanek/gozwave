package api

import (
	"encoding/binary"
	"github.com/cybojanek/gozwave/device"
	"github.com/cybojanek/gozwave/message"
	"log"
	"sync"
	// "time"
)

// Node information
type Node struct {
	ID uint8

	commandClasses        []uint8
	controlCommandClasses []uint8
	Listening             bool
	deviceClass           struct {
		basic    uint8
		generic  uint8
		specific uint8
	}
	manufacturer struct {
		id uint16
	}
	product struct {
		id  uint16
		typ uint16
	}

	api   *ZWAPI
	mutex sync.RWMutex
}

// Refresh the node information: Listening, DeviceClass, CommandClasses.
// This sends one blocking message, and multiple non-blocking messages, so not
// all fields will be immediately refreshed.
func (node *Node) Refresh() error {
	// Acquire exclusive lock, since we'll be updating fields
	node.mutex.Lock()
	defer node.mutex.Unlock()

	// Contact controller to get device description
	nodeProtocolInfo, err := node.api.zWGetNodeProtocolInfo(node.ID)
	if err != nil {
		return err
	}

	// Update fields
	node.Listening = nodeProtocolInfo.Capabilities.Listening
	node.deviceClass.basic = nodeProtocolInfo.DeviceClass.Basic
	node.deviceClass.generic = nodeProtocolInfo.DeviceClass.Generic
	node.deviceClass.specific = nodeProtocolInfo.DeviceClass.Specific

	// If it's a listening device, we can issue more commands to it
	if node.Listening {
		// Fill supported command classes
		if err := node.api.zWRequestNodeInfo(node.ID); err != nil {
			return err
		}

		// Fill manufacturer and product ids
		// ManufacturerSpecificCmd_Get = 0x04
		if err := node.api.zWSendData(node.ID, device.CommandClassManufacturerSpecific, []uint8{0x04}); err != nil {
			return err
		}
	}

	return nil
}

func (node *Node) applicationCommandHandler(command *message.ApplicationCommandHandler) {
	log.Printf("DEBUG applicationCommandHandler: node: %d command: %+v", node.ID, command)

	node.mutex.Lock()
	defer node.mutex.Unlock()

	if len(command.Body) < 2 {
		log.Printf("ERROR applicationCommandHandler: command is too short: %d", len(command.Body))
		return
	}

	commandClassID := command.Body[0]
	commandID := command.Body[1]
	commandData := command.Body[2:len(command.Body)]

	switch commandClassID {
	case device.CommandClassManufacturerSpecific:
		switch commandID {
		case 0x05: // ManufacturerSpecificCmd_Report
			if len(commandData) != 6 {
				log.Printf("ERROR MIAU")
			}
			node.manufacturer.id = binary.BigEndian.Uint16(commandData[0:2])
			node.product.typ = binary.BigEndian.Uint16(commandData[2:4])
			node.product.id = binary.BigEndian.Uint16(commandData[4:6])
		}
	}
}

func (node *Node) applicationUpdate(update *message.ZWApplicationUpdate) {
	log.Printf("DEBUG applicationUpdate: node: %d update: %+v", node.ID, update)

	node.mutex.Lock()
	defer node.mutex.Unlock()

	switch update.Status {
	case message.ZWApplicationUpdateStateReceived:
		// Update DeviceClass
		node.deviceClass.basic = update.Body[0]
		node.deviceClass.generic = update.Body[1]
		node.deviceClass.specific = update.Body[2]

		// Update CommandClasses
		node.commandClasses = []uint8{}
		node.controlCommandClasses = []uint8{}

		// NOTE: CommandClasses before CommandClassMark are those supported by
		//       the Node, while the CommandClasses after CommandClassMark are
		//       those which the Node can control
		afterMark := false
		for _, x := range update.Body[3:len(update.Body)] {
			if !afterMark && x == device.CommandClassMark {
				afterMark = true
			} else if !afterMark {
				node.commandClasses = append(node.commandClasses, x)
			} else { // afterMark
				node.controlCommandClasses = append(node.controlCommandClasses, x)
			}
		}
	}
}
