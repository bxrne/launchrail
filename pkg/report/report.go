package report

import (
	"encoding/csv"
	"fmt"

	"github.com/bxrne/launchrail/pkg/entities"

	"os"

	"github.com/bxrne/launchrail/pkg/components"
)

// Generate generates a report from the ECS.
func Generate(ecs *entities.ECS, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"Entity", "Component Name", "Mass", "Position"}
	writer.Write(header)

	// Write data
	for entity := entities.Entity(1); entity < ecs.GetNextEntity(); entity++ {
		component := ecs.GetComponent(entity)
		if component != nil {
			record := []string{entity.String(),
				component.(*components.RocketComponent).Name,
				fmt.Sprintf("Mass: %f kg", component.(*components.RocketComponent).Mass),
				fmt.Sprintf("Pos: %f ?", component.(*components.RocketComponent).Position),
			}
			writer.Write(record)
		}
	}

	return nil
}
