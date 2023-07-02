package nsmclient

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/scgolang/osc"
)

// TODO
var (
	nsmReceiverTimeoutErr = errors.New("timeout")
	NsmGotSigtermErr      = errors.New("SIGTERM")
	NsmServerInactiveErr  = errors.New("Nsm server inactive")
)

type NsmOpenCallback func(path, displayName, nsmClientId string) (outMsg string, err error)
type NsmSaveCallback func() (outMsg string, err error)
type NsmShowGuiCallback func() error
type NsmHideGuiCallback func() error
type NsmActiveCallback func(b bool) error
type NsmSessionIsLoadedCallback func() error
type NsmBroadCallback func(s string, m osc.Message) error

//type NsmBroadCallback func(s string, m osc.Message) error

type NsmLabelCallback func(userdata any) error

type nsmChannels struct {
	nsmOpenInChan            chan []string
	nsmSaveInChan            chan bool
	nsmSessionIsLoadedInChan chan bool
	nsmActiveInChan          chan bool
	nsmGuiShowInChan         chan bool
	nsmGuiHideInChan         chan bool
	nsmBroadcastChan         chan bool
	nsmSigtermSignal         chan os.Signal
	nsmReplyOutChan          chan NsmReply
	nsmAnnounceOutChan       chan bool
	nsmGuiShownOutChan       chan bool
	nsmGuiHiddenOutChan      chan bool
	nsmIsDirtyOutChan        chan bool
	nsmIsCleanOutChan        chan bool
	nsmProgressOutChan       chan bool
	nsmMessageOutChan        chan string
	nsmLabelOutChan          chan string
	nsmBroadcastOutChan      chan osc.Message
	nsmSenderErrChan         chan error
	nsmCloseSenderChan       chan bool
	nsmOscErrLogChan         chan error
}

func (c *nsmChannels) nsmInitChannels() {
	c.nsmOpenInChan = make(chan []string)
	c.nsmSaveInChan = make(chan bool)
	c.nsmSessionIsLoadedInChan = make(chan bool)
	c.nsmActiveInChan = make(chan bool)
	c.nsmGuiShowInChan = make(chan bool)
	c.nsmGuiHideInChan = make(chan bool)
	c.nsmBroadcastChan = make(chan bool)
	c.nsmSigtermSignal = make(chan os.Signal, 1)
	c.nsmReplyOutChan = make(chan NsmReply)
	c.nsmAnnounceOutChan = make(chan bool)
	c.nsmGuiShownOutChan = make(chan bool)
	c.nsmGuiHiddenOutChan = make(chan bool)
	c.nsmIsDirtyOutChan = make(chan bool)
	c.nsmIsCleanOutChan = make(chan bool)
	c.nsmProgressOutChan = make(chan bool)
	c.nsmMessageOutChan = make(chan string)
	c.nsmLabelOutChan = make(chan string)
	c.nsmBroadcastOutChan = make(chan osc.Message)
	c.nsmSenderErrChan = make(chan error)
	c.nsmCloseSenderChan = make(chan bool)
	c.nsmOscErrLogChan = make(chan error)
}

type NsmClient struct {
	nsmChannels
	osc.Conn
	nsmClientAddr         *net.UDPAddr // TODO adjust to servers protocol (udp/tcp/unix)
	nsmServerAddr         *net.UDPAddr
	nsmServerName         string
	nsmServerCapabilities string
	nsmServerIsActive     bool
	nsmClientId           string
	nsmUrl                string
	nsmPrettyClientName   string
	nsmClientCapabilities string
	nsmApiVersionMajor    int
	nsmApiVersionMinor    int
	nsmClientPid          int
	nsmAnnounceTimeout    time.Duration
	nsmOscCtx             context.Context
	nsmOscCancel          context.CancelFunc

	open NsmOpenCallback // NOTE does this need to be a pointer?

	save NsmSaveCallback

	show NsmShowGuiCallback

	hide NsmHideGuiCallback

	active NsmActiveCallback

	sessionIsLoaded NsmSessionIsLoadedCallback

	label NsmLabelCallback

	broadcast NsmBroadCallback
}

