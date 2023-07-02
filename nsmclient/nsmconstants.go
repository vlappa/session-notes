package nsmclient

const (
	nsmApiVersionMajor = 1
	nsmApiVersionMinor = 0
)

const (
	nsmOkMsg                  = "Ok"
	NsmEnvUrl                 = "NSM_URL"
	nsmOscUrlPrefix           = "osc.udp://"
	nsmDefaultAnnounceTimeout = 100000 // milliseconds
)

type nsmErr int

const (
	NSM_ERR_OK                nsmErr = 0
	NSM_ERR_GENERAL_ERROR     nsmErr = -1
	NSM_ERR_INCOMPATIBLE_API  nsmErr = -2
	NSM_ERR_BLACKLISTED       nsmErr = -3
	NSM_ERR_LAUNCH_FAILED     nsmErr = -4
	NSM_ERR_NO_SUCH_FILE      nsmErr = -5
	NSM_ERR_NO_SESSION_OPEN   nsmErr = -6
	NSM_ERR_UNSAVED_CHANGES   nsmErr = -7
	NSM_ERR_NOT_NOW           nsmErr = -8
	NSM_ERR_BAD_PROJECT       nsmErr = -9
	NSM_ERR_CREATE_FAILED     nsmErr = -10
	NSM_ERR_SESSION_LOCKED    nsmErr = -11
	NSM_ERR_OPERATION_PENDING nsmErr = -12
)

// NSM Client capabilities
const (
	NSM_SWITCH       nsmCapability = ":switch:"
	NSM_OPTIONAL_GUI nsmCapability = ":optional-gui:"
	NSM_MESSAGE      nsmCapability = ":message:"
	NSM_BROADCAST    nsmCapability = ":broadcast:"
	NSM_DIRTY        nsmCapability = ":dirty:"
	NSM_PROGRESS     nsmCapability = ":progress:"
)

// NSM server capabilities
const (
	NSM_S_OPTIONAL_GUI   nsmServerCapability = ":optional-gui:"
	NSM_S_SERVER_CONTROL nsmServerCapability = ":server_control:"
	NSM_S_BROADCAST      nsmServerCapability = ":broadcast:"
)

type nsmMsgLevel int

// NSM :message: priority levels.
const (
	NSM_MESSAGE_PRIORITY_LOWEST nsmMsgLevel = 0
	NSM_MESSAGE_PRIORITY_LOW    nsmMsgLevel = 1
	NSM_MESSAGE_PRIORITY_MED    nsmMsgLevel = 2
	NSM_MESSAGE_PRIORITY_HIGH   nsmMsgLevel = 3
)

const (
	NsmAddrReply                 = "/reply"
	NsmAddrError                 = "/error"
	NsmAddrClientProgress        = "/nsm/client/progress"
	NsmAddrClientIsDirty         = "/nsm/client/is_dirty"
	NsmAddrClientIsClean         = "/nsm/client/is_clean"
	NsmAddrClientMessage         = "/nsm/client/message"
	NsmAddrClientGuiIsHidden     = "/nsm/client/gui_is_hidden"
	NsmAddrClientGuiIsShown      = "/nsm/client/gui_is_shown"
	NsmAddrClientOpen            = "/nsm/client/open"
	NsmAddrClientSave            = "/nsm/client/save"
	NsmAddrClientSessionIsLoaded = "/nsm/client/session_is_loaded"
	NsmAddrClientShowOptionalGui = "/nsm/client/show_optional_gui"
	NsmAddrClientHideOptionalGui = "/nsm/client/hide_optional_gui"
	NsmAddrClientLabel           = "/nsm/client/label"
	NsmAddrServerBroadcast       = "/nsm/server/broadcast"
	NsmAddrServerAnnouce         = "/nsm/server/announce"
)
