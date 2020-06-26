package common

import "strconv"
import "fmt"

import "../config"

const (
	PAD_OWN_ACPI   = 0
	PAD_OWN_DRIVER = 1
)

const (
	TxLASTRxE     = 0x0
	Tx0RxDCRx0    = 0x1
	Tx0RxDCRx1    = 0x2
	Tx1RxDCRx0    = 0x3
	Tx1RxDCRx1    = 0x4
	Tx0RxE        = 0x5
	Tx1RxE        = 0x6
	HIZCRx0       = 0x7
	HIZCRx1       = 0x8
	TxDRxE        = 0x9
	StandbyIgnore = 0xf
)

const (
	IOSTERM_SAME	= 0x0
	IOSTERM_DISPUPD	= 0x1
	IOSTERM_ENPD	= 0x2
	IOSTERM_ENPU    = 0x3
)

const (
	RST_DEEP   = 1
)

const (
	TRIG_LEVEL       = 0
	TRIG_EDGE_SINGLE = 1
	TRIG_OFF         = 2
	TRIG_EDGE_BOTH   = 3
)

// PlatformSpecific - platform-specific interface
type PlatformSpecific interface {
	Rstsrc(macro *Macro)
	Pull(macro *Macro)
	GpiMacroAdd(macro *Macro)
	GpoMacroAdd(macro *Macro)
	NativeFunctionMacroAdd(macro *Macro)
	NoConnMacroAdd(macro *Macro)
}

// Macro - contains macro information and methods
// Platform : platform-specific interface
// padID    : pad ID string
// str      : macro string entirely
// Reg      : structure of configuration register values and their masks
type Macro struct {
	Platform  PlatformSpecific
	Reg       [MAX_DW_NUM]Register
	padID     string
	str       string
	ownership uint8
}

func (macro *Macro) PadIdGet() string {
	return macro.padID
}

func (macro *Macro) PadIdSet(padid string) *Macro {
	macro.padID = padid
	return macro
}

func (macro *Macro) SetPadOwnership(own uint8) *Macro {
	macro.ownership = own
	return macro
}

func (macro *Macro) IsOwnershipDriver() bool {
	return macro.ownership == PAD_OWN_DRIVER
}

// returns <Register> data configuration structure
// number : register number
func (macro *Macro) Register(number uint8) *Register {
	return &macro.Reg[number]
}

// add a string to macro
func (macro *Macro) Add(str string) *Macro {
	macro.str += str
	return macro
}

// set a string in a macro instead of its previous contents
func (macro *Macro) Set(str string) *Macro {
	macro.str = str
	return macro
}

// get macro string
func (macro *Macro) Get() string {
	return macro.str
}

// Adds PAD Id to the macro as a new argument
// return: Macro
func (macro *Macro) Id() *Macro {
	return macro.Add(macro.padID)
}

// Add Separator to macro if needed
func (macro *Macro) Separator() *Macro {
	str := macro.Get()
	c := str[len(str)-1]
	if c != '(' && c != '_' {
		macro.Add(", ")
	}
	return macro
}

// Adds the PADRSTCFG parameter from DW0 to the macro as a new argument
// return: Macro
func (macro *Macro) Rstsrc() *Macro {
	macro.Platform.Rstsrc(macro)
	return macro
}

// Adds The Pad Termination (TERM) parameter from DW1 to the macro as a new argument
// return: Macro
func (macro *Macro) Pull() *Macro {
	macro.Platform.Pull(macro)
	return macro
}

// Adds Pad GPO value to macro string as a new argument
// return: Macro
func (macro *Macro) Val() *Macro {
	dw0 := macro.Register(PAD_CFG_DW0)
	return macro.Separator().Add(strconv.Itoa(int(dw0.GetGPIOTXState())))
}

// Adds Pad GPO value to macro string as a new argument
// return: Macro
func (macro *Macro) Trig() *Macro {
	dw0 := macro.Register(PAD_CFG_DW0)
	var trig = map[uint8]string{
		0x0: "LEVEL",
		0x1: "EDGE_SINGLE",
		0x2: "OFF",
		0x3: "EDGE_BOTH",
	}
	return macro.Separator().Add(trig[dw0.GetRXLevelEdgeConfiguration()])
}

// Adds Pad Polarity Inversion Stage (RXINV) to macro string as a new argument
// return: Macro
func (macro *Macro) Invert() *Macro {
	macro.Separator()
	if macro.Register(PAD_CFG_DW0).GetRXLevelConfiguration() !=0 {
		return macro.Add("INVERT")
	}
	return macro.Add("NONE")
}

