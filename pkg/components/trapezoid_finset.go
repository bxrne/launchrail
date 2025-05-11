package components

import (
	"fmt"
	"log"
	"math"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/pkg/openrocket"
	"github.com/bxrne/launchrail/pkg/types"
)

// TrapezoidFinset represents a set of trapezoidal fins on a rocket.
// All properties are for the entire set unless otherwise specified.
// Individual fin properties are used to calculate set properties.
type TrapezoidFinset struct {
	ecs.BasicEntity
	Name          string              // Name of the component
	RootChord     float64             // Root chord of a single fin
	TipChord      float64             // Tip chord of a single fin
	Span          float64             // Span (height) of a single fin
	SweepDistance float64             // Sweep distance (or length) of the leading edge of a single fin
	Thickness     float64             // Thickness of a single fin
	FinCount      int                 // Number of fins in the set
	Material      openrocket.Material // Material of the fins
	Position      types.Vector3       // Axial position of the fin set's attachment point (e.g., leading edge of root chord of fin 0)
	Mass          float64             // Total mass of the entire fin set (all fins)
	CenterOfMass  types.Vector3       // CM of the entire fin set, relative to rocket origin
	InertiaTensor types.Matrix3x3     // Inertia tensor of the entire fin set about its CM, aligned with rocket body axes
}

// --- Inertia Calculation Helper Functions (to be fully implemented) ---

// getRectangleInertia calculates area moments of inertia (Ixx, Iyy, Ixy) for a rectangle
// about its centroid. b = base (along x-axis), h = height (along y-axis).
func getRectangleInertia(base, height float64) (ixx, iyy, ixy float64) {
	ixx = base * math.Pow(height, 3) / 12.0
	iyy = height * math.Pow(base, 3) / 12.0
	ixy = 0 // For axes aligned with sides passing through centroid
	return
}

// getRightTriangleInertia calculates area moments of inertia (Ixx, Iyy, Ixy) for a right triangle
// about its centroid. b = base (along x-axis), h = height (along y-axis). Origin at right angle.
// Formulas from Calcresource: Ix=bh^3/36, Iy=hb^3/36, Ixy=-b^2h^2/72 (sign depends on quadrant)
func getRightTriangleInertia(base, height float64, orientationFactor float64) (ixx, iyy, ixy float64) {
	ixx = base * math.Pow(height, 3) / 36.0
	iyy = height * math.Pow(base, 3) / 36.0
	// Ixy is -b^2h^2/72 if origin at right angle vertex, axes along legs.
	// The sign depends on how the triangle is oriented relative to the fin's overall geometry.
	// orientationFactor could be +1 or -1 based on the triangle's position (e.g., leading/trailing edge section)
	ixy = orientationFactor * (math.Pow(base, 2) * math.Pow(height, 2)) / 72.0
	// Note: The standard formula is often -b^2h^2/72. If the triangle is in a different quadrant
	// for a composite shape, this sign might change or the dx, dy for parallel axis theorem will handle it.
	// For now, using a factor, but this needs careful geometric consideration.
	return
}

// --- End Helper Functions ---

// calculateAndSetCenterOfMass calculates the center of mass for the entire fin set
// and updates the CenterOfMass field.
// The Position field is assumed to be the attachment point of the fin set (e.g., leading edge of root chord of fin 0).
func (f *TrapezoidFinset) calculateAndSetCenterOfMass() {
	if (f.RootChord + f.TipChord) == 0 { // Avoid division by zero
		log.Printf("Warning: Sum of RootChord and TipChord is zero for finset. Cannot calculate CM. Defaulting CM to Position.")
		f.CenterOfMass = f.Position
		return
	}

	// Calculate the x-coordinate of a single fin's CG relative to its root chord's leading edge.
	// Formula: (SweepDistance * (RootChord + 2*TipChord) + (RootChord^2 + RootChord*TipChord + TipChord^2)) / (3 * (RootChord + TipChord))
	xCgLocalNum := (f.SweepDistance * (f.RootChord + 2*f.TipChord)) + (math.Pow(f.RootChord, 2) + f.RootChord*f.TipChord + math.Pow(f.TipChord, 2))
	xCgLocalDen := 3 * (f.RootChord + f.TipChord)
	xCgLocal := xCgLocalNum / xCgLocalDen

	// The CM of the fin set (assuming symmetrical placement)
	// X-coordinate is the attachment point's X + local fin CG's X.
	// Y and Z coordinates are assumed to be the same as the attachment point's Y and Z (typically 0 for centerline mounting).
	f.CenterOfMass.X = f.Position.X + xCgLocal
	f.CenterOfMass.Y = f.Position.Y
	f.CenterOfMass.Z = f.Position.Z
}

