// Package controller reads and writes packets to a ZWave USB Serial Controller.
// All methods are goroutine safe. The same controller instance can be opened
// and closed multiple times. Closing the controller will invalidate all ongoing
// requests and drop all buffered responses.
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
	"errors"
	"fmt"
	"github.com/cybojanek/gozwave/packet"
	"github.com/tarm/serial"
	"log"
	"sync"
	"time"
)

// A request to the controller
type controllerRequest struct {
	Request  *packet.Packet // Request Packet
	Response *packet.Packet // Response Packet or nil if Err is not nil
	Err      error          // Contains processing error
	Chan     chan int       // Channel to notify on request completion
}

// Controller information and state
type Controller struct {
	DevicePath string // Path USB serial device

	mutex            sync.Mutex                                      // Controller mutex
	callbackMutex    sync.Mutex                                      // Callback mutex
	callbackChannels []map[*chan *packet.Packet]*chan *packet.Packet // Callback channels
	callbackDefaults map[*chan *packet.Packet]*chan *packet.Packet   // Default callbacks
	serial           *serial.Port                                    // Serial port connection
	responses        chan *packet.Packet                             // Channel for packets read from serial
	requests         chan *controllerRequest                         // Channel for outgoing requests
	stopResponses    chan int                                        // Exit signal channel for doResponses
	stopRequests     chan int                                        // Exit signal channel for doRequests
	stoppedResponses chan int                                        // Exit confirmation channel for doResponses
	stoppedRequests  chan int                                        // Exit confirmation channel for doRequests
}

// Maximum number of request send errors to retry
const maxRequestRetryCount = 5

// Maximum time to wait for a request ACK timeout per attempt
const requestACKTimeout = (10 * time.Second)

// Maximum number of response read errors to retry
const maxResponseRetryCount = 5

// Maximum time to wait for a response timeout per attempt
const responseTimeout = (10 * time.Second)

// Serial port read timeout for non-blocking mode
const serialPortReadTimeout = (1 * time.Second)

var ackBytes = []uint8{packet.PacketPreambleACK, '\n'}

// Write all bytes to the serial device
func (controller *Controller) writeFully(b []byte) error {
	log.Printf("DEBUG writeFully(%v)", b)

	written := 0
	for written < len(b) {
		n, err := controller.serial.Write(b[written:])
		if err != nil {
			log.Printf("ERROR writeFully error: %v", err)
			return err
		}
		written += n
	}

	log.Printf("DEBUG writeFully EXIT")
	return nil
}

// Read from serial device until array is full
func (controller *Controller) readFully(b []byte) error {
	log.Printf("DEBUG readFully(%d)", len(b))

	read := 0
	for read < len(b) {
		n, err := controller.serial.Read(b[read:])
		if err != nil {
			log.Printf("ERROR readFully error: %v", err)
			return err
		}
		read += n
	}

	log.Printf("DEBUG readFully EXIT: %v", b)
	return nil
}

// routeRespones routes a packet to a callback channel
func (controller *Controller) routeReponse(packet *packet.Packet) {
	controller.callbackMutex.Lock()
	defer func() {
		controller.callbackMutex.Unlock()
	}()

	// Get channels by message type
	callbackChannels := controller.callbackChannels[packet.MessageType]

	if len(callbackChannels) == 0 {
		// If there is no message type channels then send out to default callbacks
		callbackChannels = controller.callbackDefaults
	}

	for _, x := range callbackChannels {
		// Call in goroutine to avoid deadlock in Controller
		go func() {
			*x <- packet.Copy()
		}()
	}
}