func (c *NsmClient) NsmIsActive() bool {
	return c.nsmServerIsActive
}

func (c *NsmClient) NsmGetSessionManagerName() string {
	return c.nsmServerName
}

func (c *NsmClient) NsmGetSessionManagerFeatures() string {
	return c.nsmServerCapabilities
}

func (c *NsmClient) setNsmServerCapabilities(s string) { // TODO FIXME
	c.nsmServerCapabilities = s
}

func (c *NsmClient) NsmServerHasCapability(capability nsmServerCapability) bool {
	return strings.Contains(c.nsmServerCapabilities, capability.String())
}

func (c *NsmClient) NsmServerHasCapabilityOptionalGui() bool {
	return strings.Contains(c.nsmServerCapabilities, NSM_S_OPTIONAL_GUI.String())
}

func (c *NsmClient) NsmServerHasCapabilityBroadcast() bool {
	return strings.Contains(c.nsmServerCapabilities, NSM_S_BROADCAST.String())
}

func (c *NsmClient) NsmServerHasCapabilityServerControl() bool {
	return strings.Contains(c.nsmServerCapabilities, NSM_S_SERVER_CONTROL.String())
}

func (c *NsmClient) NsmSetClientCapabilities(capabilities ...nsmCapability) error {
	// This overwrites existing capacities
	var sep = ":"
	for _, p := range capabilities {
		if p == "" {
			return fmt.Errorf("capability is empty")
		}
		if p != NSM_BROADCAST && p != NSM_OPTIONAL_GUI && p != NSM_SWITCH && p != NSM_MESSAGE && p != NSM_DIRTY && p != NSM_PROGRESS {
			return fmt.Errorf("unknown capability: %s", p)
		}
	}

	c.nsmClientCapabilities = sep
	for _, p := range capabilities {
		c.nsmClientCapabilities = c.nsmClientCapabilities + strings.Trim(p.String(), sep) + sep

	}

	return nil
}

func (c *NsmClient) NsmClientHasCapabilityOptionalGui() bool {
	return strings.Contains(c.nsmClientCapabilities, NSM_OPTIONAL_GUI.String())
}

func (c *NsmClient) NsmClientCapabilities() string {
	return c.nsmClientCapabilities
}

func NsmNewClient() *NsmClient {
	return &NsmClient{
		nsmApiVersionMajor: nsmApiVersionMajor,
		nsmApiVersionMinor: nsmApiVersionMinor,
		nsmClientPid:       os.Getpid(),
	}
}

func (c *NsmClient) NsmSetApiVersionMajor(v int) error {
	c.nsmApiVersionMajor = v
	return nil
}

func (c *NsmClient) NsmSetApiVersionMinor(v int) error {
	c.nsmApiVersionMinor = v
	return nil
}

func (c *NsmClient) setNsmIsActive(b bool) {
	c.nsmServerIsActive = b
}

func (c *NsmClient) setSessionManagerName(name string) {
	c.nsmServerName = name
}

func (c *NsmClient) setNsmServerAddress(addr string) {
	c.nsmUrl = addr
}

/*
func (c *NsmClient) NsmSendBroadcast() {
	if c.nsmServerIsActive {
		//makeIsDirtyMsg
		// Send msg
	}
}
*/

// SET CALLBACKS

func (c *NsmClient) NsmSetOpenCallback(openCallback NsmOpenCallback) {
	c.open = openCallback
}

func (c *NsmClient) NsmSetSaveCallback(saveCallback NsmSaveCallback) {
	c.save = saveCallback
}

func (c *NsmClient) NsmSetShowCallback(showCallback NsmShowGuiCallback) {
	c.show = showCallback
}

func (c *NsmClient) NsmSetHideCallback(hideCallback NsmHideGuiCallback) {
	c.hide = hideCallback
}

