package nsmclient

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/scgolang/osc"
)

func (c *NsmClient) nsmInitOsc(nsmUrl string) error {
	var err error
	c.nsmClientAddr, err = net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	nsmUrl = strings.TrimPrefix(nsmUrl, nsmOscUrlPrefix) // This osc library can't handle it?
	nsmUrl = strings.TrimSuffix(nsmUrl, "/")

	c.nsmServerAddr, err = net.ResolveUDPAddr("udp", nsmUrl)
	if err != nil {
		return fmt.Errorf("resolve udp addr failed: %v\n", err)
	}

	c.nsmOscCtx, c.nsmOscCancel = context.WithCancel(context.Background())
	c.Conn, err = osc.DialUDPContext(c.nsmOscCtx, "udp", c.nsmClientAddr, c.nsmServerAddr)
	if err != nil {
		return fmt.Errorf("listen udp failed: %v\n", err)
	}

	return nil
}

func (c *NsmClient) nsmOscHandler() osc.PatternMatching { // TODO oscConn
	return osc.PatternMatching{
		NsmAddrError: osc.Method(func(msg osc.Message) error {
			return c.nsmOscError(msg)
		}),
		NsmAddrReply: osc.Method(func(msg osc.Message) error {
			return c.nsmOscAnnounceReply(msg)
		}),
		NsmAddrClientOpen: osc.Method(func(msg osc.Message) error {
			return c.nsmOscOpen(msg)
		}),
		NsmAddrClientSave: osc.Method(func(msg osc.Message) error {
			return c.nsmOscSave(msg)
		}),
		NsmAddrClientSessionIsLoaded: osc.Method(func(msg osc.Message) error {
			return c.nsmOscSessionIsLoaded(msg)
		}),
		NsmAddrClientShowOptionalGui: osc.Method(func(msg osc.Message) error {
			return c.nsmOscShow(msg)
		}),
		NsmAddrClientHideOptionalGui: osc.Method(func(msg osc.Message) error {
			return c.nsmOscHide(msg)
		}),
		//Broadcast
	}
}

// goroutine
func (c *NsmClient) nsmStartOscServer() {
	c.nsmOscErrLogChan <- c.Serve(1, c.nsmOscHandler()) // TODO messagehandler
}
