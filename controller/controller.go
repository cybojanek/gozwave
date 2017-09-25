// Package controller reads and writes packets to a ZWave USB Serial Controller.
// All public methods are goroutine safe. The same controller instance can be
// opened and closed multiple times. Closing the controller will invalidate all
// ongoing requests and drop all buffered responses.
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
	"github.com/cybojanek/gozwave/message"
	"github.com/cybojanek/gozwave/packet"
	"github.com/tarm/serial"
	"log"
	"math/rand"
	"sync"
	"time"
)

// Controller processes a ZWave packet and returns a response
type Controller interface {
	DoRequest(request *packet.Packet) (*packet.Packet, error)
}

// SerialController information and state
type SerialController struct {
	DevicePath   string // Path USB serial device
	DebugLogging bool   // Toggle DEBUG logging

	lastCallbackID   uint8                   // Next ZWSendData callback id
	mutex            sync.Mutex              // SerialController mutex
	callbackChannel  chan *packet.Packet     // Callback channel
	serial           *serial.Port            // Serial port connection
	responses        chan *packet.Packet     // Channel for packets read from serial
	requests         chan *controllerRequest // Channel for outgoing requests
	stopResponses    chan int                // Exit signal channel for doResponses
	stopRequests     chan int                // Exit signal channel for doRequests
	stoppedResponses chan int                // Exit confirmation channel for doResponses
	stoppedRequests  chan int                // Exit confirmation channel for doRequests
}

// A request to the controller. Used only within serial constroller
type controllerRequest struct {
	Request  *packet.Packet // Request Packet
	Response *packet.Packet // Response Packet or nil if Err is not nil
	Err      error          // Contains processing error
	Chan     chan int       // Channel to notify on request completion
}

// TODO: check constraints on this
const callbackIDMin = 0x0a + 1
const callbackIDMax = 0x7f

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
var nakBytes = []uint8{packet.PacketPreambleNAK, '\n'}

////////////////////////////////////////////////////////////////////////////////

// isOpen is an private function that does not acquire the controller mutex.
// NOTE: not goroutine safe, caller must hold controller.mutex
func (controller *SerialController) isOpen() bool {
	return controller.serial != nil
}

// Open controller. goroutine safe.
func (controller *SerialController) Open() error {
	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	if controller.isOpen() {
		return nil
	}

	c := &serial.Config{Name: controller.DevicePath, Baud: 115200,
		ReadTimeout: serialPortReadTimeout}

	s, err := serial.OpenPort(c)
	if err != nil {
		return err
	}
	controller.serial = s

	controller.serial.Flush()

	if controller.responses == nil {
		controller.responses = make(chan *packet.Packet)
		// 1 to avoid deadlock on closed submit
		controller.requests = make(chan *controllerRequest, 1)
		controller.stopRequests = make(chan int)
		controller.stoppedRequests = make(chan int)
		controller.stopResponses = make(chan int)
		controller.stoppedResponses = make(chan int)
	}

	// On startup choose a random starting callbackID
	rand.Seed(time.Now().Unix())
	controller.lastCallbackID = uint8(rand.Int31n(callbackIDMax-callbackIDMin+1) + callbackIDMin)

	go controller.doRequests()
	go controller.doResponses()

	return nil
}

// Close controller. goroutine safe.
func (controller *SerialController) Close() error {
	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	if !controller.isOpen() {
		return nil
	}

	// doRequests will always stop if triggered with stopRequqests
	controller.stopRequests <- 0
	<-controller.stoppedRequests

	// doResponses might block on sending to controller.responses, so don't
	// check for stoppedResponses until after purging all channels
	controller.stopResponses <- 0

	// Purge all requests and responses
loop:
	for {
		select {

		case request := <-controller.requests:
			request.Err = fmt.Errorf("Controller closed")
			request.Chan <- 0

		case <-controller.responses:
			// Pass and drop

		default:
			break loop
		}
	}

	// Now safe to expect response
	<-controller.stoppedResponses

	// Close after doReponses exits
	err := controller.serial.Close()

	controller.serial = nil

	return err
}

////////////////////////////////////////////////////////////////////////////////