// Read from serial device and forward parsed packets to controller.responses
func (controller *Controller) doResponses() {
	// Parser and buffer
	parser := packet.Parser{}
	buffer := make([]byte, 512)

	for {
		// Read blocking with timeout
		// FIXME: currently error ignored following serial library guidelines
		if n, _ := controller.serial.Read(buffer); n > 0 {
			// Log received bytes
			log.Printf("DEBUG doResponses bytes: %v", buffer[0:n])

			// Parse all bytes
			for _, b := range buffer[0:n] {
				if p, err := parser.Parse(b); err != nil {
					// Log error
					log.Printf("ERROR failed parsing response: %v", err)
				} else if p != nil {
					// Forward parsed packet
					controller.responses <- p
				}
			}
		} else if n < 0 {
			log.Printf("ERROR doResponses bad n value: %d", n)
		} else {
			log.Printf("DEBUG doResponses waiting...")
		}

		select {
		case <-controller.stopResponses:
			// Received exit signal
			log.Printf("DEBUG doResponses EXIT")
			controller.stoppedResponses <- 0
			return

		default:
			// Default for non-blocking continue
		}
	}
}

// Process controller.requests and controller.responses
func (controller *Controller) doRequests() {
	for {
		select {

		case response := <-controller.responses:
			// Handle packets not generated by requests, i.e. status reports
			// from switches after they are toggled on or off
			log.Printf("DEBUG doRequests response Packet: %v", response)
			if response == nil {
				log.Printf("ERROR doRequests response got nil Packet")
			} else {
				switch response.Preamble {
				case packet.PacketPreambleSOF:
					// Immediately ACK back to SOF messages
					if err := controller.writeFully(ackBytes); err != nil {
						log.Printf("ERROR doRequests response ACK writeFully error: %v",
							err)
					}
					controller.routeReponse(response)

				case packet.PacketPreambleACK, packet.PacketPreambleCAN, packet.PacketPreambleNAK:
					log.Printf("ERROR doRequests response got unexpected Preamble: 0x%02x",
						response.Preamble)
				default:
					log.Printf("ERROR doRequests response got unknown Preamble: 0x%02x",
						response.Preamble)
				}
			}

		case request := <-controller.requests:
			// Handle request packets
			log.Printf("DEBUG doRequests request Packet: %v", request.Request)

			requestBytes, reqErr := request.Request.Bytes()
			if reqErr != nil {
				log.Printf("ERROR doRequests request Bytes() error: %v", reqErr)
				request.Err = fmt.Errorf("Failed to Bytes() packet: %v", reqErr)
				request.Chan <- 0
				break
			}
			requestBytes = append(requestBytes, '\n')

			// Send out request
			gotACK := false
			for attempt := 0; attempt < maxRequestRetryCount && !gotACK; {
				// Write Packet
				controller.writeFully(requestBytes)

				select {

				case response := <-controller.responses:
					switch response.Preamble {
					case packet.PacketPreambleSOF:
						// Recieved a packet not generatd by this request
						// Immediately ACK back to SOF messages
						if err := controller.writeFully(ackBytes); err != nil {
							log.Printf("ERROR doRequests request ACK writeFully error: %v",
								err)
						}
						// Try to route it anyway...
						controller.routeReponse(response)

					case packet.PacketPreambleCAN:
						// Recieved a message while trying to send ours - not
						// an error. If we have lots of nodes, this could be
						// frequent, so don't count this as an error
						log.Printf("ERROR doRequests request got unexpected CAN")

					case packet.PacketPreambleACK:
						// Got what we were expecting
						log.Printf("DEBUG doRequests request got expected ACK")
						gotACK = true

					case packet.PacketPreambleNAK:
						// FIXME: what does this mean?
						log.Printf("ERROR doRequests request got unexpected NAK")
						attempt++

					default:
						log.Printf("ERROR doRequests request got unknown Preamble: 0x%02x",
							response.Preamble)
						attempt++
					}

				case <-time.After(requestACKTimeout):
					log.Printf("ERROR doRequests request timeout out after %v waiting for ACK",
						requestACKTimeout)
					attempt++

				case <-controller.stopRequests:
					log.Printf("INFO Dropping doRequests request due to close: %v",
						request.Request)

					request.Err = fmt.Errorf("Controller closed")
					request.Chan <- 0

					log.Printf("DEBUG doRequests EXIT")
					controller.stoppedRequests <- 0

					return
				}
			}

			// Check if we got an ACK
			if !gotACK {
				log.Printf("ERROR doRequests request failed after %d attempts",
					maxRequestRetryCount)
				request.Err = errors.New("Failed to send request")
				request.Chan <- 0
				break
			}

			// Await response
			gotResponse := false
			for attempt := 0; attempt < maxResponseRetryCount && !gotResponse; {

				select {

				case response := <-controller.responses:
					switch response.Preamble {
					case packet.PacketPreambleSOF:
						// Recieved our response!
						// Immediately ACK back to SOF messages
						if err := controller.writeFully(ackBytes); err != nil {
							log.Printf("ERROR doRequests request response "+
								"ACK writeFully error: %v", err)
						}

						if request.Request.MessageType == response.MessageType {
							// TODO: is this always true?
							log.Printf("DEBUG doRequests request reponse %v", response)
							request.Response = response
							gotResponse = true
							request.Chan <- 0
						} else {
							log.Printf("ERROR doRequests request response "+
								"expected MessageType: 0x%02x got 0x%02x, "+
								"will try to route Response: %v",
								request.Request.MessageType,
								response.MessageType, response)
							attempt++
							controller.routeReponse(response)
						}

					case packet.PacketPreambleCAN:
						// FIXME: how to handle this
						log.Printf("ERROR doRequests request response got unexpected CAN")
						attempt++

					case packet.PacketPreambleACK:
						// FIXME: how to handle this
						log.Printf("DEBUG doRequests request response got unexpected ACK")
						attempt++

					case packet.PacketPreambleNAK:
						// FIXME: how to handle this
						log.Printf("ERROR doRequests request response got unexpected NAK")
						attempt++

					default:
						log.Printf("ERROR doRequests request response got unknown Preamble: 0x%02x",
							response.Preamble)
						attempt++
					}

				case <-time.After(responseTimeout):
					log.Printf("ERROR doRequests request response timeout out after %v waiting for SOF",
						responseTimeout)
					attempt++

				case <-controller.stopRequests:
					log.Printf("INFO Dropping doRequests request due to close: %v",
						request.Request)

					request.Err = fmt.Errorf("Controller closed")
					request.Chan <- 0

					log.Printf("DEBUG doRequests EXIT")
					controller.stoppedRequests <- 0

					return
				}
			}

			if !gotResponse {
				log.Printf("ERROR doRequests request response failed after %d attempts",
					maxResponseRetryCount)
				request.Err = errors.New("Failed to get request response")
				request.Chan <- 0
			}

		case <-controller.stopRequests:
			log.Printf("DEBUG doRequests EXIT")
			controller.stoppedRequests <- 0
			return
		}
	}
}

