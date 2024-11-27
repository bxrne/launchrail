package report

import (
	"encoding/csv"
	"fmt"
	"github.com/bxrne/launchrail/pkg/entities"
	"github.com/bxrne/launchrail/pkg/systems"
	"os"
)

// Generate generates a report from the stored states.
func Generate(states []map[entities.Entity]systems.State, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"Tick", "Entity", "Position X", "Position Y", "Position Z", "Velocity X", "Velocity Y", "Velocity Z", "Acceleration X", "Acceleration Y", "Acceleration Z"}
	writer.Write(header)

	// Write data
	for tick, stateMap := range states {
		for entity, state := range stateMap {
			record := []string{
				fmt.Sprintf("%d", tick),
				fmt.Sprintf("%d", entity),
				fmt.Sprintf("%f", state.Position.X),
				fmt.Sprintf("%f", state.Position.Y),
				fmt.Sprintf("%f", state.Position.Z),
				fmt.Sprintf("%f", state.Velocity.X),
				fmt.Sprintf("%f", state.Velocity.Y),
				fmt.Sprintf("%f", state.Velocity.Z),
				fmt.Sprintf("%f", state.Acceleration.X),
				fmt.Sprintf("%f", state.Acceleration.Y),
				fmt.Sprintf("%f", state.Acceleration.Z),
			}
			writer.Write(record)
		}
	}

	return nil
}
