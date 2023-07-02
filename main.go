package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/pwiecz/go-fltk"

	nsm "nsm-notes/nsmclient"
)

type app struct {
	Win        *fltk.Window
	TextBuffer *fltk.TextBuffer
	TextEditor *fltk.TextEditor
	saveButton *fltk.LightButton
	box        *fltk.Box
	fileName   string
	appIsDirty bool

	*nsm.NsmClient
}

func (a *app) buildGUI() {
	fltk.SetScheme(fltkScheme)
	_, _, w, h := fltk.ScreenWorkArea(fltkScreen)
	a.Win = fltk.NewWindowWithPosition(w/fltkWDivider, h/fltkHDivider, widgetWidth, widgetHeight)
	a.Win.SetLabel(APP_TITLE)
	a.Win.SetCallback(func() {
		a.setGuiHidden()
	})
	a.Win.SetColor(windowColor)
	//a.Win.SetShortcut(fltk.CTRL + 'q')

	if resizableWin {
		a.Win.Resizable(a.Win)
	}

	col := fltk.NewFlex(widgetPaddingWidth/2, widgetPaddingWidth/2, widgetWidth-widgetPaddingWidth, widgetHeight-widgetPaddingWidth)

	col.SetType(fltk.COLUMN)
	col.SetSpacing(widgetPaddingWidth)

	a.saveButton = fltk.NewLightButton(buttonXoffset, buttonYoffset, widgetWidth, buttonHeight, buttonName)
	a.saveButton.Visible()
	a.saveButton.SetValue(false)
	a.saveButton.SetCallbackCondition(fltk.WhenChanged)
	a.saveButton.SetCallback(func() {
		a.callbackMenuFileSave()
	})
	a.saveButton.SetShortcut(fltk.CTRL + 's')
	a.saveButton.SetColor(buttonColor)

	col.Fixed(a.saveButton, buttonHeight)

	a.TextBuffer = fltk.NewTextBuffer()
	a.TextEditor = fltk.NewTextEditor(editorXoffset, editorYoffset, a.Win.W(), a.Win.H()-buttonHeight)

	a.TextEditor.SetBuffer(a.TextBuffer)
	a.TextEditor.SetWrapMode(fltk.WRAP_AT_COLUMN, wrapTextAtLine)
	a.TextEditor.SetLabelColor(editorLabelColor)

	a.TextEditor.SetCallbackCondition(fltk.WhenChanged)
	a.TextEditor.SetCallback(func() {
		a.setAppDirty()
	})
	if resizableWin {
		a.TextEditor.Parent().Resizable(a.TextEditor)
	}
	a.box = fltk.NewBox(fltk.NO_BOX, widgetPaddingWidth, widgetHeight-5, widgetWidth-widgetPaddingWidth, 8)
	a.box.SetLabelSize(10)
	a.box.SetLabel("Esc to hide")
	//a.box.SetAlign(fltk.ALIGN_RIGHT)
	col.Fixed(a.box, 8)

	col.End()

	a.Win.End()
	a.setAppClean()

}

func (a *app) setAppClean() {
	if a.appIsDirty == true {
		a.NsmSendIsClean()
	}
	a.appIsDirty = false
}

func (a *app) setAppDirty() {
	a.saveButton.SetValue(true)
	if a.appIsDirty == false {
		a.appIsDirty = true
		a.NsmSendIsDirty()
	}
}

func (a *app) setGuiShown() {
	a.Win.Show()
	a.NsmSendGuiShown()
}

func (a *app) setGuiHidden() {
	a.Win.Hide()
	a.NsmSendGuiHidden()
}

func (a *app) setNsmCallbacksRequired() error {
	// set open callback
	a.NsmSetOpenCallback(func(path, displayName, clientId string) (outMsg string, err error) {
		a.fileName = path

		if err = a.openFile(); err != nil {
			outMsg = "failed to open file"
		}
		a.Win.SetLabel(displayName)
		return outMsg, err
	})

	// set save callback
	a.NsmSetSaveCallback(func() (outMsg string, err error) {
		if err = a.fileSave(); err != nil {
			outMsg = "failed to save"
		}
		return outMsg, err
	})

	return nil
}

func (a *app) setNsmCallbacksOptional() error {
	// set active callback. // is actually optional
	a.NsmSetSessionIsLoadedCallback(func() error {
		//fmt.Println("Nsm server tells us session is loaded")
		return nil
	})

	// broadcast
	// progress
	// message

	return nil
}

func (a *app) setNsmGuiCallbacks() error {
	a.NsmSetShowCallback(func() error {
		a.setGuiShown()
		return nil
	})
	a.NsmSetHideCallback(func() error {
		a.setGuiHidden()
		return nil
	})

	return nil
}

func (a *app) setNsmPreAnnounceSettings() error {
	if err := a.NsmSetClientCapabilities(nsm.NSM_OPTIONAL_GUI, nsm.NSM_DIRTY); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
	a.NsmSetPrettyName(APP_TITLE) // optional
	a.NsmSetAnnounceTimeout(5000) // optional
	return nil
}

func main() {

	nsmUrl, found := nsm.NsmUrlIsSet()
	if !found {
		fmt.Fprintf(os.Stderr, "Fatal: We can't connect to NSM, %s not set.\n", nsm.NsmEnvUrl)
		os.Exit(1)
	}

	a := app{}

	a.NsmClient = nsm.NsmNewClient()

	a.setNsmCallbacksRequired()
	a.setNsmCallbacksOptional()

	if err := a.NsmInit(nsmUrl); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
	defer a.NsmStop()

	a.NsmHandleSigterm()

	a.setNsmPreAnnounceSettings()

	if err := a.NsmAnnounce(); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}

	if a.NsmServerHasCapabilityOptionalGui() && a.NsmClientHasCapabilityOptionalGui() {
		a.setNsmGuiCallbacks()
	}

	a.buildGUI()

	if hideWinAtLaunch {
		a.setGuiHidden()
	} else {
		a.setGuiShown()
	}
	for {
		if err := a.NsmCheckWait(1); err != nil {
			if errors.Is(err, nsm.NsmGotSigtermErr) {
				fmt.Printf("[%v] got SIGTERM, bye\n", os.Args[0])
				os.Exit(0)
			} else {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		}

		fltk.Wait(0.17)
	}
}

func (a *app) openFile() error {
	// if file not exists, we need to create it.
	if _, err := os.Stat(a.fileName); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		if errors.Is(err, os.ErrNotExist) {
			f, err := os.Create(a.fileName)
			if err != nil {
				return fmt.Errorf("%v", err)
			}
			defer f.Close()
		}
	}

	textByte, err := os.ReadFile(a.fileName)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	a.TextBuffer.SetText(string(textByte))

	return nil
}
func (a *app) callbackMenuFileSave() { //error
	// send to chan? same as nsm? or Mutex?

	a.saveButton.SetValue(false)
	if err := a.fileSave(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}

	a.setAppClean()
	//a.saveButton.SetValue(false)
}

func (a *app) fileSave() error {
	if a.appIsDirty {
		info, _ := os.Stat(a.fileName)
		if err := os.WriteFile(a.fileName, []byte(a.TextBuffer.Text()), info.Mode()); err != nil {
			return err
		}

		a.saveButton.SetValue(false)

	}

	return nil
}
