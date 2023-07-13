package logger

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

var (
	LoggerCfg *Config
)

type Config struct {
	InfoPrefix    string
	WarningPrefix string
	ErrorPrefix   string
	Colors        bool
	SBL, SBR      string
}

func init() {
	LoggerCfg = &Config{}

	LoggerCfg.InfoPrefix = " INFO"
	LoggerCfg.ErrorPrefix = " ERROR"
	LoggerCfg.WarningPrefix = " WARN"

	if os.Getenv("TERM") == "dumb" || os.Getenv("NO_COLOR") == "1" {
		LoggerCfg.Colors = false

		LoggerCfg.SBR = "] "
		LoggerCfg.SBL = "["

		LoggerCfg.InfoPrefix = LoggerCfg.InfoPrefix + "] "
		LoggerCfg.WarningPrefix = LoggerCfg.WarningPrefix + "] "
		LoggerCfg.ErrorPrefix = LoggerCfg.ErrorPrefix + "] "
	} else {
		LoggerCfg.SBR = "\x1b[90m]\x1b[0m "
		LoggerCfg.SBL = "\x1b[90m[\x1b[0m"

		LoggerCfg.Colors = true
		LoggerCfg.InfoPrefix = "\x1b[36m" + LoggerCfg.InfoPrefix + "\x1b[0m" + LoggerCfg.SBR
		LoggerCfg.WarningPrefix = "\x1b[33m" + LoggerCfg.WarningPrefix + "\x1b[0m" + LoggerCfg.SBR
		LoggerCfg.ErrorPrefix = "\x1b[31m" + LoggerCfg.ErrorPrefix + "\x1b[0m" + LoggerCfg.SBR
	}
}

func nowFormated() string {
	now := time.Now()

	return strconv.Itoa(now.Year()) + "/" +
		strconv.Itoa(int(now.Month())) + "/" +
		strconv.Itoa(now.Day()) + " " +
		strconv.Itoa(now.Hour()) + ":" +
		strconv.Itoa(now.Minute()) + ":" +
		strconv.Itoa(now.Second()) + "." +
		strconv.Itoa(now.Nanosecond()/1000)
}

func Info(format string, args ...any) {
	fm := fmt.Sprintf(LoggerCfg.InfoPrefix+format, args...)
	os.Stdout.WriteString("\r" + LoggerCfg.SBL + nowFormated() + fm + "\n")
}

func Warn(format string, args ...any) {
	fm := fmt.Sprintf(LoggerCfg.WarningPrefix+format, args...)
	os.Stdout.WriteString("\r" + LoggerCfg.SBL + nowFormated() + fm + "\n")
}

func Error(format string, args ...any) {
	fm := fmt.Sprintf(LoggerCfg.ErrorPrefix+format, args...)
	os.Stderr.WriteString("\r" + LoggerCfg.SBL + nowFormated() + fm + "\n")
}

func Fatal(args ...any) {
	args = append([]any{LoggerCfg.ErrorPrefix}, args...)
	fm := fmt.Sprintln(args...)
	os.Stderr.WriteString("\r" + LoggerCfg.SBL + nowFormated() + fm + "\n")
	os.Exit(1)
}