func (c *NsmClient) NsmSetActiveCallback(activeCallback NsmActiveCallback) {
	c.active = activeCallback
}

func (c *NsmClient) NsmSetSessionIsLoadedCallback(sessionIsLoadedCallback NsmSessionIsLoadedCallback) {
	c.sessionIsLoaded = sessionIsLoadedCallback
}

func (c *NsmClient) NsmSetBroadcastCallback(broadcastCallback NsmBroadCallback) {
	c.broadcast = broadcastCallback
}

func (c *NsmClient) nsmOscOpen(msg osc.Message) error {
	if expected, got := 3, len(msg.Arguments); expected != got {
		return fmt.Errorf("nsmOscOpen, expected %d arguments, got %d", expected, got)
	}

	if c.open == nil { // NOTE here or in goroutine?
		return fmt.Errorf("open callback not set")
	}

	path, err := msg.Arguments[0].ReadString()
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	displayName, err := msg.Arguments[1].ReadString()
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	nsmClientId, err := msg.Arguments[2].ReadString()
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	var args = []string{path, displayName, nsmClientId}

	c.nsmOpenInChan <- args

	return nil
}

func (c *NsmClient) nsmOscSave(msg osc.Message) error {
	if expected, got := 0, len(msg.Arguments); expected != got {
		return fmt.Errorf("nsmOscSave, expected %d arguments, got %d", expected, got)
	}
	if c.save == nil {
		return fmt.Errorf("save callback not set")
	}

	c.nsmSaveInChan <- true

	return nil
}

func (c *NsmClient) nsmOscAnnounceReply(msg osc.Message) error {
	if got := len(msg.Arguments); got != 4 {
		return fmt.Errorf("expected 4 arguments in announce reply, got %d", got)
	}

	p, err := msg.Arguments[0].ReadString()
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	_, err = msg.Arguments[1].ReadString()
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	smName, err := msg.Arguments[2].ReadString()
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	capabilities, err := msg.Arguments[3].ReadString()
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	if p != NsmAddrServerAnnouce {
		return fmt.Errorf("announce reply: addr %v, doesn't match %v", p, NsmAddrServerAnnouce)
	}

	c.setNsmIsActive(true)
	c.setSessionManagerName(smName)
	c.setNsmServerAddress(msg.Address) // TODO
	c.setNsmServerCapabilities(capabilities)

	//fmt.Printf("NSM: Successfully registered. NSM server says: %s \n", serverMsg)

	c.nsmActiveInChan <- c.nsmServerIsActive

	return nil

}

func (c *NsmClient) nsmOscError(msg osc.Message) error {
	if expected, got := 3, len(msg.Arguments); expected != got {
		return fmt.Errorf("nsmOscError, expected %d arguments, got %d", expected, got)
	}
	p, err := msg.Arguments[0].ReadString()
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	if p != NsmAddrServerAnnouce {
		return fmt.Errorf("err: addr %v, doesn't match %v", p, NsmAddrServerAnnouce)
	}

	server, err := msg.Arguments[1].ReadString()
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	fmt.Fprintf(os.Stderr, "NSM: Failed to register with NSM server: %v", server)

	c.setNsmIsActive(false)

	c.nsmActiveInChan <- c.nsmServerIsActive

	return nil
}

func (c *NsmClient) nsmOscSessionIsLoaded(msg osc.Message) error {
	if expected, got := 0, len(msg.Arguments); expected != got {
		return fmt.Errorf("nsmOscSessionIsLoaded, expected %d arguments, got %d", expected, got)
	}

	if c.sessionIsLoaded == nil {
		return fmt.Errorf("sessionIsLoaded callback not set")
	}

	c.nsmSessionIsLoadedInChan <- true

	return nil
}

