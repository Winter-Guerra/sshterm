//go:build x11

package wire

type ReqCode uint8

const (
	CreateWindow            = ReqCode(1)
	ChangeWindowAttributes  = ReqCode(2)
	GetWindowAttributes     = ReqCode(3)
	DestroyWindow           = ReqCode(4)
	DestroySubwindows       = ReqCode(5)
	ChangeSaveSet           = ReqCode(6)
	ReparentWindow          = ReqCode(7)
	MapWindow               = ReqCode(8)
	MapSubwindows           = ReqCode(9)
	UnmapWindow             = ReqCode(10)
	UnmapSubwindows         = ReqCode(11)
	ConfigureWindow         = ReqCode(12)
	CirculateWindow         = ReqCode(13)
	GetGeometry             = ReqCode(14)
	QueryTree               = ReqCode(15)
	InternAtom              = ReqCode(16)
	GetAtomName             = ReqCode(17)
	ChangeProperty          = ReqCode(18)
	DeleteProperty          = ReqCode(19)
	GetProperty             = ReqCode(20)
	ListProperties          = ReqCode(21)
	SetSelectionOwner       = ReqCode(22)
	GetSelectionOwner       = ReqCode(23)
	ConvertSelection        = ReqCode(24)
	SendEvent               = ReqCode(25)
	GrabPointer             = ReqCode(26)
	UngrabPointer           = ReqCode(27)
	GrabButton              = ReqCode(28)
	UngrabButton            = ReqCode(29)
	ChangeActivePointerGrab = ReqCode(30)
	GrabKeyboard            = ReqCode(31)
	UngrabKeyboard          = ReqCode(32)
	GrabKey                 = ReqCode(33)
	UngrabKey               = ReqCode(34)
	AllowEvents             = ReqCode(35)
	GrabServer              = ReqCode(36)
	UngrabServer            = ReqCode(37)
	QueryPointer            = ReqCode(38)
	GetMotionEvents         = ReqCode(39)
	TranslateCoords         = ReqCode(40)
	WarpPointer             = ReqCode(41)
	SetInputFocus           = ReqCode(42)
	GetInputFocus           = ReqCode(43)
	QueryKeymap             = ReqCode(44)
	OpenFont                = ReqCode(45)
	CloseFont               = ReqCode(46)
	QueryFont               = ReqCode(47)
	QueryTextExtents        = ReqCode(48)
	ListFonts               = ReqCode(49)
	ListFontsWithInfo       = ReqCode(50)
	SetFontPath             = ReqCode(51)
	GetFontPath             = ReqCode(52)
	CreatePixmap            = ReqCode(53)
	FreePixmap              = ReqCode(54)
	CreateGC                = ReqCode(55)
	ChangeGC                = ReqCode(56)
	CopyGC                  = ReqCode(57)
	SetDashes               = ReqCode(58)
	SetClipRectangles       = ReqCode(59)
	FreeGC                  = ReqCode(60)
	ClearArea               = ReqCode(61)
	CopyArea                = ReqCode(62)
	CopyPlane               = ReqCode(63)
	PolyPoint               = ReqCode(64)
	PolyLine                = ReqCode(65)
	PolySegment             = ReqCode(66)
	PolyRectangle           = ReqCode(67)
	PolyArc                 = ReqCode(68)
	FillPoly                = ReqCode(69)
	PolyFillRectangle       = ReqCode(70)
	PolyFillArc             = ReqCode(71)
	PutImage                = ReqCode(72)
	GetImage                = ReqCode(73)
	PolyText8               = ReqCode(74)
	PolyText16              = ReqCode(75)
	ImageText8              = ReqCode(76)
	ImageText16             = ReqCode(77)
	CreateColormap          = ReqCode(78)
	FreeColormap            = ReqCode(79)
	CopyColormapAndFree     = ReqCode(80)
	InstallColormap         = ReqCode(81)
	UninstallColormap       = ReqCode(82)
	ListInstalledColormaps  = ReqCode(83)
	AllocColor              = ReqCode(84)
	AllocNamedColor         = ReqCode(85)
	AllocColorCells         = ReqCode(86)
	AllocColorPlanes        = ReqCode(87)
	FreeColors              = ReqCode(88)
	StoreColors             = ReqCode(89)
	StoreNamedColor         = ReqCode(90)
	QueryColors             = ReqCode(91)
	LookupColor             = ReqCode(92)
	CreateCursor            = ReqCode(93)
	CreateGlyphCursor       = ReqCode(94)
	FreeCursor              = ReqCode(95)
	RecolorCursor           = ReqCode(96)
	QueryBestSize           = ReqCode(97)
	QueryExtension          = ReqCode(98)
	ListExtensions          = ReqCode(99)
	ChangeKeyboardMapping   = ReqCode(100)
	GetKeyboardMapping      = ReqCode(101)
	ChangeKeyboardControl   = ReqCode(102)
	GetKeyboardControl      = ReqCode(103)
	Bell                    = ReqCode(104)
	ChangePointerControl    = ReqCode(105)
	GetPointerControl       = ReqCode(106)
	SetScreenSaver          = ReqCode(107)
	GetScreenSaver          = ReqCode(108)
	ChangeHosts             = ReqCode(109)
	ListHosts               = ReqCode(110)
	SetAccessControl        = ReqCode(111)
	SetCloseDownMode        = ReqCode(112)
	KillClient              = ReqCode(113)
	RotateProperties        = ReqCode(114)
	ForceScreenSaver        = ReqCode(115)
	SetPointerMapping       = ReqCode(116)
	GetPointerMapping       = ReqCode(117)
	SetModifierMapping      = ReqCode(118)
	GetModifierMapping      = ReqCode(119)
	NoOperation             = ReqCode(127)
	XInputOpcode            = ReqCode(131)
	BigRequestsOpcode       = ReqCode(133)
)