// calculateAndSetInertiaTensor calculates the inertia tensor for the entire fin set
// about its center of mass, aligned with the rocket body axes.
func (f *TrapezoidFinset) calculateAndSetInertiaTensor() {
	// --- Step 1 & 2: Single Fin Geometry and Mass ---
	if f.FinCount <= 0 {
		log.Printf("Error: FinCount is %d, cannot calculate inertia for finset '%s'. Setting to zero.", f.FinCount, f.Material.Name)
		f.InertiaTensor = types.Matrix3x3{}
		return
	}
	singleFinMass := f.Mass / float64(f.FinCount)
	finArea := (f.RootChord + f.TipChord) * f.Span / 2.0
	if finArea <= 1e-9 { // Avoid division by zero or near-zero area
		log.Printf("Error: FinArea is %f, cannot calculate inertia for finset '%s'. Setting to zero.", finArea, f.Material.Name)
		f.InertiaTensor = types.Matrix3x3{}
		return
	}
	massPerUnitArea := singleFinMass / finArea

	// --- Step 3: Decompose Fin Planform into Rectangle and Triangles (Revised) ---
	// Origin for local fin coordinates: Leading edge of the root chord.
	// X-axis along the root chord, Y-axis along the span.

	// --- Step 5: Overall Fin Planform's CM (local coordinates) ---
	// xFinCmLocal: Chordwise CM of the fin planform relative to its root chord leading edge.
	numXCm := f.RootChord*f.RootChord + f.RootChord*f.TipChord + f.TipChord*f.TipChord + f.SweepDistance*(f.RootChord+2*f.TipChord)
	denCm := 3 * (f.RootChord + f.TipChord)
	if denCm == 0 { // Should be caught by FinArea check, but as a safeguard
		log.Printf("Error: Sum of RootChord and TipChord is zero for finset '%s'. Setting inertia to zero.", f.Material.Name)
		f.InertiaTensor = types.Matrix3x3{}
		return
	}
	xFinCmLocal := numXCm / denCm

	// yFinCmLocal: Spanwise CM of the fin planform relative to its root chord.
	// Corrected formula for y_cm of a trapezoid: (h/3) * (b + 2a) / (b + a)
	// where h=Span, b=RootChord, a=TipChord
	if (f.RootChord + f.TipChord) == 0 { // Safeguard for yFinCmLocal denominator
		log.Printf("Error: Sum of RootChord and TipChord is zero for yFinCmLocal in finset '%s'. Setting inertia to zero.", f.Material.Name)
		f.InertiaTensor = types.Matrix3x3{}
		return
	}
	yFinCmLocal := (f.Span / 3.0) * (f.RootChord + 2.0*f.TipChord) / (f.RootChord + f.TipChord)

	var totalIxxFinAreaCm, totalIyyFinAreaCm, totalIxyFinAreaCm float64

	// Component 1: Leading Edge Triangle (if SweepDistance > 0)
	// Covers the area from x=0 to x=SweepDistance
	if f.SweepDistance > 1e-9 {
		leTriBase := f.SweepDistance
		leTriHeight := f.Span
		leTriArea := 0.5 * leTriBase * leTriHeight
		// Vertices: (0,0), (Ls, Span), (Ls,0). Centroid relative to fin local origin (0,0): (Ls*2/3, Span/3)
		leTriCmxLocal := leTriBase * 2.0 / 3.0
		leTriCmyLocal := leTriHeight / 3.0
		leTriIxxCm, leTriIyyCm, leTriIxyCm := getRightTriangleInertia(leTriBase, leTriHeight, 1.0) // Orientation factor +1.0 assuming Ixy = -b^2h^2/72 for right angle at origin, legs along +axes

		dxLeTri := leTriCmxLocal - xFinCmLocal
		dyLeTri := leTriCmyLocal - yFinCmLocal
		totalIxxFinAreaCm += leTriIxxCm + leTriArea*dyLeTri*dyLeTri
		totalIyyFinAreaCm += leTriIyyCm + leTriArea*dxLeTri*dxLeTri
		totalIxyFinAreaCm += leTriIxyCm + leTriArea*dxLeTri*dyLeTri
	}

	// Component 2: Middle Rectangle (if TipChord > 0)
	// Covers the area from x=SweepDistance to x=SweepDistance+TipChord
	if f.TipChord > 1e-9 {
		rectBase := f.TipChord // This is the width of the rectangle
		rectHeight := f.Span
		rectArea := rectBase * rectHeight
		// CM relative to fin local origin (0,0): (Ls + TipChord/2, Span/2)
		rectCmxLocal := f.SweepDistance + rectBase/2.0
		rectCmyLocal := rectHeight / 2.0
		rectIxxCm, rectIyyCm, _ := getRectangleInertia(rectBase, rectHeight) // Ixy for rectangle about its CM is 0

		dxRect := rectCmxLocal - xFinCmLocal
		dyRect := rectCmyLocal - yFinCmLocal
		totalIxxFinAreaCm += rectIxxCm + rectArea*dyRect*dyRect
		totalIyyFinAreaCm += rectIyyCm + rectArea*dxRect*dxRect
		totalIxyFinAreaCm += 0 + rectArea*dxRect*dyRect // rectIxyCm is 0 for rectangle CM
	}

	// Component 3: Trailing Edge Triangle
	// Covers the area from x=SweepDistance+TipChord to x=RootChord
	teTriFinBase := f.RootChord - (f.SweepDistance + f.TipChord)
	if teTriFinBase > 1e-9 {
		teTriHeight := f.Span
		teTriArea := 0.5 * teTriFinBase * teTriHeight
		// Vertices: (Ls+a,0), (RootChord,0), (Ls+a,Span). Right angle at (Ls+a, 0).
		// Centroid relative to fin local origin (0,0): ( (Ls+a) + (Ls+a) + RootChord )/3, Span/3
		teTriCmxLocal := (2.0*(f.SweepDistance+f.TipChord) + f.RootChord) / 3.0
		teTriCmyLocal := teTriHeight / 3.0
		teTriIxxCm, teTriIyyCm, teTriIxyCm := getRightTriangleInertia(teTriFinBase, teTriHeight, 1.0) // Orientation factor +1.0

		dxTeTri := teTriCmxLocal - xFinCmLocal
		dyTeTri := teTriCmyLocal - yFinCmLocal
		totalIxxFinAreaCm += teTriIxxCm + teTriArea*dyTeTri*dyTeTri
		totalIyyFinAreaCm += teTriIyyCm + teTriArea*dxTeTri*dxTeTri
		totalIxyFinAreaCm += teTriIxyCm + teTriArea*dxTeTri*dyTeTri
	}

	// --- Step 8: Convert to Mass Moments for Single Fin (Thin Plate) ---
	iXxSingleFinMassCm := massPerUnitArea * totalIxxFinAreaCm
	iYySingleFinMassCm := massPerUnitArea * totalIyyFinAreaCm
	// For a thin plate, Izz = Ixx + Iyy. Assuming fin lies in XY plane of its local CM.
	iZzSingleFinMassCm := iXxSingleFinMassCm + iYySingleFinMassCm
	iXySingleFinMassCm := massPerUnitArea * totalIxyFinAreaCm
	iZxSingleFinMassCm := 0.0 // Product of inertia involving z-axis is zero for thin plate in xy plane
	iYzSingleFinMassCm := 0.0

	// Local Inertia Tensor for ONE fin, about its own CM, aligned with fin's local x,y,z axes
	// (x along chord, y along span, z normal to fin surface)
	// Tensor components are: [[Ixx, -Ixy, -Izx], [-Ixy, Iyy, -Iyz], [-Izx, -Iyz, Izz]]
	localFinInertiaTensorCm := types.Matrix3x3{
		M11: iXxSingleFinMassCm, M12: -iXySingleFinMassCm, M13: -iZxSingleFinMassCm,
		M21: -iXySingleFinMassCm, M22: iYySingleFinMassCm, M23: -iYzSingleFinMassCm,
		M31: -iZxSingleFinMassCm, M32: -iYzSingleFinMassCm, M33: iZzSingleFinMassCm,
	}

	// --- Step 9: Rotate and Sum for All Fins in the Set ---
	// Initialize the total inertia tensor for the finset to zero.
	f.InertiaTensor = types.Matrix3x3{} // Zero matrix

	// Fin thickness contribution to Ixx and Iyy (treating fin as a cuboid for this part)
	// This is an approximation often used. For a thin plate, I_thickness = m * t^2 / 12 about axis in plane.
	// I_local_fin_cm.M11 (Ixx) += SingleFinMass * f.Thickness * f.Thickness / 12.0 // About chord axis
	// I_local_fin_cm.M22 (Iyy) += SingleFinMass * f.Thickness * f.Thickness / 12.0 // About span axis
	// I_local_fin_cm.M33 (Izz) is sum of these for thin plate. If we add thickness terms here, it becomes more complex.
	// For now, stick to thin plate assumption (Izz = Ixx + Iyy based on area moments).
	// More accurate would be to use cuboid formulas if thickness is significant, but this is a common simplification.

	// The rocket's X-axis is longitudinal (roll).
	// The fin's local x-axis is chordwise, y-axis is spanwise, z-axis is normal to fin surface.
	// Fin 0 is typically aligned with the rocket's +Y axis (or +Z if different convention).

	// Attachment point of the finset (e.g. leading edge of root chord of fin 0) is f.Position
	// CM of a single fin in its own local 2D planform coordinates (origin at LE root): (xFinCmLocal, yFinCmLocal)

	totalInertia := types.Matrix3x3{}
	for i := 0; i < f.FinCount; i++ {
		finAngle := (2.0 * math.Pi / float64(f.FinCount)) * float64(i)

		// Rotation Matrix Ri: Transforms from fin's local axes to rocket body axes.
		// Fin's local x (chord) aligns with rocket's X (longitudinal).
		// Fin's local y (span) rotates in rocket's YZ plane.
		// Fin's local z (normal) also rotates in rocket's YZ plane.
		// This means we rotate about the rocket's X-axis by finAngle.
		cosAngle := math.Cos(finAngle)
		sinAngle := math.Sin(finAngle)

		// Rotation matrix for rotation around X-axis
		ri := types.Matrix3x3{
			M11: 1, M12: 0, M13: 0,
			M21: 0, M22: cosAngle, M23: -sinAngle,
			M31: 0, M32: sinAngle, M33: cosAngle,
		}

		// Transform local fin inertia tensor to rocket body axes (still about fin's own CM)
		// I_fin_body_axes_cm = Ri * localFinInertiaTensorCm * Ri_transpose
		riT := ri.Transpose()
		iFinBodyAxesCmPtr := ri.MultiplyMatrix(&localFinInertiaTensorCm).MultiplyMatrix(riT) // Corrected: &RiT -> RiT
		iFinBodyAxesCm := *iFinBodyAxesCmPtr

		// Calculate displacement vector d_i from finset's CM (f.CenterOfMass) to this fin's CM.
		// CM of this fin (cm_fin_i) in rocket body coordinates:
		// The fin's own CM (xFinCmLocal, yFinCmLocal) is in the fin's local coordinate system.
		// We need to rotate this local CM vector by R_i to get it into body axes,
		// then add the offset of the fin's root leading edge from the rocket's reference point (usually nose tip).
		// Finally, we add the finset's CG position.
		cmFinLocal := types.Vector3{X: xFinCmLocal, Y: yFinCmLocal, Z: 0} // Z=0 in fin's own 2D planform coords
		cmFinBodyAxes := ri.MultiplyVector(&cmFinLocal)

		// d_x, d_y, d_z are components of the vector from finset CG to fin_i CG in rocket body axes
		dX := f.Position.X + cmFinBodyAxes.X - f.CenterOfMass.X
		dY := f.Position.Y + cmFinBodyAxes.Y - f.CenterOfMass.Y // Corrected: removed extra *cosAngle
		dZ := f.Position.Z + cmFinBodyAxes.Z - f.CenterOfMass.Z // Corrected: use cmFinBodyAxes.Z

		// Parallel Axis Theorem: I_total_fin_i = I_fin_body_axes_cm + M_fin * ( (d . d)E - d (outer_product) d )
		dSq := dX*dX + dY*dY + dZ*dZ // d . d

		term2DiagXx := singleFinMass * (dSq - dX*dX)
		term2DiagYy := singleFinMass * (dSq - dY*dY)
		term2DiagZz := singleFinMass * (dSq - dZ*dZ)
		term2OffdiagXy := -singleFinMass * dX * dY
		term2OffdiagXz := -singleFinMass * dX * dZ
		term2OffdiagYz := -singleFinMass * dY * dZ

		iTotalFinI := types.Matrix3x3{
			M11: iFinBodyAxesCm.M11 + term2DiagXx,
			M12: iFinBodyAxesCm.M12 + term2OffdiagXy,
			M13: iFinBodyAxesCm.M13 + term2OffdiagXz,
			M21: iFinBodyAxesCm.M21 + term2OffdiagXy, // M21 = M12
			M22: iFinBodyAxesCm.M22 + term2DiagYy,
			M23: iFinBodyAxesCm.M23 + term2OffdiagYz,
			M31: iFinBodyAxesCm.M31 + term2OffdiagXz, // M31 = M13
			M32: iFinBodyAxesCm.M32 + term2OffdiagYz, // M32 = M23
			M33: iFinBodyAxesCm.M33 + term2DiagZz,
		}
		totalInertia = totalInertia.Add(iTotalFinI) // Corrected: pass value, not pointer
	}

	f.InertiaTensor = totalInertia
	log.Printf("SUCCESS: Inertia tensor for finset component (name: %s) calculated. Tensor: %+v", f.Name, f.InertiaTensor)
}

