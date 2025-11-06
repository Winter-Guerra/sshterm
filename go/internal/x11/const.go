//go:build x11

package x11

type reqCode uint8

const (
	CreateWindow            = reqCode(1)
	ChangeWindowAttributes  = reqCode(2)
	GetWindowAttributes     = reqCode(3)
	DestroyWindow           = reqCode(4)
	DestroySubwindows       = reqCode(5)
	ChangeSaveSet           = reqCode(6)
	ReparentWindow          = reqCode(7)
	MapWindow               = reqCode(8)
	MapSubwindows           = reqCode(9)
	UnmapWindow             = reqCode(10)
	UnmapSubwindows         = reqCode(11)
	ConfigureWindow         = reqCode(12)
	CirculateWindow         = reqCode(13)
	GetGeometry             = reqCode(14)
	QueryTree               = reqCode(15)
	InternAtom              = reqCode(16)
	GetAtomName             = reqCode(17)
	ChangeProperty          = reqCode(18)
	DeleteProperty          = reqCode(19)
	GetProperty             = reqCode(20)
	ListProperties          = reqCode(21)
	SetSelectionOwner       = reqCode(22)
	GetSelectionOwner       = reqCode(23)
	ConvertSelection        = reqCode(24)
	SendEvent               = reqCode(25)
	GrabPointer             = reqCode(26)
	UngrabPointer           = reqCode(27)
	GrabButton              = reqCode(28)
	UngrabButton            = reqCode(29)
	ChangeActivePointerGrab = reqCode(30)
	GrabKeyboard            = reqCode(31)
	UngrabKeyboard          = reqCode(32)
	GrabKey                 = reqCode(33)
	UngrabKey               = reqCode(34)
	AllowEvents             = reqCode(35)
	GrabServer              = reqCode(36)
	UngrabServer            = reqCode(37)
	QueryPointer            = reqCode(38)
	GetMotionEvents         = reqCode(39)
	TranslateCoords         = reqCode(40)
	WarpPointer             = reqCode(41)
	SetInputFocus           = reqCode(42)
	GetInputFocus           = reqCode(43)
	QueryKeymap             = reqCode(44)
	OpenFont                = reqCode(45)
	CloseFont               = reqCode(46)
	QueryFont               = reqCode(47)
	QueryTextExtents        = reqCode(48)
	ListFonts               = reqCode(49)
	ListFontsWithInfo       = reqCode(50)
	SetFontPath             = reqCode(51)
	GetFontPath             = reqCode(52)
	CreatePixmap            = reqCode(53)
	FreePixmap              = reqCode(54)
	CreateGC                = reqCode(55)
	ChangeGC                = reqCode(56)
	CopyGC                  = reqCode(57)
	SetDashes               = reqCode(58)
	SetClipRectangles       = reqCode(59)
	FreeGC                  = reqCode(60)
	ClearArea               = reqCode(61)
	CopyArea                = reqCode(62)
	CopyPlane               = reqCode(63)
	PolyPoint               = reqCode(64)
	PolyLine                = reqCode(65)
	PolySegment             = reqCode(66)
	PolyRectangle           = reqCode(67)
	PolyArc                 = reqCode(68)
	FillPoly                = reqCode(69)
	PolyFillRectangle       = reqCode(70)
	PolyFillArc             = reqCode(71)
	PutImage                = reqCode(72)
	GetImage                = reqCode(73)
	PolyText8               = reqCode(74)
	PolyText16              = reqCode(75)
	ImageText8              = reqCode(76)
	ImageText16             = reqCode(77)
	CreateColormap          = reqCode(78)
	FreeColormap            = reqCode(79)
	CopyColormapAndFree     = reqCode(80)
	InstallColormap         = reqCode(81)
	UninstallColormap       = reqCode(82)
	ListInstalledColormaps  = reqCode(83)
	AllocColor              = reqCode(84)
	AllocNamedColor         = reqCode(85)
	AllocColorCells         = reqCode(86)
	AllocColorPlanes        = reqCode(87)
	FreeColors              = reqCode(88)
	StoreColors             = reqCode(89)
	StoreNamedColor         = reqCode(90)
	QueryColors             = reqCode(91)
	LookupColor             = reqCode(92)
	CreateCursor            = reqCode(93)
	CreateGlyphCursor       = reqCode(94)
	FreeCursor              = reqCode(95)
	RecolorCursor           = reqCode(96)
	QueryBestSize           = reqCode(97)
	QueryExtension          = reqCode(98)
	ListExtensions          = reqCode(99)
	ChangeKeyboardMapping   = reqCode(100)
	GetKeyboardMapping      = reqCode(101)
	ChangeKeyboardControl   = reqCode(102)
	GetKeyboardControl      = reqCode(103)
	Bell                    = reqCode(104)
	ChangePointerControl    = reqCode(105)
	GetPointerControl       = reqCode(106)
	SetScreenSaver          = reqCode(107)
	GetScreenSaver          = reqCode(108)
	ChangeHosts             = reqCode(109)
	ListHosts               = reqCode(110)
	SetAccessControl        = reqCode(111)
	SetCloseDownMode        = reqCode(112)
	KillClient              = reqCode(113)
	RotateProperties        = reqCode(114)
	ForceScreenSaver        = reqCode(115)
	SetPointerMapping       = reqCode(116)
	GetPointerMapping       = reqCode(117)
	SetModifierMapping      = reqCode(118)
	GetModifierMapping      = reqCode(119)
	NoOperation             = reqCode(127)
)

const (
	ValueErrorCode     byte = 2
	ColormapErrorCode  byte = 12
	IDChoiceErrorCode  byte = 14
	WindowErrorCode    byte = 3
	GContextErrorCode  byte = 13
	PixmapErrorCode    byte = 4
	CursorErrorCode    byte = 6
	ColormapNotifyCode byte = 13
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

	// Graphics functions
	GXclear        = 0x0 // 0
	GXand          = 0x1 // src AND dst
	GXandReverse   = 0x2 // src AND NOT dst
	GXcopy         = 0x3 // src
	GXandInverted  = 0x4 // NOT src AND dst
	GXnoop         = 0x5 // dst
	GXxor          = 0x6 // src XOR dst
	GXor           = 0x7 // src OR dst
	GXnor          = 0x8 // NOT (src OR dst)
	GXequiv        = 0x9 // NOT (src XOR dst)
	GXinvert       = 0xa // NOT dst
	GXorReverse    = 0xb // src OR NOT dst
	GXcopyInverted = 0xc // NOT src
	GXorInverted   = 0xd // NOT src OR dst
	GXnand         = 0xe // NOT (src AND dst)
	GXset          = 0xf // 1
)

const (
	DoRed   byte = 1 << 0
	DoGreen byte = 1 << 1
	DoBlue  byte = 1 << 2
)
