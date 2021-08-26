package physics

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/atedja/go-vector"
)

// particleData is part of the Particle struct and is used for fields which are or may be serialized.
// See https://stackoverflow.com/a/11129474/5061881 for the pattern used here to
// deal with serializing non-exported / non-directly-accessible fields, while also allowing serializing
// exported / accessible fields, and having exported non-serialized field & non-exported non-serialized fields.
// The pattern: anything to be serialized is in this struct and is exported (capitalized) and if it needs to be
// accessible from outside this package it needs a getter/setter. It also includes non-serialized fields (non-exported)
// which might conceivably be serialized (which then requires a simply rename refactor) - basically it holds all fields
// that are fundamental & (semi)permanent.
// This struct is then embedded in the other, main struct which also receives all methods, and the main struct has any
// fields that *definitely* won't be serialized (to make future changes less likely, it might make sense to have the
// main struct completely empty except the embedded struct and simply have ALL non-serialized members be non-exported
// and if they need to be accessible give them getters/setters, but I like having known transient / runtime fields in
// the main struct (or more to the point, NOT in the state struct).
// The main struct then needs to implement the json.Marshaler and json.Unmarshaler interfaces by simply returning the
// results of json.Marshal/Unmarshal on this struct.
type particleData struct {
	// Gravity is inversely proportional to distance^2.
	// It is always positive and therefore attractive.
	// Masses add. Radius is proxy.
	Mass float64 `json:"mass"`
	// closeCharge is inversely proportional to distance^3.
	// It may be negative or positive and therefore repulsive or attractive
	// Charges average. Red (negative) and green (positive) are proxy (zero is black),
	// with charge min/max +/- 1.
	CloseCharge float64 `json:"close_charge"`
	// farCharge is *proportional* to distance.
	// It is always positive and therefore attractive.
	// Charges average. Alpha is proxy with charge range  0-1.
	FarCharge float64       `json:"far_charge"`
	Position  vector.Vector `json:"position"`
	Velocity  vector.Vector `json:"velocity"`

	// trackHistory indicates whether the previous position should be stored in positionHistory
	// during Particle.UpdatePosition.
	trackHistory bool
	// historySize is the FIFO length of positionHistory
	historySize int
	// positionHistory is the slice of previous positions of the Particle.
	positionHistory []vector.Vector
}

// The Particle struct holds particle data. It is composed of the particleData struct which holds the more fundamental
// fields of the particle, and those which are or may be serialized, along with secondary/calculated fields (radius and
// colors used for display and calculated from mass and charges), and ephemeral merge and bounce states.
type Particle struct {
	// particleData embeds the particleData struct. It is named rather than anonymous so that any exported fields (which
	// aren't already masked by getters of the same name) aren't accessible externally.
	particleData particleData
	// Radius is a proxy for mass. Updated with SetMass.
	Radius int
	// R (red) & G (green) are proxies for closeCharge. Updated with SetCloseCharge.
	R, G uint8
	// A (alpha) is proxy for farCharge (min 64). Updated with SetFarCharge.
	A uint8

	// merging indicates whether the particle is currently merging with one or more other particle(s)
	merging bool
	// MergingWith is a "list" of particles this particle is currently merging with. It is a map of empty structs
	// instead of a slice to facilitate deleting members during iteration through the map and for efficient lookup of
	// whether a particle is in the list.
	MergingWith map[*Particle]struct{}
	// bouncing indicates whether the particle is currently bouncing against another (application of forces is suspended
	// until the bounce completes, as determined by EngineData.bounceCompleteDistFactor).
	bouncing bool
	// bouncingAgainst is the particle which this particle is currently bouncing against (if any / if bouncing is true).
	bouncingAgainst *Particle
}

//region Creation & Initialization

// NewParticle is a factory for creating a new, basic Particle (without a velocity, history info, etc.).
func NewParticle(mass, closeCharge, farCharge, x, y float64) *Particle {
	p := &Particle{particleData: particleData{
		Position: vector.NewWithValues([]float64{x, y}),
		Velocity: vector.New(2)}}

	p.initializeWithValues(mass, closeCharge, farCharge)

	return p
}

// Clone creates a copy of Particle p.
func (p *Particle) Clone() *Particle {
	// NewParticle is used to ensure the copy is properly created and initialized (and so that non-exported values,
	// such as Radius, are copied).
	c := NewParticle(p.Mass(), p.CloseCharge(), p.FarCharge(), p.Position()[0], p.Position()[1])
	// Velocity is not set by NewParticle, so we set it here to complete the copy.
	c.SetVelocity(p.Velocity())
	return c
}

// initialize is used to initialize a particle; see initializeWithValues.
func (p *Particle) initialize() {
	// Assumes the particle already has properties set (but needs proxies set) - e.g. because created by deserialization
	p.initializeWithValues(p.Mass(), p.CloseCharge(), p.FarCharge())
}

// initializeWithValues is used to initialize particles, that is to calculate proxy values (e.g. radius)
// and make the (empty) particleData.positionHistory and MergingWith "lists"
func (p *Particle) initializeWithValues(mass, closeCharge, farCharge float64) {
	// Use setters so the proxies get initialized
	p.SetMass(mass)
	p.SetCloseCharge(closeCharge)
	p.SetFarCharge(farCharge)

	p.particleData.positionHistory = make([]vector.Vector, 0, 0)

	p.MergingWith = make(map[*Particle]struct{})
}