// GetMass returns the total mass of the finset
func (f *TrapezoidFinset) GetMass() float64 {
	return f.Mass
}

// GetPlanformArea returns the planform area of a single fin in the set.
func (f *TrapezoidFinset) GetPlanformArea() float64 {
	if f.Span <= 0 || (f.RootChord <= 0 && f.TipChord <= 0) {
		return 0.0
	}
	return (f.RootChord + f.TipChord) * f.Span / 2.0
}

// calculateAndSetMass calculates and sets the mass of the finset.
// TODO: Implement the actual mass calculation logic.
func (f *TrapezoidFinset) calculateAndSetMass() {
	// Placeholder: Mass calculation logic will go here.
	// For now, let's assume it might use GetPlanformArea, Thickness, Material.Density, FinCount
	planformArea := f.GetPlanformArea()
	if f.Material.Density > 0 && planformArea > 0 && f.Thickness > 0 && f.FinCount > 0 {
		f.Mass = planformArea * f.Thickness * f.Material.Density * float64(f.FinCount)
	} else {
		f.Mass = 0
		log.Printf("Warning: Could not calculate mass for finset '%s' due to zero/negative inputs (Area: %f, Thick: %f, Density: %f, Count: %d). Setting mass to 0.",
			f.Name, planformArea, f.Thickness, f.Material.Density, f.FinCount)
	}
}