// Adds input/output buffer state
// return: Macro
func (macro *Macro) Bufdis() *Macro {
	var buffDisStat = map[uint8]string{
		0x0: "NO_DISABLE",    // both buffers are enabled
		0x1: "TX_DISABLE",    // output buffer is disabled
		0x2: "RX_DISABLE",    // input buffer is disabled
		0x3: "TX_RX_DISABLE", // both buffers are disabled
	}
	state := macro.Register(PAD_CFG_DW0).GetGPIORxTxDisableStatus()
	return macro.Separator().Add(buffDisStat[state])
}

// Adds macro to set the host software ownership
// return: Macro
func (macro *Macro) Own() *Macro {
	if macro.IsOwnershipDriver() {
		return macro.Separator().Add("DRIVER")
	}
	return macro.Separator().Add("ACPI")
}

//Adds pad native function (PMODE) as a new argument
//return: Macro
func (macro *Macro) Padfn() *Macro {
	dw0 := macro.Register(PAD_CFG_DW0)
	nfnum := int(dw0.GetPadMode())
	if nfnum != 0 {
		return macro.Separator().Add("NF" + strconv.Itoa(nfnum))
	}
	// GPIO used only for PAD_FUNC(x) macro
	return macro.Add("GPIO")
}

// Add a line to the macro that defines IO Standby State
// return: macro
func (macro *Macro) IOSstate() *Macro {
	var stateMacro = map[uint8]string{
		TxLASTRxE:     "TxLASTRxE",
		Tx0RxDCRx0:    "Tx0RxDCRx0",
		Tx0RxDCRx1:    "Tx0RxDCRx1",
		Tx1RxDCRx0:    "Tx1RxDCRx0",
		Tx1RxDCRx1:    "Tx1RxDCRx1",
		Tx0RxE:        "Tx0RxE",
		Tx1RxE:        "Tx1RxE",
		HIZCRx0:       "HIZCRx0",
		HIZCRx1:       "HIZCRx1",
		TxDRxE:        "TxDRxE",
		StandbyIgnore: "IGNORE",
	}
	dw1 := macro.Register(PAD_CFG_DW1)
	str, valid := stateMacro[dw1.GetIOStandbyState()]
	if !valid {
		// ignore setting for incorrect value
		str = "IGNORE"
	}
	return macro.Separator().Add(str)
}

// Add a line to the macro that defines IO Standby Termination
// return: macro
func (macro *Macro) IOTerm() *Macro {
	var ioTermMacro = map[uint8]string{
		IOSTERM_SAME:    "SAME",
		IOSTERM_DISPUPD: "DISPUPD",
		IOSTERM_ENPD:    "ENPD",
		IOSTERM_ENPU:    "ENPU",
	}
	dw1 := macro.Register(PAD_CFG_DW1)
	return macro.Separator().Add(ioTermMacro[dw1.GetIOStandbyTermination()])
}

// Check created macro
func (macro *Macro) check() *Macro {
	if !macro.Register(PAD_CFG_DW0).MaskCheck() {
		return macro.Advanced()
	}
	return macro
}

// or - Set " | " if its needed
func (macro *Macro) or() *Macro {

		if str := macro.Get(); str[len(str) - 1] == ')' {
			macro.Add(" | ")
		}
		return macro
}

