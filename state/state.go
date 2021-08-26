// Package state is used to make the Data struct available to both the main app and the guis package.
package state

import (
	"GoGoGadgetGravity/physics"
)

// Data is the primary struct for GGGG, used by the main app and the guis package to hold state information.
type Data struct {
	// PhysicsEngine is a pointer to the physics.Engine variable (single physics.EngineData instance)
	PhysicsEngine *physics.EngineData `json:"physics_engine"`
	// NumberOfParticles is the (desired) number of physics.Engine.Particles
	NumberOfParticles int `json:"number_of_particles"`
	// AverageMass is the desired average mass of physics.Engine.Particles to be generated
	AverageMass int `json:"average_mass"`
	// HistoryTrail indicates whether physics.Particle position histories are being tracked/displayed
	HistoryTrail bool `json:"history_trail"`
	// HistoryLength is the number of previous physics.Particle positions stored/displayed
	HistoryLength int `json:"history_length"`
	// PhysicsLoopSpeed is the frequency with which the simulation is updated, in milliseconds. Essentially, how often
	// physics.UpdateParticles is called.
	PhysicsLoopSpeed int `json:"physics_loop_speed"`
}
