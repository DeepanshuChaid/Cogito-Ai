package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func LogFatal(debug bool, format string, args ...interface{}) {
	if debug {
		log.Printf(format, args...)
	}
	// Always output valid JSON even on error (so AI CLI doesn't crash)
	errOutput := map[string]interface{}{
		"continue":       true,
		"suppressOutput": true,
	}
	jsonOut, _ := json.Marshal(errOutput)
	fmt.Println(string(jsonOut))
	os.Exit(0) // Exit 0 so AI CLI continues working
}