// NewTrapezoidFinsetFromORK creates a new TrapezoidFinset component from OpenRocket data.
// TODO: This is a basic stub and needs to be fully implemented based on how
// openrocket.TrapezoidFinset fields map to the simulation's TrapezoidFinset.
func NewTrapezoidFinsetFromORK(orkFinset *openrocket.TrapezoidFinset, orkPosition types.Vector3, orkMaterial openrocket.Material) (*TrapezoidFinset, error) {
	if orkFinset == nil {
		return nil, fmt.Errorf("NewTrapezoidFinsetFromORK: provided OpenRocket finset is nil")
	}

	simFinset := &TrapezoidFinset{
		Name:          orkFinset.Name,
		FinCount:      orkFinset.FinCount,
		RootChord:     orkFinset.RootChord,
		TipChord:      orkFinset.TipChord,
		Span:          orkFinset.Height, // ORK 'height' is fin span
		SweepDistance: orkFinset.SweepLength,
		Thickness:     orkFinset.Thickness, // Assuming ORK provides this directly, might need calculation if it's 'thicknessfraction'
		Material:      orkMaterial,
		// Position, Mass, CenterOfMass, InertiaTensor are calculated by methods
	}

	// Set Position from ORK data (this needs careful interpretation of ORK position)
	// For now, let's assume orkPosition is the attachment point (e.g., LE of root chord of fin 0)
	simFinset.Position = orkPosition

	// Calculate mass, CM, and Inertia
	simFinset.calculateAndSetMass()
	simFinset.calculateAndSetCenterOfMass()
	simFinset.calculateAndSetInertiaTensor()

	if simFinset.Mass == 0 && orkFinset.Material.Density > 0 && simFinset.GetPlanformArea() > 0 && simFinset.Thickness > 0 && simFinset.FinCount > 0 {
		log.Printf("Warning: Mass for finset '%s' calculated as 0. Check calculateAndSetMass and ORK data. ORK Density: %f, Area: %f, Thick: %f, Count: %d",
			simFinset.Name, orkFinset.Material.Density, simFinset.GetPlanformArea(), simFinset.Thickness, simFinset.FinCount)
	}

	return simFinset, nil
}

