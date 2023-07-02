package nsmclient

import (
	"fmt"
	"os"

	"github.com/scgolang/osc"
)

func (c *NsmClient) nsmSendReply(nsmReply NsmReply) error {
	if c.nsmServerIsActive {
		oscMsg := okReplyOscMsg(nsmReply)
		if err := c.Send(oscMsg); err != nil {
			return err
		}
	}
	return nil
}

func (c *NsmClient) nsmSendErrorReply(nsmReply NsmReply) error {
	if c.nsmServerIsActive {
		oscMsg := errorReplyOscMsg(nsmReply)
		if err := c.Send(oscMsg); err != nil {
			return err
		}
	}
	return nil
}

func (c *NsmClient) nsmSendIsDirty() error {
	if c.nsmServerIsActive {
		oscMsg := isDirtyOscMsg()
		if err := c.Send(oscMsg); err != nil {
			return err
		}
	}
	return nil
}

func (c *NsmClient) nsmSendIsClean() error {
	if c.nsmServerIsActive {
		oscMsg := isCleanOscMsg()
		if err := c.Send(oscMsg); err != nil {
			return err
		}
	}
	return nil
}

func (c *NsmClient) nsmSendIsShown() error {
	if c.nsmServerIsActive {
		oscMsg := guiShownOscMsg()
		if err := c.Send(oscMsg); err != nil {
			return err
		}
	}
	return nil
}

func (c *NsmClient) nsmSendIsHidden() error {
	if c.nsmServerIsActive {
		oscMsg := guiHiddenOscMsg()
		if err := c.Send(oscMsg); err != nil {
			return err
		}
	}
	return nil
}

/*
func (c *NsmClient) NsmSendProgress() {
	if c.nsmServerIsActive {
		oscMsg := progressOscMsg()
		if err := c.Send( oscMsg); err != nil {
			jlog.Errorf("%v", err)
		}
	}
}
*/

func (c *NsmClient) nsmSendAnnounce() error {
	name := os.Args[0]
	if c.nsmPrettyClientName == "" {
		c.nsmPrettyClientName = name
	}
	if c.nsmClientCapabilities == "" {
		return fmt.Errorf("err: no capabilities, can't send empty osc field, because of a bug in scgolang/osc")
	}
	oscMsg := c.announceOscMsg(c.nsmPrettyClientName, c.nsmClientCapabilities, name, c.nsmClientPid)
	if err := c.Send(oscMsg); err != nil {
		return fmt.Errorf("%v", err)
	}
	return nil
}

func okReplyOscMsg(nsmReply NsmReply) osc.Message {
	return osc.Message{Address: NsmAddrReply,
		Arguments: osc.Arguments{
			osc.String(nsmReply.Addr()),
			osc.String(nsmReply.Msg())}}
}

func errorReplyOscMsg(nsmReply NsmReply) osc.Message {
	return osc.Message{Address: NsmAddrError,
		Arguments: osc.Arguments{
			osc.String(nsmReply.Addr()),
			osc.Int(int32(nsmReply.Code())),
			osc.String(nsmReply.Msg()),
		}}
}

func isDirtyOscMsg() osc.Message {
	var addr = NsmAddrClientIsDirty
	return osc.Message{Address: addr} // NOTE just send address also in nsmd??
}

func isCleanOscMsg() osc.Message {
	var addr = NsmAddrClientIsClean
	return osc.Message{Address: addr}
}

func guiShownOscMsg() osc.Message {
	var addr = NsmAddrClientGuiIsShown
	return osc.Message{Address: addr}
}

func guiHiddenOscMsg() osc.Message {
	var addr = NsmAddrClientGuiIsHidden
	return osc.Message{Address: addr}
}

func progressOscMsg(x float32) osc.Message {
	var addr = NsmAddrClientProgress
	return osc.Message{Address: addr, Arguments: osc.Arguments{osc.Float(x)}}
}

func (c *NsmClient) announceOscMsg(prettyName, capabilities, name string, pid int) osc.Message {

	return osc.Message{Address: NsmAddrServerAnnouce,
		Arguments: osc.Arguments{
			osc.String(prettyName),
			osc.String(capabilities),
			osc.String(name), //os.Args[0]),
			osc.Int(c.nsmApiVersionMajor),
			osc.Int(c.nsmApiVersionMinor),
			osc.Int(int32(pid))}} //c.PID)
}
