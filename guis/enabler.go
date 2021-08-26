// Package guis provides an interface (along with supporting struct(s) etc.) for GUIs to implement to meet the basic
// requirements to display and control GoGoGadgetGravity particle simulations.
package guis

import (
	"GoGoGadgetGravity/physics"
	"GoGoGadgetGravity/state"
)

// GUIInitializationData holds state (provided by main) for GUI creation/initialization (GUIEnabler.CreateGUI) and
// reloading (when loading state from file with GUIEnabler.LoadState). This includes particles to be drawn in their
// initial state - future updates to particles are requested by main, with particle state provided, using the
// GUIEnabler.DrawParticles method.
type GUIInitializationData struct {
	*state.Data

	// WinMinWidth is the minimum (and typically initial) GUI window width
	WinMinWidth int
	// WinMinHeight is the minimum (and typically initial) GUI window height
	WinMinHeight int
}

// GUIEnabler is an interface for GUIs to implement to meet the basic requirements to display and control
// GoGoGadgetGravity particle simulations.
type GUIEnabler interface {
	// CreateGUI instructs the GUI to create it's window, initialize control states/values,
	// and draw the initial slice of particles.
	// Expected to wait until the GUI window is closed.
	CreateGUI(initialValues GUIInitializationData)
	// LoadState sets gui control states/values & draws particles (e.g. to values from loading state from file).
	LoadState(initialValues GUIInitializationData)
	// SetPhysicsLoopSpeed instructs the GUI that the main program has changed the physics loop speed (because the
	// requested loop is too quick), so the GUI can adjust its control position/value.
	SetPhysicsLoopSpeed(loopTime int)
	// SetStatusText instructs the GUI to print the requested string in its status text control.
	SetStatusText(text string, time int)

	// DrawParticles instructs the GUI to draw the particles within its display area.
	DrawParticles(particles []*physics.Particle)
	// UpdateView instructs the GUI to redraw the entire environment / recreate its display, such as when the
	// EnvironmentSize is changed.
	UpdateView(particles []*physics.Particle)

	// ConnectSaveStateEvent provides the GUI with the function to call when the user uses the GUI to request saving
	// the current state to file.
	// The GUI is expected to provide a file picker, and then call this function, passing it the file path/name.
	ConnectSaveStateEvent(func(file string))
	// ConnectLoadStateEvent provides the GUI with the function to call when the user uses the GUI to request loading
	// a saved state from file.
	// The GUI is expected to provide a file picker, and then call this function, passing it the file path/name.
	ConnectLoadStateEvent(func(file string))
	// ConnectEnvironmentSizeChangedEvent provides the GUI with the function to call when the user uses the GUI to
	// request an environment size change.
	// The GUI is expected to resize/redraw its display area and then call this function, passing it the new size.
	// Particles will be generated and GUI instructed to draw them if currently paused.
	ConnectEnvironmentSizeChangedEvent(func(value int))
	// ConnectNumParticlesChangedEvent provides the GUI with the function to call when the user uses the GUI to request
	// a change in the number of (to be generated) particles.
	// The GUI is expected to change its state accordingly and then call this function, passing it the new number
	// of particles. Particles will be generated and GUI instructed to draw them if currently paused.
	ConnectNumParticlesChangedEvent(func(value int))
	// ConnectAverageMassChangedEvent provides the GUI with the function to call when the user uses the GUI to request
	// a change in the average mass of (to be generated) particles.
	// The GUI is expected to change its state accordingly and then call this function, passing it the new average mass.
	// Particles will be generated and GUI instructed to draw them if currently paused.
	ConnectAverageMassChangedEvent(func(value int))
	// ConnectRegenParticlesEvent provides the GUI with the function to call when the user uses the GUI to request
	// new particles be generated.
	// The GUI is expected to call this method, which will generate new particles and instruct the GUI to draw them.
	ConnectRegenParticlesEvent(func())
	// ConnectGravityStrengthChangedEvent provides the GUI with the function to call when the user uses the GUI to
	// request a change in the physics engine gravity strength.
	// The GUI is expected to change its state accordingly and then call this function, passing it the new gravity
	// strength.
	ConnectGravityStrengthChangedEvent(func(value float64))
	// ConnectCloseChargeStrengthChangedEvent provides the GUI with the function to call when the user uses the GUI to
	// request a change in the physics engine "close charge" strength.
	// The GUI is expected to change its state accordingly and then call this function, passing it the new close charge
	// strength.
	ConnectCloseChargeStrengthChangedEvent(func(value float64))
	// ConnectFarChargeStrengthChangedEvent provides the GUI with the function to call when the user uses the GUI to
	// request a change in the physics engine "far charge" strength.
	// The GUI is expected to change its state accordingly and then call this function, passing it the new far charge
	// strength.
	ConnectFarChargeStrengthChangedEvent(func(value float64))
	// ConnectAllowMergeChangedEvent provides the GUI with the function to call when the user uses the GUI to request
	// particle mergers be enabled/disabled.
	// The GUI is expected to change its state accordingly and then call this function, passing it a bool indicating
	// whether particle mergers should presently be allowed/disallowed.
	ConnectAllowMergeChangedEvent(func(enabled bool))
	// ConnectWallBounceChangedEvent provides the GUI with the function to call when the user uses the GUI to request
	// to enable/disable particles bouncing off environment walls.
	// The GUI is expected to change its state accordingly and then call this function, passing it a bool indicating
	// whether particle wall bounces should presently be enabled/disabled.
	ConnectWallBounceChangedEvent(func(enabled bool))
	// ConnectHistoryTrailChangedEvent provides the GUI with the function to call when the user uses the GUI to request
	// to enable/disable particle position history (trail).
	// The GUI is expected to change its state accordingly (and begin using the history state stored with the particles
	// to display the trails) and then call this function, passing it a bool indicating whether history should presently
	// be tracked or not.
	ConnectHistoryTrailChangedEvent(func(enabled bool))
	// ConnectHistoryTrailLengthChangedEvent provides the GUI with the function to call when the user uses the GUI to
	// request a change in the number of previous positions (trail length) of a particle the physics engine should track.
	// The GUI is expected to change its state accordingly and then call this function, passing it the new trail length.
	ConnectHistoryTrailLengthChangedEvent(func(value int))
	// ConnectPhysicsLoopSpeedChangedEvent provides the GUI with the function to call when the user uses the GUI to
	// request a change in the physics iteration speed.
	// The GUI is expected to change its state accordingly and then call this function, passing it the new speed
	// (iteration interval in ms).
	ConnectPhysicsLoopSpeedChangedEvent(func(value int))
	// ConnectResetEnvironmentEvent provides the GUI with the function to call when the user uses the GUI to request
	// that the environment be reset - that is, that the particles will be returned to their original (generated)
	// position and their historical positions removed.
	// The GUI is expected to call this method, which will in turn instruct the GUI to draw the particles.
	ConnectResetEnvironmentEvent(func())
	// ConnectPauseResumeEvent provides the GUI with the function to call when the user uses the GUI to request the
	// simulation pause or resume.
	// The GUI is expected to call this method, which will return a bool indicating whether the simulation is currently
	// paused or running. The GUI will then update its state accordingly (e.g. disabling controls while simulation is
	// running).
	ConnectPauseResumeEvent(func() (paused bool))
}
