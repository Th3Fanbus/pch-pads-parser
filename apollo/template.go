package apollo

import (
	"strings"
)

// KeywordCheck - This function is used to filter parsed lines of the configuration file and
//                returns true if the keyword is contained in the line.
// line      : string from the configuration file
func KeywordCheck(line string) bool {
	for _, keyword := range []string{
		"GPIO_", "TCK", "TRST_B", "TMS", "TDI", "CX_PMODE", "CX_PREQ_B", "JTAGX", "CX_PRDY_B",
		"TDO", "CNV_BRI_DT", "CNV_BRI_RSP", "CNV_RGI_DT", "CNV_RGI_RSP", "SVID0_ALERT_B",
		"SVID0_DATA", "SVID0_CLK", "PMC_SPI_FS", "PMC_SPI_RXD", "PMC_SPI_TXD", "PMC_SPI_CLK",
		"PMIC_PWRGOOD", "PMIC_RESET_B", "PMIC_THERMTRIP_B", "PMIC_STDBY", "PROCHOT_B",
		"PMIC_I2C_SCL", "PMIC_I2C_SDA", "FST_SPI_CLK_FB", "OSC_CLK_OUT_", "PMU_AC_PRESENT",
		"PMU_BATLOW_B", "PMU_PLTRST_B", "PMU_PWRBTN_B", "PMU_RESETBUTTON_B", "PMU_SLP_S0_B",
		"PMU_SLP_S3_B", "PMU_SLP_S4_B", "PMU_SUSCLK", "PMU_WAKE_B", "SUS_STAT_B", "SUSPWRDNACK",
		"SMB_ALERTB", "SMB_CLK", "SMB_DATA", "LPC_ILB_SERIRQ", "LPC_CLKOUT", "LPC_AD", "LPC_CLKRUNB",
		"LPC_FRAMEB",
	} {
		if strings.Contains(line, keyword) {
			return true
		}
	}
	return false
}