// BlockingRequest issues a request and awaits a response
func (controller *Controller) BlockingRequest(request *packet.Packet) (*packet.Packet, error) {
	log.Printf("DEBUG BlockingRequest(%v)", request)

	if request.Preamble != packet.PacketPreambleSOF {
		return nil, fmt.Errorf("Packet has non SOF Preamble: 0x%02x",
			request.Preamble)
	}

	if request.PacketType != packet.PacketTypeRequest {
		return nil, fmt.Errorf("Packet has non Request Packet Type: 0x%02x",
			request.PacketType)
	}

	// Check before with lock to avoid deadlock on pre first open nil channel
	if !controller.IsOpen() {
		return nil, fmt.Errorf("Controller is not open")
	}

	// Send request
	// NOTE: make chan 1 to not block controller routine
	controllerRequest := controllerRequest{Request: request, Chan: make(chan int, 1)}
	controller.requests <- &controllerRequest

	controller.mutex.Lock()
	if !controller.isOpen() {
		// Error on all requests, ours might be there too
		// NOTE: this will only loop indefinitely if there is an unending
		//       stream of requests...
		for {
			finished := false

			select {

			case request := <-controller.requests:
				log.Printf("INFO Dropping BlockingRequest request due to close: %v",
					request)

				request.Err = fmt.Errorf("Controller closed")
				request.Chan <- 0

			default:
				finished = true
			}

			if finished {
				break
			}
		}
	}
	controller.mutex.Unlock()

	// Await reply
	<-controllerRequest.Chan
	return controllerRequest.Response, controllerRequest.Err
}