// DoRequest issues a request and awaits a response. goroutine safe.
func (controller *SerialController) DoRequest(request *packet.Packet) (*packet.Packet, error) {
	if request.Preamble != packet.PacketPreambleSOF {
		return nil, fmt.Errorf("Packet has non SOF Preamble: 0x%02x",
			request.Preamble)
	}

	if request.PacketType != packet.PacketTypeRequest {
		return nil, fmt.Errorf("Packet has non Request Packet Type: 0x%02x",
			request.PacketType)
	}

	// Check before with lock to avoid deadlock on pre first open nil channel
	controller.mutex.Lock()
	if !controller.isOpen() {
		controller.mutex.Unlock()
		return nil, fmt.Errorf("Controller is not open")
	}
	controller.mutex.Unlock()

	// Send request
	// NOTE: make chan 1 to not block controller routine
	controllerRequest := controllerRequest{Request: request, Chan: make(chan int, 1)}
	controller.requests <- &controllerRequest

	controller.mutex.Lock()
	if !controller.isOpen() {
		// Error on all requests, ours might be there too
		// NOTE: this will only loop indefinitely if there is an unending
		//       stream of new requests...
	loop:
		for {
			select {

			case request := <-controller.requests:
				request.Err = fmt.Errorf("Controller closed")
				request.Chan <- 0

			default:
				break loop
			}
		}
	}
	controller.mutex.Unlock()

	// Await reply
	<-controllerRequest.Chan
	return controllerRequest.Response, controllerRequest.Err
}

// SetCallbackChannel set the channel to the callback list, can be null.
// goroutine safe.
func (controller *SerialController) SetCallbackChannel(channel chan *packet.Packet) {
	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	controller.callbackChannel = channel
}

////////////////////////////////////////////////////////////////////////////////

