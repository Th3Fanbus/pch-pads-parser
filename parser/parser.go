package parser

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

import "../sunrise"
import "../apollo"

type generateMacro func(id string, dw0 uint32, dw1 uint32) string
type keywordFilter func(line string) bool

// padInfo - information about pad
// id       : pad id string
// offset   : the offset of the register address relative to the base
// function : the string that means the pad function
// dw0      : DW0 register value
// dw1      : DW1 register value
type padInfo struct {
	id       string
	offset   uint16
	function string
	dw0      uint32
	dw1      uint32
}

// titleFprint - print GPIO group title to file
// gpio : gpio.c file descriptor
func (info *padInfo) titleFprint(gpio *os.File) {
	fmt.Fprintf(gpio, "\n\t/* %s */\n", info.function)
}

// reservedFprint - print reserved GPIO to file as comment
// gpio : gpio.c file descriptor
func (info *padInfo) reservedFprint(gpio *os.File) {
	// small comment about reserved port
	fmt.Fprintf(gpio, "\t/* %s - %s */\n", info.id, info.function)
}

// padInfoRawFprint - print information about current pad to file using
// raw format:
// _PAD_CFG_STRUCT(GPP_F1, 0x84000502, 0x00003026), /* SATAXPCIE4 */
// gpio : gpio.c file descriptor
func (info *padInfo) padInfoRawFprint(gpio *os.File) {
	fmt.Fprintf(gpio,
		"\t_PAD_CFG_STRUCT(%s, 0x%0.8x, 0x%0.8x), /* %s */\n",
		info.id,
		info.dw0,
		(info.dw1 & 0xffffff00), // Interrupt Select - RO
		info.function)
}

// padInfoMacroFprint - print information about current pad to file using
// special macros:
// /* GPP_F1 - SATAXPCIE4 */
// PAD_CFG_NF(GPP_F1, 20K_PU, PLTRST, NF1),
// gpio : gpio.c file descriptor
func (info *padInfo) padInfoMacroFprint(gpio *os.File, genMacro generateMacro) {
	if len(info.function) > 0 {
		fmt.Fprintf(gpio, "\t/* %s - %s */\n", info.id, info.function)
	}
	fmt.Fprintf(gpio, "\t%s\n", genMacro(info.id, info.dw0, info.dw1))
}

// ParserData - global data
// line       : string from the configuration file
// padmap     : pad info map
// ConfigFile : file name with pad configuration in text form
// RawFmt     : flag for generating pads config file with DW0/1 reg raw values
// Template   : structure template type of ConfigFile
type ParserData struct {
	line       string
	padmap     []padInfo
	ConfigFile *os.File
	RawFmt     bool
	Template   int
	Platform   string
}

// padInfoExtract - adds a new entry to pad info map
// line : string from file with pad config map
func (parser *ParserData) padInfoExtract() int {
	var function, id string
	var dw0, dw1 uint32
	var template = map[int]template{
		0: useInteltoolLogTemplate, // inteltool.log
		1: useGpioHTemplate,        // gpio.h
		2: useYourTemplate,         // your file
	}
	if applyTemplate, valid := template[parser.Template]; valid {
		if applyTemplate(parser.line, &function, &id, &dw0, &dw1) == 0 {
			pad := padInfo{id: id, function: function, dw0: dw0, dw1: dw1}
			parser.padmap = append(parser.padmap, pad)
			return 0
		}
		fmt.Printf("This template (index %d) does not match"+
			" the entry in the configuration file!\n", parser.Template)
		return -1
	}
	fmt.Printf("There is no template for this type index %d\n", parser.Template)
	return -1
}

// communityGroupExtract
func (parser *ParserData) communityGroupExtract() {
	pad := padInfo{function: parser.line}
	parser.padmap = append(parser.padmap, pad)
}

// platformSpecMacroFuncGet - returns a platform specific macro generation function
func (parser *ParserData) platformSpecMacroFuncGet() generateMacro {
	if parser.Platform == "apollo" {
		return apollo.GenMacro
	}
	return sunrise.GenMacro
}

// platformSpecKeywordCheckFuncGet - returns a platform specific function to filter
//                                   keywords
func (parser *ParserData) platformSpecKeywordCheckFuncGet() keywordFilter {
	if parser.Platform == "apollo" {
		return apollo.KeywordCheck
	}
	return sunrise.KeywordCheck
}

// PadMapFprint - print pad info map to file
// gpio : gpio.c descriptor file
// raw  : in the case when this flag is false, pad information will be print
//        as macro
func (parser *ParserData) PadMapFprint(gpio *os.File) {
	gpio.WriteString("\n/* Pad configuration in ramstage */\n")
	gpio.WriteString("static const struct pad_config gpio_table[] = {\n")
	genMacro := parser.platformSpecMacroFuncGet()
	for _, pad := range parser.padmap {
		switch pad.dw0 {
		case 0:
			pad.titleFprint(gpio)
		case 0xffffffff:
			pad.reservedFprint(gpio)
		default:
			if parser.RawFmt {
				pad.padInfoRawFprint(gpio)
			} else {
				pad.padInfoMacroFprint(gpio, genMacro)
			}
		}
	}
	gpio.WriteString("};\n")

	// Add early configuration
	gpio.WriteString(`/* Early pad configuration in romstage */
static const struct pad_config early_gpio_table[] = {
	/*
	 * FIXME: This array cannot be generated automatically. Use macros here as
	 * you see fit.
	 */
};
`)
}

// Register - read specific platform registers (32 bits)
// line         : string from file with pad config map
// nameTemplate : register name femplate to filter parsed lines
// return
//     valid  : true if the dump of the register in intertool.log is set in accordance
//              with the template
//     name   : full register name
//     offset : register offset relative to the base address
//     value  : register value
func (parser *ParserData) Register(nameTemplate string) (
		valid bool, name string, offset uint32, value uint32) {
	if strings.Contains(parser.line, nameTemplate) && parser.Template == 0 {
		if registerInfoTemplate(parser.line, &name, &offset, &value) == 0 {
			fmt.Printf("\n\t/* %s : 0x%x : 0x%x */\n", name, offset, value)
			return true, name, offset, value
		}
	}
	return false, "ERROR", 0, 0
}

// Parse pads groupe information in the inteltool log file
// ConfigFile : name of inteltool log file
func (parser *ParserData) Parse() {
	// Read all lines from inteltool log file
	fmt.Println("Parse IntelTool Log File...")
	scanner := bufio.NewScanner(parser.ConfigFile)
	keywordFilterApply := parser.platformSpecKeywordCheckFuncGet()

	for scanner.Scan() {
		parser.line = scanner.Text()
		if strings.Contains(parser.line, "GPIO Community") || strings.Contains(parser.line, "GPIO Group") {
			parser.communityGroupExtract()
		} else if keywordFilterApply(parser.line) {
			if parser.padInfoExtract() != 0 {
				fmt.Println("...error!")
			}
		}
	}
	fmt.Println("...done!")
}