const (
	// XInputExtensionName is the name of the XInput extension.
	XInputExtensionName = "XInputExtension"
)

const (
	RequestErrorCode        byte = 1
	ValueErrorCode          byte = 2
	WindowErrorCode         byte = 3
	PixmapErrorCode         byte = 4
	AtomErrorCode           byte = 5
	CursorErrorCode         byte = 6
	FontErrorCode           byte = 7
	MatchErrorCode          byte = 8
	DrawableErrorCode       byte = 9
	AccessErrorCode         byte = 10
	AllocErrorCode          byte = 11
	ColormapErrorCode       byte = 12
	GContextErrorCode       byte = 13
	IDChoiceErrorCode       byte = 14
	NameErrorCode           byte = 15
	LengthErrorCode         byte = 16
	ImplementationErrorCode byte = 17
	DeviceErrorCode         byte = 20
)

const (
	ColormapNotifyCode byte = 13
)

// XInput event types
const (
	DeviceButtonPress   = 2
	DeviceButtonRelease = 3
	DeviceKeyPress      = 4
	DeviceKeyRelease    = 5
)

const (
	XI_DeviceButtonPress   = 1
	XI_DeviceButtonRelease = 2
	XI_DeviceKeyPress      = 3
	XI_DeviceKeyRelease    = 4
	XI_DeviceFocusIn       = 6
	XI_DeviceStateNotify   = 9
	XI_DeviceMappingNotify = 11
)

const (
	CWBackPixmap       = 1 << 0
	CWBackPixel        = 1 << 1
	CWBorderPixmap     = 1 << 2
	CWBorderPixel      = 1 << 3
	CWBitGravity       = 1 << 4
	CWWinGravity       = 1 << 5
	CWBackingStore     = 1 << 6
	CWBackingPlanes    = 1 << 7
	CWBackingPixel     = 1 << 8
	CWOverrideRedirect = 1 << 9
	CWSaveUnder        = 1 << 10
	CWEventMask        = 1 << 11
	CWDontPropagate    = 1 << 12
	CWColormap         = 1 << 13
	CWCursor           = 1 << 14
	CWSibling          = 1 << 15
	CWStackMode        = 1 << 16
)

const (
	DoRed   byte = 1 << 0
	DoGreen byte = 1 << 1
	DoBlue  byte = 1 << 2
)

// Constants for Keyboard Control
const (
	KBKeyClickPercent = 1 << 0
	KBBellPercent     = 1 << 1
	KBBellPitch       = 1 << 2
	KBBellDuration    = 1 << 3
	KBLed             = 1 << 4
	KBLedMode         = 1 << 5
	KBKey             = 1 << 6
	KBAutoRepeatMode  = 1 << 7
)

const (
	KeyPressMask             = 1 << 0
	KeyReleaseMask           = 1 << 1
	ButtonPressMask          = 1 << 2
	ButtonReleaseMask        = 1 << 3
	EnterWindowMask          = 1 << 4
	LeaveWindowMask          = 1 << 5
	PointerMotionMask        = 1 << 6
	PointerMotionHintMask    = 1 << 7
	Button1MotionMask        = 1 << 8
	Button2MotionMask        = 1 << 9
	Button3MotionMask        = 1 << 10
	Button4MotionMask        = 1 << 11
	Button5MotionMask        = 1 << 12
	ButtonMotionMask         = 1 << 13
	KeymapStateMask          = 1 << 14
	ExposureMask             = 1 << 15
	VisibilityChangeMask     = 1 << 16
	StructureNotifyMask      = 1 << 17
	ResizeRedirectMask       = 1 << 18
	SubstructureNotifyMask   = 1 << 19
	SubstructureRedirectMask = 1 << 20
	FocusChangeMask          = 1 << 21
	PropertyChangeMask       = 1 << 22
	ColormapChangeMask       = 1 << 23
	OwnerGrabButtonMask      = 1 << 24
)

// XInput event masks
const (
	DeviceKeyPressMask      = 1 << 0
	DeviceKeyReleaseMask    = 1 << 1
	DeviceButtonPressMask   = 1 << 2
	DeviceButtonReleaseMask = 1 << 3
)

const (
	ShiftMask   = 1 << 0
	LockMask    = 1 << 1
	ControlMask = 1 << 2
	Mod1Mask    = 1 << 3
	Mod2Mask    = 1 << 4
	Mod3Mask    = 1 << 5
	Mod4Mask    = 1 << 6
	Mod5Mask    = 1 << 7
	Button1Mask = 1 << 8
	Button2Mask = 1 << 9
	Button3Mask = 1 << 10
	Button4Mask = 1 << 11
	Button5Mask = 1 << 12
	AnyModifier = 1 << 15
)

const (
	GrabSuccess     byte = 0
	AlreadyGrabbed  byte = 1
	GrabInvalidTime byte = 2
	GrabNotViewable byte = 3
	GrabFrozen      byte = 4
)

const (
	InputOutput = 1 // Window class
)

const (
	NorthWestGravity = 1 // Bit gravity
)

const (
	NotUseful = 0 // Backing store
)

const (
	IsUnmapped = 0 // Map state
)