// GetPosition returns the finset's reference position (attachment point) in rocket coordinates.
func (fs *TrapezoidFinset) GetPosition() types.Vector3 {
	return fs.Position
}

// GetCenterOfMassLocal returns the finset's center of mass relative to its Position (attachment point).
func (fs *TrapezoidFinset) GetCenterOfMassLocal() types.Vector3 {
	// fs.CenterOfMass is stored as the global CG of the finset.
	// To get local CG relative to fs.Position, we subtract fs.Position.
	// This requires Vector3.Subtract() to be implemented.
	// For now, assuming it exists and is called as fs.CenterOfMass.Subtract(fs.Position)
	// If Vector3.Subtract() is not available, this will need adjustment or direct calculation.
	// Placeholder if Subtract isn't ready, but the logic is to make it local.
	// return fs.CenterOfMass.Subtract(fs.Position) // Ideal

	// Temporary direct calculation assuming X is the primary axis of displacement from position
	// This depends on how fs.CenterOfMass and fs.Position are defined and calculated initially.
	// Reviewing NewTrapezoidFinsetFromORK: fs.CenterOfMass = fs.Position.Add(cmLocal)
	// So, cmLocal = fs.CenterOfMass.Subtract(fs.Position)
	// This requires fs.CenterOfMass to be correctly populated before this call.

	// Assuming fs.CenterOfMass = {GlobalX, GlobalY, GlobalZ}
	// and fs.Position = {AttachX, AttachY, AttachZ}
	// LocalCG = {GlobalX - AttachX, GlobalY - AttachY, GlobalZ - AttachZ}
	// This is essentially fs.CenterOfMass.Subtract(fs.Position)
	// Until Vector3 operations are added, this will be a conceptual placeholder:
	return types.Vector3{
		X: fs.CenterOfMass.X - fs.Position.X,
		Y: fs.CenterOfMass.Y - fs.Position.Y,
		Z: fs.CenterOfMass.Z - fs.Position.Z,
	}
}

// GetInertiaTensorLocal returns the finset's inertia tensor about its own CG, aligned with rocket axes.
func (fs *TrapezoidFinset) GetInertiaTensorLocal() types.Matrix3x3 {
	// The comment for fs.InertiaTensor states it's "of the entire fin set about its CM".
	// This is assumed to be what's needed for the local inertia tensor in aggregation.
	return fs.InertiaTensor
}