// dw0Decode - decode value of DW0 register
func (macro *Macro) dw0Decode() *Macro {
	dw0 := macro.Register(PAD_CFG_DW0)

	irqroute := func(macro *Macro, name string) {
		for key, isRoute := range map[string]func()uint8{
			"IOAPIC": dw0.GetGPIOInputRouteIOxAPIC,
			"SCI":    dw0.GetGPIOInputRouteSCI,
			"SMI":    dw0.GetGPIOInputRouteSMI,
			"NMI":    dw0.GetGPIOInputRouteNMI,
		} {
			if name == key && isRoute() != 0 {
				macro.or().Add("PAD_IRQ_ROUTE(").Add(name).Add(")")
			}
		}
	}

	somebits := func(macro *Macro, name string) {
		for key, isRoute := range map[string]func()uint8{
			"(1 << 29)": dw0.GetRXPadStateSelect,
			"(1 << 28)": dw0.GetRXRawOverrideStatus,
			"1":         dw0.GetGPIOTXState,
		} {
			if name == key && isRoute() != 0 {
				macro.or().Add(name)
			}
		}
	}

	for _, slice := range []struct {
        name   string
        action func(macro *Macro, name string)
	} {
		{	// PAD_FUNC(NF3)
			"PAD_FUNC",
			func(macro *Macro, name string) {
				if dw0.GetPadMode() != 0 || config.InfoLevelGet() <= 3 {
					macro.or().Add(name).Add("(").Padfn().Add(")")
				}
			},
		},

		{	// PAD_RESET(DEEP)
			"PAD_RESET",
			func(macro *Macro, name string) {
				if dw0.GetResetConfig() != 0 {
					macro.or().Add(name).Add("(").Rstsrc().Add(")")
				}
			},
		},

		{	// PAD_TRIG(OFF)
			"PAD_TRIG",
			func(macro *Macro, name string) {
				if dw0.GetRXLevelEdgeConfiguration() != 0 {
					macro.or().Add(name).Add("(").Trig().Add(")")
				}
			},
		},

		{	// PAD_IRQ_ROUTE(IOAPIC)
			"IOAPIC",
			irqroute,
		},

		{	// PAD_IRQ_ROUTE(SCI)
			"SCI",
			irqroute,
		},

		{	// PAD_IRQ_ROUTE(SMI)
			"SMI",
			irqroute,
		},

		{	// PAD_IRQ_ROUTE(NMI)
			"NMI",
			irqroute,
		},

		{	// PAD_RX_POL(EDGE_SINGLE)
			"PAD_RX_POL",
			func(macro *Macro, name string) {
				if dw0.GetRXLevelConfiguration() != 0 {
					macro.or().Add(name).Add("(").Invert().Add(")")
				}
			},
		},

		{	// PAD_BUF(TX_RX_DISABLE)
			"PAD_BUF",
			func(macro *Macro, name string) {
				if dw0.GetGPIORxTxDisableStatus() != 0 {
					macro.or().Add(name).Add("(").Bufdis().Add(")")
				}
			},
		},

		{	// (1 << 29)
			"(1 << 29)",
			somebits,
		},

		{	// (1 << 28)
			"(1 << 28)",
			somebits,
		},

		{	// 1
			"1",
			somebits,
		},
	} {
		slice.action(macro, slice.name)
	}
	return macro
}

// dw1Decode - decode value of DW1 register
func (macro *Macro) dw1Decode() *Macro {
	dw1 := macro.Register(PAD_CFG_DW1)
	for _, slice := range []struct {
        name   string
        action func(macro *Macro, name string)
	} {
		// PAD_PULL(DN_20K)
		{
			"PAD_PULL",
			func(macro *Macro, name string) {
				if dw1.GetTermination() != 0 {
					macro.or().Add(name).Add("(").Pull().Add(")")
				}
			},
		},

		// PAD_IOSSTATE(HIZCRx0)
		{
			"PAD_IOSSTATE",
			func(macro *Macro, name string) {
				if dw1.GetIOStandbyState() != 0 {
					macro.or().Add(name).Add("(").IOSstate().Add(")")
				}
			},
		},

		// PAD_IOSTERM(ENPU)
		{
			"PAD_IOSTERM",
			func(macro *Macro, name string)  {
				if dw1.GetIOStandbyTermination() != 0 {
					macro.or().Add(name).Add("(").IOTerm().Add(")")
				}
			},
		},

		// PAD_CFG_OWN_GPIO(DRIVER)
		{
			"PAD_CFG_OWN_GPIO",
			func(macro *Macro, name string)  {
				if macro.IsOwnershipDriver() {
					macro.or().Add(name).Add("(").Own().Add(")")
				}
			},
		},
	} {
		slice.action(macro, slice.name)
	}
	return macro
}

