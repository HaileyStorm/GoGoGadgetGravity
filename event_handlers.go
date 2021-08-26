package main

import (
	"encoding/json"
	"os"
	"strconv"
	"time"

	"GoGoGadgetGravity/guis"
	"GoGoGadgetGravity/physics"
	"GoGoGadgetGravity/state"
)

// SaveStateEvent saves the current simulation state to file.
// It is triggered by the GUI after it provides a file picker to the user (the selected file path is passed to this
// function).
func SaveStateEvent(file string) {
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0755)
	if err == nil {
		defer f.Close()
		// Clear file and seek to start
		err = f.Truncate(0)
		if err == nil {
			_, err = f.Seek(0, 0)
		}
		// Create a json encoder that uses the file as its output
		enc := json.NewEncoder(f)
		enc.SetIndent("", "\t")
		// Encode (output to file)
		if err == nil {
			err = enc.Encode(State)
		}
		if err == nil {
			err = f.Sync()
		}
		if err == nil {
			GUI.SetStatusText("Current settings and "+strconv.Itoa(len(State.PhysicsEngine.Particles))+
				" particles saved to file: "+file, 0)
		} else {
			GUI.SetStatusText("Saving state failed. Error: "+err.Error(), 0)
		}
	} else {
		GUI.SetStatusText("Saving state to file failed. Error: "+err.Error(), 0)
	}
}

// LoadStateEvent loads the simulation state saved in a file.
// It is triggered by the GUI after it provides a file picker to the user (the selected file path is passed to this
// function).
func LoadStateEvent(file string) {
	f, err := os.OpenFile(file, os.O_RDONLY, 0755)
	if err == nil {
		defer f.Close()
		// Create a state.Data struct and decode the json data from the file into it
		var data *state.Data
		err = json.NewDecoder(f).Decode(&data)
		if err == nil {
			// The values of State are assigned the values we just read
			*State = *data
			// Since State.PhysicsEngine is a pointer, the values from the file aren't populated to the engine; set the
			// engine data to the values from file
			physics.Engine = *data.PhysicsEngine
			// Reset the State.PhysicsEngine to point to the physics.Engine (was it wiped by *Sate = *data?)
			// Todo: check if this is necessary
			State.PhysicsEngine = &physics.Engine

			// Calculate the proxies etc.
			physics.InitializeParticles()
			physics.SaveInitialParticleStates()

			// Tell the GUI to set control values and redraw the scene
			initialValues := guis.GUIInitializationData{
				// Important to use State instead of data because particle initialization has been done on State now
				Data: State,
				// Not used by LoadState
				WinMinWidth: 0,
				// Not used by LoadState
				WinMinHeight: 0,
			}
			GUI.LoadState(initialValues)

			// Individual particle TrackHistory and particle HistorySize (and the history slice) are not stored in file.
			// Restore these settings (and initialize history slice) using the global State settings as read from file.
			HistoryTrailChangedEvent(data.HistoryTrail)
			HistoryTrailLengthChangedEvent(data.HistoryLength)

			GUI.SetStatusText("Settings and "+strconv.Itoa(len(State.PhysicsEngine.Particles))+
				" particles loaded from file: "+file, 0)
		} else {
			GUI.SetStatusText("Loading state from file failed. Error: "+err.Error(), 0)
		}
	} else {
		GUI.SetStatusText("Loading state from file failed. Error: "+err.Error(), 0)
	}
}

// EnvironmentSizeChangedEvent updates the physics.Engine.EnvironmentSize and, if the simulation is currently paused,
// generates new particles randomly within that environment.
// It is triggered by the GUI.
func EnvironmentSizeChangedEvent(value int) {
	State.PhysicsEngine.EnvironmentSize = value
	if paused {
		GenerateParticles()
		GUI.UpdateView(State.PhysicsEngine.Particles)
	}
}

// NumParticlesChangedEvent updates the desired number of particles, and if the simulation is paused generates those
// particles.
// It is triggered by the GUI.
func NumParticlesChangedEvent(value int) {
	State.NumberOfParticles = value
	if paused {
		GenerateParticles()
		GUI.DrawParticles(State.PhysicsEngine.Particles)
	}
}

