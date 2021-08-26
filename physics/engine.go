// Package physics serves as the backend for GGGG, storing a list of particles, physics constants, settings/states, etc.
// and providing the methods which perform the physics calculations to update particle velocities and positions (which
// are called iteratively/repeatedly via the main app physics loop).
package physics

// Engine is the EngineData instance, effectively the physics engine instance.
// Particle objects use the fields of this struct instance. To control the behavior of the physics engine, set the
// fields of this instance (via a pointer if desired). Do not create any other objects of this type (you will not be
// able to effect the behavior of the engine using them).
var Engine EngineData

// EngineData is the type for Engine. DO NOT create any other instances of this type. This type is exported so that
// pointers to Engine can be created outside this package (e.g. as a field in state.Data).
type EngineData struct {
	// GravityStrength is the gravitational constant, essentially (acts on Mass)
	GravityStrength float64 `json:"gravity_strength"`
	// CloseChargeStrength is the Coulomb constant, essentially (acts on CloseCharge)
	CloseChargeStrength float64 `json:"close_charge_strength"`
	// FarChargeStrength is the Coulomb constant, essentially (acts on FarCharge)
	FarChargeStrength float64 `json:"far_charge_strength"`

	// EnvironmentSize is the quantized size of the environment (relative to particle size, which is determined by mass)
	EnvironmentSize int `json:"environment_size"`
	// AllowMerge determines whether particles may merge when the collide. If disabled, particles always bounce. If
	// enabled, they may merge or bounce depending on their relative masses and close charges.
	AllowMerge bool `json:"allow_merge"`
	// WallBounce determines whether particles bounce off the "walls" of the environment (or more accurately, whether
	// the environment - as represented here in the physics engine and particle positions - is bounded by
	// EnvironmentSize or is unbounded)
	WallBounce bool `json:"wall_bounce"`

	// bounceCompleteDistFactor is used to determine when a particle bounce is complete (so forces don't get
	// exceptionally large when particles get very close to each other)
	bounceCompleteDistFactor float64
	// mergeMassRatioThreshold is the ratio between particle masses above which particles may merge if not overridden
	// by close charge repulsion
	mergeMassRatioThreshold float64
	// mergeCloseChargeThreshold is the summed (same sign) close charge of two particles above which particles cannot
	// merge. If particles have opposite sign close charges, they are allowed to merge if AllowMerge is true and one is
	// sufficiently larger than the other.
	mergeCloseChargeThreshold float64

	// Particles is the slice of particles the physics engine acts on.
	Particles []*Particle `json:"particles"`
	// initialParticles is used to reset particles to their original state
	initialParticles []*Particle
}

// Initialize initializes the physics Engine and sets all default values (call before setting any Engine field values).
// Does NOT initialize Particles.
// Presently, *only* sets default values, but a it's good idea to call it even if you're initializing all values,
// in case other logic is added in future.
func (*EngineData) Initialize() {
	Engine.GravityStrength = 15
	Engine.CloseChargeStrength = 150000000
	Engine.FarChargeStrength = 7.5

	Engine.EnvironmentSize = 800
	Engine.AllowMerge = true
	Engine.WallBounce = true

	Engine.bounceCompleteDistFactor = 1.5
	Engine.mergeMassRatioThreshold = 2.5
	Engine.mergeCloseChargeThreshold = 0.25
}