func (c *NsmClient) nsmOscShow(msg osc.Message) error {
	if expected, got := 0, len(msg.Arguments); expected != got {
		return fmt.Errorf("nsmOscShow, expected %d arguments, got %d", expected, got)
	}

	if c.show == nil {
		return fmt.Errorf("show callback not set")
	}

	c.nsmGuiShowInChan <- true

	return nil
}

func (c *NsmClient) nsmOscHide(msg osc.Message) error {
	if expected, got := 0, len(msg.Arguments); expected != got {
		return fmt.Errorf("nsmOscHide, expected %d arguments, got %d", expected, got)
	}
	if c.hide == nil {
		return fmt.Errorf("hide callback not set")
	}

	c.nsmGuiHideInChan <- true
	return nil
}

func (c *NsmClient) nsmOscBroadcast(msg osc.Message) error { return nil }

func (c *NsmClient) NsmSendIsClean() {
	c.nsmIsCleanOutChan <- true
}

func (c *NsmClient) NsmSendIsDirty() {
	c.nsmIsDirtyOutChan <- true
}

func (c *NsmClient) NsmSendGuiHidden() {
	c.nsmGuiHiddenOutChan <- true
}

func (c *NsmClient) NsmSendGuiShown() {
	c.nsmGuiShownOutChan <- true
}

func (c *NsmClient) NsmSetAnnounceTimeout(t time.Duration) {
	c.nsmAnnounceTimeout = t
}

func (c *NsmClient) NsmAnnounce() error {

	c.nsmAnnounceOutChan <- true

	var announceTimeout time.Duration
	if c.nsmAnnounceTimeout == 0 {
		announceTimeout = nsmDefaultAnnounceTimeout
	} else {
		announceTimeout = c.nsmAnnounceTimeout
	}
	timeout := time.After(announceTimeout * time.Millisecond)
	if err := c.nsmReceiver(timeout); err != nil {
		if errors.Is(err, NsmServerInactiveErr) {
			return err
		} else {
			return err
		}
	}

	return nil
}

func (c *NsmClient) NsmCheckWait(t time.Duration) error {
	timeout := time.After(t * time.Millisecond)
	if err := c.nsmReceiver(timeout); err != nil {
		if errors.Is(err, nsmReceiverTimeoutErr) {
			return nil
		} else {
			return err
		}
	}
	return nil
}
func (c *NsmClient) NsmCheckNoWait() error {
	if err := c.NsmCheckWait(0); err != nil {
		return err
	}
	return nil
}

func (c *NsmClient) nsmReceiver(timeout <-chan time.Time) error { // nsmCheckWait
	select {
	case args := <-c.nsmOpenInChan:
		var (
			outMsg string
			err    error
		)
		outMsg, err = c.open(args[0], args[1], args[2])
		if err != nil {
			c.nsmReplyOutChan <- NsmReply{NsmAddrClientOpen, NsmError{NSM_ERR_GENERAL_ERROR, outMsg}}
			break
		}
		c.nsmReplyOutChan <- NsmReply{NsmAddrClientOpen, NsmError{NSM_ERR_OK, nsmOkMsg}}
	case <-c.nsmSaveInChan:
		outMsg, err := c.save()
		if err != nil {
			c.nsmReplyOutChan <- NsmReply{NsmAddrClientSave, NsmError{NSM_ERR_GENERAL_ERROR, outMsg}}
			break
		}
		c.nsmReplyOutChan <- NsmReply{NsmAddrClientSave, NsmError{NSM_ERR_OK, nsmOkMsg}}
		c.nsmIsCleanOutChan <- true
	case <-c.nsmSessionIsLoadedInChan:
		c.sessionIsLoaded()
	case active := <-c.nsmActiveInChan:
		if !active {
			return fmt.Errorf("%w", NsmServerInactiveErr)
		} else if c.active != nil {
			c.active(active)
		}
	case <-c.nsmGuiShowInChan:
		if err := c.show(); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
	case <-c.nsmGuiHideInChan:
		if err := c.hide(); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
	case <-c.nsmBroadcastChan:
	case err := <-c.nsmSenderErrChan:
		fmt.Fprintf(os.Stderr, "%v\n", err)
	case <-c.nsmSigtermSignal:
		return fmt.Errorf("%w", NsmGotSigtermErr)
	case err := <-c.nsmOscErrLogChan:
		fmt.Printf("%v\n", err)
	case <-timeout:
		return fmt.Errorf("%w", nsmReceiverTimeoutErr)
	}

	return nil
}