// AverageMassChangedEvent updates the desired average mass of generated particles, and if the simulation is paused
// generates those particles.
// It is triggered by the GUI.
func AverageMassChangedEvent(value int) {
	State.AverageMass = value
	if paused {
		GenerateParticles()
		GUI.DrawParticles(State.PhysicsEngine.Particles)
	}
}

// RegenParticlesEvent generates new random particles.
// It is triggered by GUI.
func RegenParticlesEvent() {
	GenerateParticles()
	GUI.DrawParticles(State.PhysicsEngine.Particles)
}

// GravityStrengthChangedEvent updates the physics.Engine.GravityStrength.
// It is triggered by the GUI.
func GravityStrengthChangedEvent(value float64) {
	State.PhysicsEngine.GravityStrength = value
}

// CloseChargeStrengthChangedEvent updates the physics.Engine.CloseChargeStrength.
// It is triggered by the GUI.
func CloseChargeStrengthChangedEvent(value float64) {
	State.PhysicsEngine.CloseChargeStrength = value
}

// FarChargeStrengthChangedEvent updates the physics.Engine.FarChargeStrength.
// It is triggered by the GUI.
func FarChargeStrengthChangedEvent(value float64) {
	State.PhysicsEngine.FarChargeStrength = value
}

// AllowMergeChangedEvent updates physics.Engine.AllowMerge.
// It is triggered by the GUI.
func AllowMergeChangedEvent(checked bool) {
	State.PhysicsEngine.AllowMerge = checked
}

// WallBounceChangedEvent updates physics.Engine.WallBounce.
// It is triggered by the GUI.
func WallBounceChangedEvent(checked bool) {
	State.PhysicsEngine.WallBounce = checked
}

// HistoryTrailChangedEvent updates State.HistoryTrail, and updates all physics.Engine.Particles accordingly.
// It is triggered by the GUI.
func HistoryTrailChangedEvent(checked bool) {
	State.HistoryTrail = checked
	for _, p := range State.PhysicsEngine.Particles {
		p.SetTrackHistory(checked)
		p.SetHistorySize(State.HistoryLength)
	}
}

// HistoryTrailLengthChangedEvent updates State.HistoryLength, and updates all physics.Engine.Particles accordingly.
// It is triggered by the GUI.
func HistoryTrailLengthChangedEvent(value int) {
	State.HistoryLength = value
	for _, p := range State.PhysicsEngine.Particles {
		// Truncate the position history slice if it's longer than the newly requested length
		if len(p.PositionHistory()) > value {
			p.SetPositionHistory(p.PositionHistory()[len(p.PositionHistory())-value:])
		}
		p.SetHistorySize(value)
	}
}

// PhysicsLoopSpeedChangedEvent updates the State.PhysicsLoopSpeed. If the simulation is running, it restarts the
// physics loop timer accordingly.
// It is triggered by the GUI.
func PhysicsLoopSpeedChangedEvent(value int) {
	State.PhysicsLoopSpeed = value
	if !paused {
		physicsTicker.Reset(time.Duration(value) * time.Millisecond)
	}
}

// ResetEnvironmentEvent restores the physics.Engine.Particles to the states stored when they were first
// generated/loaded.
// It is triggered by the GUI.
func ResetEnvironmentEvent() {
	physics.RestoreInitialParticleStates()

	// Clear existing particle history trails (while preserving the selected trail length)
	hold := State.HistoryLength
	HistoryTrailLengthChangedEvent(0)
	State.HistoryLength = hold
	HistoryTrailChangedEvent(State.HistoryTrail)

	GUI.DrawParticles(State.PhysicsEngine.Particles)
}

// PauseResumeEvent pauses and resumes the simulation (physics loop).
// It is triggered by the GUI.
func PauseResumeEvent() bool {
	//Now resuming
	if paused {
		paused = false
		physicsTicker = time.NewTicker(time.Duration(State.PhysicsLoopSpeed) * time.Millisecond)
		physicsDoneChan = make(chan bool)
		go physicsLoop()
		//Now pausing
	} else {
		paused = true
		physicsDoneChan <- true
		physicsTicker.Stop()
	}

	return paused
}
