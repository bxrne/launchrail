package openrocket

import (
	"encoding/xml"
	"fmt"
	"math"
)

// Parachute represents the parachute element of the XML document
type Parachute struct {
	XMLName          xml.Name         `xml:"parachute"`
	Name             string           `xml:"name"`
	ID               string           `xml:"id"`
	AxialOffset      AxialOffset      `xml:"axialoffset"`
	Position         Position         `xml:"position"`
	PackedLength     float64          `xml:"packedlength"`
	PackedRadius     float64          `xml:"packedradius"`
	RadialPosition   float64          `xml:"radialposition"`
	RadialDirection  float64          `xml:"radialdirection"`
	CD               string           `xml:"cd"` // WARN: May be 'auto' and num
	Material         Material         `xml:"material"`
	DeployEvent      string           `xml:"deployevent"`
	DeployAltitude   float64          `xml:"deployaltitude"`
	DeployDelay      float64          `xml:"deploydelay"`
	DeploymentConfig DeploymentConfig `xml:"deploymentconfiguration"`
	Diameter         float64          `xml:"diameter"`
	LineCount        int              `xml:"linecount"`
	LineLength       float64          `xml:"linelength"`
	LineMaterial     LineMaterial     `xml:"linematerial"`
}

// String returns full string representation of the parachute
func (p *Parachute) String() string {
	return fmt.Sprintf("Parachute{Name=%s, ID=%s, AxialOffset=%s, Position=%s, PackedLength=%.2f, PackedRadius=%.2f, RadialPosition=%.2f, RadialDirection=%.2f, CD=%s, Material=%s, DeployEvent=%s, DeployAltitude=%.2f, DeployDelay=%.2f, DeploymentConfig=%s, Diameter=%.2f, LineCount=%d, LineLength=%.2f, LineMaterial=%s}", p.Name, p.ID, p.AxialOffset.String(), p.Position.String(), p.PackedLength, p.PackedRadius, p.RadialPosition, p.RadialDirection, p.CD, p.Material.String(), p.DeployEvent, p.DeployAltitude, p.DeployDelay, p.DeploymentConfig.String(), p.Diameter, p.LineCount, p.LineLength, p.LineMaterial.String())
}

// GetMass calculates the mass of the parachute based on canopy and line materials/dimensions.
// NOTE: This calculation assumes Parachute.Material.Density represents AREAL density (e.g., kg/m^2).
// Line mass calculation uses LineMaterial density (assumed linear).
func (p *Parachute) GetMass() float64 {
	if p.Material.Density <= 0 || p.Diameter <= 0 {
		return 0.0 // Cannot calculate without material density and diameter
	}

	// Canopy Mass (assuming flat circle and areal density)
	canopyRadius := p.Diameter / 2.0
	canopyArea := math.Pi * canopyRadius * canopyRadius
	canopyMass := canopyArea * p.Material.Density // Density is assumed areal (mass/area)

	// Line Mass
	lineMass := p.LineLength * float64(p.LineCount) * p.LineMaterial.Density

	totalMass := canopyMass + lineMass

	if math.IsNaN(totalMass) || totalMass < 0 {
		fmt.Printf("Warning: Invalid mass (%.4f) calculated for Parachute '%s', returning 0.\n", totalMass, p.Name)
		return 0.0
	}

	return totalMass
}

// DeploymentConfig represents the deployment configuration element of the XML document
type DeploymentConfig struct {
	XMLName        xml.Name `xml:"deploymentconfiguration"`
	ConfigID       string   `xml:"configid,attr"`
	DeployEvent    string   `xml:"deployevent"`
	DeployAltitude float64  `xml:"deployaltitude"`
	DeployDelay    float64  `xml:"deploydelay"`
}

// String returns full string representation of the deployment clusterconfiguration
func (d *DeploymentConfig) String() string {
	return fmt.Sprintf("DeploymentConfig{ConfigID=%s, DeployEvent=%s, DeployAltitude=%.2f, DeployDelay=%.2f}", d.ConfigID, d.DeployEvent, d.DeployAltitude, d.DeployDelay)
}
