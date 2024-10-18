package main

import (
	"os"

	"github.com/bxrne/launchrail/internal/logger"
	"github.com/bxrne/launchrail/internal/ork"
)

func main() {
	log := logger.GetLogger()

	if len(os.Args) < 2 {
		log.Error("Usage: main <path_to_ork_file>")
		return
	}

	filePath := os.Args[1]

	ork_data, err := ork.Decompress(filePath)
	if err != nil {
		log.Error(err)
		return
	}

	log.Infof("Loaded %s designed by %s", ork_data.Rocket.Name, ork_data.Rocket.Designer)
	log.Debugf("%s (v%s)", ork_data.Creator, ork_data.Version)
}