func (c *NsmClient) NsmSetPrettyName(name string) {
	c.nsmPrettyClientName = name
}

// Sender goroutine

// handleNsmClientInfo runs a goroutine that handles the client to server informational messages.
// This is a persistent goroutine.
func (c *NsmClient) nsmSender() error {
	for {
		select {
		case <-c.nsmCloseSenderChan:
			return nil
		case <-c.nsmAnnounceOutChan:
			if err := c.nsmSendAnnounce(); err != nil {
				c.nsmSenderErrChan <- err
			}
		case nsmReply := <-c.nsmReplyOutChan:
			var err error
			if nsmReply.Code() == 0 {
				err = c.nsmSendReply(nsmReply)
			} else {
				err = c.nsmSendErrorReply(nsmReply)
			}
			if err != nil {
				c.nsmSenderErrChan <- err
			}
		case <-c.nsmGuiShownOutChan:
			if err := c.nsmSendIsShown(); err != nil {
				c.nsmSenderErrChan <- err
			}
		case <-c.nsmGuiHiddenOutChan:
			if err := c.nsmSendIsHidden(); err != nil {
				c.nsmSenderErrChan <- err
			}
		case <-c.nsmIsDirtyOutChan:
			if err := c.nsmSendIsDirty(); err != nil {
				c.nsmSenderErrChan <- err
			}
		case <-c.nsmIsCleanOutChan:
			if err := c.nsmSendIsClean(); err != nil {
				c.nsmSenderErrChan <- err
			}
			/*
				case x := <-c.nsmProgressOutChan:
					if err := c.nsmSendProgress(x); err != nil {
						c.nsmSenderErrChan <- err
					}
				case msg := <-c.nsmMessageOutChan:
					if err := c.nsmSendMessage(msg); err != nil {
						c.nsmSenderErrChan <- err
					}
				case msg := <-c.nsmLabelOutChan:
					if err := c.nsmSendLabel(msg); err != nil {
						c.nsmSenderErrChan <- err
					}
				case msg := <-c.nsmBroadcastOutChan:
					//c.nsmSenderErrChan <- err
			*/
		}
	}
}

func NsmUrlIsSet() (string, bool) {
	url, found := os.LookupEnv(NsmEnvUrl)
	return url, found
}

func (c *NsmClient) NsmInit(nsmUrl string) error {
	c.setNsmServerAddress(nsmUrl) // NOTE also done in announce?
	if err := c.nsmInitOsc(nsmUrl); err != nil {
		return err
	}
	c.nsmInitChannels()

	go c.nsmStartOscServer() // starts a goroutine
	go c.nsmSender()         // starts goroutine for sending msg to the NSM server.

	return nil
}

func (c *NsmClient) NsmCancelOscServer() error {
	// cancels the osc server.
	c.nsmOscCancel()
	return nil
}

func (c *NsmClient) NsmCancelOscSender() error {
	// stops the nsmOscSender goroutine
	c.nsmCloseSenderChan <- true
	return nil
}

func (c *NsmClient) NsmStop() error {
	// 1. cancels the osc server.
	// 2. cancels the nsmOscSender goroutine
	c.NsmCancelOscServer()
	c.NsmCancelOscSender()
	return nil
}

func (c *NsmClient) NsmHandleSigterm() error {
	signal.Notify(c.nsmSigtermSignal, os.Interrupt, syscall.SIGTERM)
	return nil
}

// func makeMessageMsg() {} TODO
// func makeBroadcastMsg() {} TODO
// func makeLabelMsg() {} TODO