//endregion Creation & Initialization

//region Serialization and Stringification

// MarshalJSON implements json.Marshaler in order to allow serializing using the non-exported struct.
func (p *Particle) MarshalJSON() ([]byte, error) {
	return json.Marshal(&p.particleData)
}

// UnmarshalJSON implements json.Unmarshaler in order to allow serializing using the non-exported struct.
func (p *Particle) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &p.particleData)
}

// String gets a string representation of the Particle, which is more verbose / plain English than string(particle) but
// does not include every field.
func (p *Particle) String() string {
	return fmt.Sprintf("{Position: %v; Velocity: %v, Mass: %f, Close Charge: %f, Far Charge: %f}",
		p.Position(), p.Velocity(), p.Mass(), p.CloseCharge(), p.FarCharge())
}

// ShortString gets a compact string representation of the most relevant Particle fields, without any labels and with
// values rounded.
func (p *Particle) ShortString() string {
	return strings.ReplaceAll(strings.ReplaceAll(
		regexp.MustCompile(`\s+`).ReplaceAllString(fmt.Sprintf("{%-6.1f; %-6.3v; %-6.3v}",
			p.Mass(), p.Position(), p.Velocity()), " "),
		" ]", "]"), " ;", ";")
}

//endregion Serialization and Stringification

//region Mass (gravity)

// Mass gets the mass.
func (p *Particle) Mass() float64 {
	return p.particleData.Mass
}

// SetMass sets the mass and updates the proxy Radius.
func (p *Particle) SetMass(mass float64) {
	p.particleData.Mass = mass
	//p.Radius = int(math.Max(math.Round(math.Sqrt(mass) / math.SqrtPi), 1))
	p.Radius = int(math.Max(math.Round(math.Sqrt(mass)/(2*math.SqrtPi)), 1))
}

//endregion Mass (gravity)

//region CloseCharge

// CloseCharge gets the closeCharge.
func (p *Particle) CloseCharge() float64 {
	return p.particleData.CloseCharge
}

// SetCloseCharge sets the closeCharge and updates the proxies R (red) & G (green).
func (p *Particle) SetCloseCharge(closeCharge float64) {
	closeCharge = math.Max(-1, math.Min(closeCharge, 1))
	p.particleData.CloseCharge = closeCharge

	// Negative charge is red, the closer to -1 the more red (0 is black)
	if closeCharge < 0 {
		// We don't want to round here, it's expensive and unnecessary
		p.R = uint8(255.0 * math.Abs(closeCharge))
		p.G = 0
		// Positive charge is green, the closer to 1 the more green (0 is black)
	} else {
		p.R = 0
		// We don't want to round here, it's expensive and unnecessary
		p.G = uint8(255.0 * math.Abs(closeCharge))
	}
}

//endregion CloseCharge

//region FarCharge

// FarCharge gets the farCharge.
func (p *Particle) FarCharge() float64 {
	return p.particleData.FarCharge
}

// SetFarCharge sets the farCharge.
func (p *Particle) SetFarCharge(farCharge float64) {
	farCharge = math.Max(0, math.Min(farCharge, 1))
	p.particleData.FarCharge = farCharge

	// Alpha range 64 - 255 (we don't want 0 charge to be fully transparent, we want to always be able to see particles)
	p.A = uint8(207*math.Abs(farCharge)) + 48
}

//endregion FarCharge

//region Position

// Position gets the Position
func (p *Particle) Position() vector.Vector {
	return p.particleData.Position
}

// SetPosition sets the Position
func (p *Particle) SetPosition(position vector.Vector) {
	p.particleData.Position = position
}

// UpdatePosition adds the velocity to the current position
func (p *Particle) UpdatePosition() {
	if p.particleData.trackHistory {
		p.particleData.positionHistory = append(p.particleData.positionHistory, p.Position())
		// If longer than historySize, truncate it (remove from end since it's FIFO)
		if len(p.particleData.positionHistory) > p.particleData.historySize {
			p.particleData.positionHistory = p.particleData.positionHistory[1:]
		}
	}
	p.SetPosition(vector.Add(p.Position(), p.Velocity()))
}

//endregion Position

//region Velocity

// Velocity gets the Velocity
func (p *Particle) Velocity() vector.Vector {
	return p.particleData.Velocity
}

// SetVelocity sets the Velocity
func (p *Particle) SetVelocity(velocity vector.Vector) {
	p.particleData.Velocity = velocity
}

//endregion Velocity

//region trackHistory

// TrackHistory gets the trackHistory
func (p *Particle) TrackHistory() bool {
	return p.particleData.trackHistory
}

// SetTrackHistory sets the trackHistory
func (p *Particle) SetTrackHistory(trackHistory bool) {
	p.particleData.trackHistory = trackHistory
}

//endregion trackHistory

//region historySize

// HistorySize gets the historySize
func (p *Particle) HistorySize() int {
	return p.particleData.historySize
}

// SetHistorySize sets the historySize
func (p *Particle) SetHistorySize(historySize int) {
	p.particleData.historySize = historySize
}

//endregion historySize

//region positionHistory

// PositionHistory gets the positionHistory
func (p *Particle) PositionHistory() []vector.Vector {
	return p.particleData.positionHistory
}

// SetPositionHistory sets the (entire) positionHistory
func (p *Particle) SetPositionHistory(positionHistory []vector.Vector) {
	p.particleData.positionHistory = positionHistory
}

//endregion positionHistory