// AddToMacroIgnoredMask - Print info about ignored field mask
// title - warning message
func (macro *Macro) AddToMacroIgnoredMask(title string) *Macro {
	dw0 := macro.Register(PAD_CFG_DW0)
	dw1 := macro.Register(PAD_CFG_DW1)

	// Get mask of ignored bit fields.
	dw0Ignored := dw0.IgnoredFieldsGet()
	dw1Ignored := dw1.IgnoredFieldsGet()
	if (dw0Ignored != 0 || dw1Ignored != 0) && config.InfoLevelGet() >= 3 {
		// If some fields were ignored when the macro was generated, then we will
		// show them in the comment
		dw0info := fmt.Sprintf("DW0(0x%0.8x) ", dw0Ignored)
		macro.Add("\n\t/* ").Add(title).Add(dw0info)
		if dw1Ignored != 0 {
			dw1info := fmt.Sprintf("DW1(0x%0.8x) ", dw1Ignored)
			macro.Add(dw1info)
		}
		macro.Add("*/")
		// Decode ignored mask
		if config.InfoLevelGet() >= 4 {
			if dw0Ignored != 0 {
				dw0temp := dw0.ValueGet()
				dw0.ValueSet(dw0Ignored)
				macro.Add("\n\t/* (!) DW0 : ").dw0Decode().Add(" - IGNORED */")
				dw0.ValueSet(dw0temp)
			}
			if dw1Ignored != 0 {
				dw1temp	:= dw1.ValueGet()
				dw1.ValueSet(dw1Ignored)
				macro.Add("\n\t/* (!) DW1 : ").dw1Decode().Add(" - IGNORED */")
				dw1.ValueSet(dw1temp)
			}
		}
	}
	return macro
}

// Generate Advanced Macro
func (macro *Macro) Advanced() *Macro {
	dw0 := macro.Register(PAD_CFG_DW0)
	dw1 := macro.Register(PAD_CFG_DW1)

	// Get mask of ignored bit fields.
	dw0Ignored := dw0.IgnoredFieldsGet()
	dw1Ignored := dw1.IgnoredFieldsGet()

	if config.InfoLevelGet() <= 1 {
		macro.Set("")
	} else if config.InfoLevelGet() >= 2 {
		// Add string of reference macro as a comment
		reference := macro.Get()
		macro.Set("/* ").Add(reference).Add(" */")
		macro.AddToMacroIgnoredMask("(!) NEED TO IGNORE THESE FIELDS: ")
		macro.Add("\n\t")
	}
	macro.Add("_PAD_CFG_STRUCT(").Id().Add(",")

	if config.AreFieldsIgnored() {
		// Consider bit fields that should be ignored when regenerating
		// advansed macros
		var tempVal uint32 = dw0.ValueGet() & ^dw0Ignored
		dw0.ValueSet(tempVal)

		tempVal = dw1.ValueGet() & ^dw1Ignored
		dw1.ValueSet(tempVal)
	}

	macro.Add("\n\t\t").dw0Decode()
	if dw1.ValueGet() != 0 {
		macro.Add(",\n\t\t").dw1Decode().Add("),")
	} else {
		macro.Add(", 0),")
	}
	return macro
}

// Generate macro for bi-directional GPIO port
func (macro *Macro) Bidirection() {
	dw1 := macro.Register(PAD_CFG_DW1)
	ios := dw1.GetIOStandbyState() != 0 || dw1.GetIOStandbyTermination() != 0
	macro.Set("PAD_CFG_GPIO_BIDIRECT")
	if ios {
		macro.Add("_IOS")
	}
	// PAD_CFG_GPIO_BIDIRECT(pad, val, pull, rst, trig, own)
	macro.Add("(").Id().Val().Pull().Rstsrc().Trig()
	if ios {
		// PAD_CFG_GPIO_BIDIRECT_IOS(pad, val, pull, rst, trig, iosstate, iosterm, own)
		macro.IOSstate().IOTerm()
	}
	macro.Own().Add("),")
}

const (
	rxDisable uint8 = 0x2
	txDisable uint8 = 0x1
)

// Gets base string of current macro
// return: string of macro
func (macro *Macro) Generate() string {
	macro.Set("PAD_CFG")
	dw0 := macro.Register(PAD_CFG_DW0)
	if dw0.GetPadMode() == 0 {
		// GPIO
		switch dw0.GetGPIORxTxDisableStatus() {
		case txDisable:
			macro.Platform.GpiMacroAdd(macro) // GPI

		case rxDisable:
			macro.Platform.GpoMacroAdd(macro) // GPO

		case rxDisable | txDisable:
			macro.Platform.NoConnMacroAdd(macro) // NC

		default:
			macro.Bidirection()
		}
	} else {
		macro.Platform.NativeFunctionMacroAdd(macro)
	}

	if config.IsAdvancedFormatUsed() {
		// Clear control mask to generate advanced macro only
		return macro.Advanced().Get()
	}

	if config.IsNonCheckingFlagUsed() {
		macro.AddToMacroIgnoredMask("(!) THESE FIELDS WERE IGNORED: ")
		return macro.Get()
	}

	return macro.check().Get()
}