// Read from serial device and forward parsed packets to controller.responses
func (controller *SerialController) doResponses() {
	// Parser and buffer
	parser := packet.Parser{}
	buffer := make([]byte, 512)

	for {
		// Read blocking with timeout
		// FIXME: currently error ignored following serial library guidelines
		if n, _ := controller.serial.Read(buffer); n > 0 {
			// Log received bytes
			if controller.DebugLogging {
				log.Printf("DEBUG doResponses bytes: %v", buffer[0:n])
			}

			// Parse all bytes
			for _, b := range buffer[0:n] {
				if p, err := parser.Parse(b); err != nil {
					// Log error
					log.Printf("ERROR failed parsing response: %v", err)
					// Reply with NAK, however its not safe to do that from
					// here, so send a nil packet to controller.responses
					// FIXME: implement a better signaling method
					controller.responses <- nil
				} else if p != nil {
					// Forward parsed packet
					controller.responses <- p
				}
			}
		} else if n < 0 {
			log.Printf("ERROR doResponses bad n value: %d", n)
		}

		select {
		case <-controller.stopResponses:
			// Received exit signal
			controller.stoppedResponses <- 0
			return

		default:
			// Default for non-blocking continue
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

// Write all bytes to the serial device
// Assumptions: called only from doRequests
func (controller *SerialController) writeFully(b []byte) error {
	written := 0
	for written < len(b) {
		n, err := controller.serial.Write(b[written:])
		if err != nil {
			log.Printf("ERROR writeFully error: %v", err)
			return err
		}
		written += n
	}
	return nil
}

// Get the callback ID to use for the next request
// Assumptions: called only from doRequests
func (controller *SerialController) getZWaveCallbackID() uint8 {
	// Reset if to min if not within liimits
	if controller.lastCallbackID >= callbackIDMax || controller.lastCallbackID < callbackIDMin {
		controller.lastCallbackID = callbackIDMin
	}

	ret := controller.lastCallbackID
	controller.lastCallbackID++
	return ret
}

// sendToCallback sends a packet to a callback channel
// NOTE: goroutine safe, acquires controller lock
func (controller *SerialController) sendToCallback(packet *packet.Packet) {
	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	// NOTE: extract to local variable to not refernce controller in goroutine
	channel := controller.callbackChannel
	if channel != nil {
		// Call in goroutine to avoid deadlock in Controller
		go func() {
			channel <- packet.Copy()
		}()
	}
}

// Process controller.requests and controller.responses
func (controller *SerialController) doRequests() {
	// Start with NAK bytes to reset stream
	if err := controller.writeFully(nakBytes); err != nil {
		log.Printf("ERROR doRequests initial NAK error: %v", err)
	}

	for {
		select {

		case response := <-controller.responses:
			// Handle packets not generated by requests, i.e. status reports
			// from switches after they are toggled on or off, or by out of
			// sync or timed out requests
			if controller.DebugLogging {
				log.Printf("DEBUG doRequests response Packet: %v", response)
			}

			if response == nil {
				// Send a NAK
				if err := controller.writeFully(nakBytes); err != nil {
					log.Printf("ERROR doRequests response NAK writeFully error: %v", err)
				}
				continue
			}

			switch response.Preamble {
			case packet.PacketPreambleSOF:
				// Immediately ACK back to SOF messages
				if err := controller.writeFully(ackBytes); err != nil {
					log.Printf("ERROR doRequests response ACK writeFully error: %v", err)
				}
				controller.sendToCallback(response)

			case packet.PacketPreambleACK, packet.PacketPreambleCAN, packet.PacketPreambleNAK:
				log.Printf("DEBUG doRequests response got ACK, CAN, or NAK: 0x%02x",
					response.Preamble)

			default:
				log.Printf("ERROR doRequests response got unknown Preamble: 0x%02x",
					response.Preamble)
			}

		case request := <-controller.requests:
			// Handle request packets
			if controller.DebugLogging {
				log.Printf("DEBUG doRequests request Packet: %v", request.Request)
			}

			// For ZWSendData, we need to inspect and potentially inject a
			// random callback id. This is ugly, since we're mixing protocol
			// layers, but at least we can transparently handle this.
			var callbackID uint8
			if request.Request.MessageType == message.MessageTypeZWSendData {
				callbackID = controller.getZWaveCallbackID()

				body := request.Request.Body
				// Body: | NODE_ID | LENGTH_OF_PAYLOAD + 1 | COMMAND_CLASS |
				//       | PAYLOAD | TRANSMIT_OPTIONS | CALLBACK_ID |
				if len(body) < 4 {
					request.Err = fmt.Errorf("ZWSendData request is too small")
					request.Chan <- 0
					break
				}
				payloadLength := body[1]
				if int(payloadLength)+4 == len(body) {
					// Callback id is last byte - error, not allowed
					request.Err = fmt.Errorf("Specifying a custom ZWSendData " +
						"CallbackID is not allowed")
					request.Chan <- 0
					break
				} else if int(payloadLength)+3 == len(body) {
					// No callback id
					request.Request.Body = append(request.Request.Body, callbackID)
				} else {
					request.Err = fmt.Errorf("ZWSendData request has unexpected length")
					request.Chan <- 0
					break
				}

				if controller.DebugLogging {
					log.Printf("DEBUG doRequests request modified Packet: %v", request.Request)
				}
			}

			requestBytes, reqErr := request.Request.Bytes()
			if reqErr != nil {
				request.Err = fmt.Errorf("Failed to Bytes() packet: %v", reqErr)
				request.Chan <- 0
				break
			}
			// Serial port requires packets to be newline terminated
			requestBytes = append(requestBytes, '\n')

			// Send out request

			resend := true
			gotACK := false
			for attempt := 0; attempt < maxRequestRetryCount && !gotACK; {

				// Write Packet
				if resend {
					if err := controller.writeFully(requestBytes); err != nil {
						request.Err = fmt.Errorf("Failed to write bytes to serial device %v", err)
						request.Chan <- 0
						break
					}
					resend = false
				}

				select {

				case response := <-controller.responses:
					if controller.DebugLogging {
						log.Printf("DEBUG doRequests request response A: %v", response)
					}

					if response == nil {
						// Send a NAK
						log.Printf("ERROR doRequests request unexpected NAK request")
						if err := controller.writeFully(nakBytes); err != nil {
							log.Printf("ERROR doRequests request NAK writeFully error: %v", err)
						}
						attempt++
						continue
					}

					switch response.Preamble {
					case packet.PacketPreambleSOF:
						// Recieved a packet not generatd by this request
						// Immediately ACK back to SOF messages
						if err := controller.writeFully(ackBytes); err != nil {
							log.Printf("ERROR doRequests request ACK writeFully error: %v", err)
						}
						// Try to route it anyway...
						controller.sendToCallback(response)

					case packet.PacketPreambleCAN:
						// Recieved a message while trying to send ours - not
						// an error. If we have lots of nodes, this could be
						// frequent, so don't count this as an error
						log.Printf("ERROR doRequests request got unexpected CAN")
						resend = true

					case packet.PacketPreambleACK:
						// Got what we were expecting
						gotACK = true

					case packet.PacketPreambleNAK:
						// Error in packet transmission, or our packet encoding
						// is bad...
						log.Printf("ERROR doRequests request got unexpected NAK")
						attempt++
						resend = true

					default:
						// Parser should never let this happen
						log.Printf("ERROR doRequests request got unknown Preamble: 0x%02x",
							response.Preamble)
						attempt++
					}

				case <-time.After(requestACKTimeout):
					log.Printf("ERROR doRequests request timeout out after %v waiting for ACK",
						requestACKTimeout)
					attempt++
					resend = true

				case <-controller.stopRequests:
					request.Err = fmt.Errorf("Controller closed")
					request.Chan <- 0
					controller.stoppedRequests <- 0
					return
				}
			}

			// Check if we got an ACK
			if !gotACK {
				request.Err = errors.New("Failed to send request")
				request.Chan <- 0
				break
			}

			responseCount := 0
		wait_for_response:
			// Await response
			gotResponse := false
			for attempt := 0; attempt < maxResponseRetryCount && !gotResponse; {

				select {

				case response := <-controller.responses:

					if controller.DebugLogging {
						log.Printf("DEBUG doRequests request response B: %v", response)
					}

					if response == nil {
						// Send a NAK
						if err := controller.writeFully(nakBytes); err != nil {
							log.Printf("ERROR doRequests request response "+
								"NAK writeFully error: %v", err)
						}
						attempt++
						continue
					}

					switch response.Preamble {
					case packet.PacketPreambleSOF:
						// Recieved our response!
						// Immediately ACK back to SOF messages
						if err := controller.writeFully(ackBytes); err != nil {
							log.Printf("ERROR doRequests request response "+
								"ACK writeFully error: %v", err)
						}

						if request.Request.MessageType != response.MessageType {
							if controller.DebugLogging {
								log.Printf("DEBUG doRequests request response "+
									"expected MessageType: 0x%02x got 0x%02x, "+
									"will try to route Response: %v",
									request.Request.MessageType,
									response.MessageType, response)
							}
							attempt++
							controller.sendToCallback(response)
							continue
						}

						if response.MessageType == message.MessageTypeZWSendData {
							if responseCount == 0 {
								// This is the first ZWSendData response
								if len(response.Body) != 1 {
									log.Printf("ERROR doRequests request response "+
										"ZWSendData reply too long: %d", len(response.Body))
									// Try to route it anyways
									controller.sendToCallback(response)
									attempt++
									continue
								} else if response.Body[0] != 1 {
									log.Printf("ERROR doRequests request response "+
										"ZWSendData reply not TransmitCompleteOK (0x00): 0x%02x",
										response.Body[0])
									attempt++
									continue
								} else {
									// This is just the 1 byte response confirming
									// our request. We need to wait for one more
									// response with the final data
									attempt = 0
									responseCount = 1
									goto wait_for_response
								}
							} else {
								// Check for matching callback id
								if len(response.Body) != 4 {
									log.Printf("ERROR doRequests request response "+
										"ZWSendData reply not == 4: 0x%02x", len(response.Body))
									attempt++
									continue
								} else if actualCallbackID := response.Body[0]; actualCallbackID != callbackID {
									// FIXME: better checking
									log.Printf("ERROR doRequests request response "+
										"ZWSendData reply callback mismatch 0x%02x != 0x%02x",
										actualCallbackID, callbackID)
									attempt++
									controller.sendToCallback(response)
									continue
								} else {
									// It matches! Nothing to do, fall through
								}
							}
						}

						request.Response = response
						gotResponse = true
						request.Chan <- 0

					case packet.PacketPreambleCAN:
						// FIXME: how to handle this
						log.Printf("ERROR doRequests request response got unexpected CAN")
						attempt++

					case packet.PacketPreambleACK:
						// FIXME: how to handle this
						log.Printf("ERROR doRequests request response got unexpected ACK")
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
					attempt++

				case <-controller.stopRequests:
					request.Err = fmt.Errorf("Controller closed")
					request.Chan <- 0
					controller.stoppedRequests <- 0
					return
				}
			}

			if !gotResponse {
				request.Err = errors.New("Failed to get request response")
				request.Chan <- 0
			}

		case <-controller.stopRequests:
			controller.stoppedRequests <- 0
			return
		}
	}
}