// AddCallbackChannel adds the channel to the callback list for the message type
func (controller *Controller) AddCallbackChannel(messageType uint8, channel *chan *packet.Packet) {
	controller.callbackMutex.Lock()
	defer func() {
		controller.callbackMutex.Unlock()
	}()

	controller.callbackChannels[messageType][channel] = channel
}

// RemoveCallbackChannel removes the channel from the callback list for the message type
func (controller *Controller) RemoveCallbackChannel(messageType uint8, channel *chan *packet.Packet) {
	controller.callbackMutex.Lock()
	defer func() {
		controller.callbackMutex.Unlock()
	}()

	delete(controller.callbackChannels[messageType], channel)
}

// AddDefaultCallbackChannel adds the channel to the default callback list
func (controller *Controller) AddDefaultCallbackChannel(channel *chan *packet.Packet) {
	controller.callbackMutex.Lock()
	defer func() {
		controller.callbackMutex.Unlock()
	}()

	controller.callbackDefaults[channel] = channel
}

// RemoveDefaultCallbackChannel removes the channel from the default callback list
func (controller *Controller) RemoveDefaultCallbackChannel(channel *chan *packet.Packet) {
	controller.callbackMutex.Lock()
	defer func() {
		controller.callbackMutex.Unlock()
	}()

	delete(controller.callbackDefaults, channel)
}

// IsOpen checks if Controller is open
func (controller *Controller) IsOpen() bool {
	controller.mutex.Lock()
	defer func() {
		controller.mutex.Unlock()
	}()
	return controller.serial != nil
}

// isOpen is an private function that does not acquire the controller mutex
func (controller *Controller) isOpen() bool {
	return controller.serial != nil
}

// Open controller
func (controller *Controller) Open() error {
	controller.mutex.Lock()
	defer func() {
		controller.mutex.Unlock()
	}()

	if controller.isOpen() {
		return nil
	}

	// rtscts True, dsrdtr True
	c := &serial.Config{Name: controller.DevicePath, Baud: 115200,
		ReadTimeout: serialPortReadTimeout}

	s, err := serial.OpenPort(c)
	if err != nil {
		return err
	}
	controller.serial = s

	controller.serial.Flush()

	// Create an array of callback channel maps for each possible message type
	// We use a map for easy add/remove
	if controller.callbackChannels == nil {
		controller.callbackChannels = make([]map[*chan *packet.Packet]*chan *packet.Packet, 256)
		for i := 0; i < 256; i++ {
			controller.callbackChannels[i] = make(map[*chan *packet.Packet]*chan *packet.Packet)
		}
		controller.callbackDefaults = make(map[*chan *packet.Packet]*chan *packet.Packet)

		controller.responses = make(chan *packet.Packet)
		// 1 to avoid deadlock on closed submit
		controller.requests = make(chan *controllerRequest, 1)
		controller.stopRequests = make(chan int)
		controller.stoppedRequests = make(chan int)
		controller.stopResponses = make(chan int)
		controller.stoppedResponses = make(chan int)
	}

	go controller.doRequests()
	go controller.doResponses()

	return nil
}

// Close controller
func (controller *Controller) Close() error {
	controller.mutex.Lock()
	defer func() {
		controller.mutex.Unlock()
	}()

	var err error

	if controller.serial != nil {
		// doRequests will always stop if triggered with stopRequqests
		controller.stopRequests <- 0
		<-controller.stoppedRequests

		// doResponses might block on controller.responses, which we will purge
		log.Printf("Sending controller.stopResponses!")
		controller.stopResponses <- 0

		// Purge all requests and responses
		for {
			finished := false

			select {

			case request := <-controller.requests:
				log.Printf("INFO Dropping Close request: %v", request)

				request.Err = fmt.Errorf("Controller closed")
				request.Chan <- 0

			case response := <-controller.responses:
				log.Printf("INFO Dropping Close response: %v", response)

			default:
				finished = true
			}

			if finished {
				break
			}
		}

		<-controller.stoppedResponses

		// Close after doReponses exits
		err = controller.serial.Close()
	}

	controller.serial = nil

	return err
}
