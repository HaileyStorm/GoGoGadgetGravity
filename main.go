// GoGoGadgetGravity is a particle simulator, including physics engine and gui packages, which uses artificial physics:
// Gravity is inversely proportional to distance^2.
// It is always positive and therefore attractive.
// Masses add. Radius is proxy.
// closeCharge is inversely proportional to distance^3.
// It may be negative or positive and therefore repulsive or attractive
// Charges average. Red (negative) and green (positive) are proxy (zero is black),
// with charge min/max +/- 1.
// farCharge is *proportional* to distance.
// It is always positive and therefore attractive.
// Charges average. Alpha is proxy with charge range  0-1.
package main

import (
	cryptorand "crypto/rand"
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"os"
	"reflect"
	"time"

	log "github.com/sirupsen/logrus"

	"GoGoGadgetGravity/guis"
	"GoGoGadgetGravity/guis/qt"
	"GoGoGadgetGravity/physics"
	"GoGoGadgetGravity/state"
)

var (
	// GUI is the guis.GUIEnabler instance. It is the front end to the application.
	GUI guis.GUIEnabler

	// State holds simulation state information.
	State *state.Data

	// physicsTicker is the ticker used for the execution of the physicsLoop (essentially, periodic calls to
	// physics.UpdateParticles).
	physicsTicker *time.Ticker
	// physicsDoneChan is the channel used to pause/stop the physicsLoop.
	physicsDoneChan chan bool
	// paused indicates whether the physicsLoop is currently running.
	paused bool
)

const (
	// Initial / minimum window size
	minW, minH = 1175, 855
	// See physics.EngineData and state.Data. These are starting values passed to the GUI for initialization.
	initialEnvironmentSize     = 800
	initialNumParticles        = 50
	initialAverageMass         = 250
	initialGravityStrength     = 15
	initialCloseChargeStrength = 150000000
	initialFarChargeStrength   = 7.5
	initialHistLength          = 15
	initialLoopSpeed           = 75
)

// main is ... well, you know...
func main() {
	paused = true

	State = &state.Data{
		NumberOfParticles: initialNumParticles,
		AverageMass:       initialAverageMass,
		HistoryTrail:      true,
		HistoryLength:     initialHistLength,
		PhysicsEngine:     &physics.Engine,
		PhysicsLoopSpeed:  initialLoopSpeed,
	}

	State.PhysicsEngine.Initialize()
	State.PhysicsEngine.GravityStrength = initialGravityStrength
	State.PhysicsEngine.CloseChargeStrength = initialCloseChargeStrength
	State.PhysicsEngine.FarChargeStrength = initialFarChargeStrength
	State.PhysicsEngine.EnvironmentSize = initialEnvironmentSize

	GUI = &qt.Qt{}
	// Set up to get notified of GUI events (user control interaction)
	GUI.ConnectSaveStateEvent(SaveStateEvent)
	GUI.ConnectLoadStateEvent(LoadStateEvent)
	GUI.ConnectEnvironmentSizeChangedEvent(EnvironmentSizeChangedEvent)
	GUI.ConnectNumParticlesChangedEvent(NumParticlesChangedEvent)
	GUI.ConnectAverageMassChangedEvent(AverageMassChangedEvent)
	GUI.ConnectRegenParticlesEvent(RegenParticlesEvent)
	GUI.ConnectGravityStrengthChangedEvent(GravityStrengthChangedEvent)
	GUI.ConnectCloseChargeStrengthChangedEvent(CloseChargeStrengthChangedEvent)
	GUI.ConnectFarChargeStrengthChangedEvent(FarChargeStrengthChangedEvent)
	GUI.ConnectAllowMergeChangedEvent(AllowMergeChangedEvent)
	GUI.ConnectWallBounceChangedEvent(WallBounceChangedEvent)
	GUI.ConnectHistoryTrailChangedEvent(HistoryTrailChangedEvent)
	GUI.ConnectHistoryTrailLengthChangedEvent(HistoryTrailLengthChangedEvent)
	GUI.ConnectPhysicsLoopSpeedChangedEvent(PhysicsLoopSpeedChangedEvent)
	GUI.ConnectResetEnvironmentEvent(ResetEnvironmentEvent)
	GUI.ConnectPauseResumeEvent(PauseResumeEvent)

	initRandom()
	GenerateParticles()

	// Create the GUI and set initial control values, and show the GUI & draw the particles
	initialValues := guis.GUIInitializationData{
		Data: &state.Data{
			PhysicsEngine: &physics.EngineData{
				GravityStrength:     initialGravityStrength,
				CloseChargeStrength: initialCloseChargeStrength,
				FarChargeStrength:   initialFarChargeStrength,
				EnvironmentSize:     initialEnvironmentSize,
				AllowMerge:          true,
				WallBounce:          true,
				Particles:           State.PhysicsEngine.Particles,
			},
			NumberOfParticles: initialNumParticles,
			AverageMass:       initialAverageMass,
			HistoryLength:     initialHistLength,
			PhysicsLoopSpeed:  initialLoopSpeed,
		},
		WinMinWidth:  minW,
		WinMinHeight: minH,
	}
	GUI.CreateGUI(initialValues)

	//Called after the window is closed
	//physicsDoneChan <- true
	//physicsTicker.Stop()
	os.Exit(0)
}

// physicsLoop loops forever / calls physics.UpdateParticles on the particles when the ticker ticks
// and stops/returns when the physicsDoneChan is written to (or when paused).
// The ticker is set up & started, or stopped, and this function is called as a goroutine, or physicsDoneChan is used
// to exit from it, from PauseResumeEvent
func physicsLoop() {
	var startPhysicsExecTime time.Time

	// Loop until done channel, executing the physics logic whenever the timer ticks
	for {
		// Waits for one of the conditions (since both cases are channels, it blocks rather than
		// immediately falling through): either a value to written to physicsDoneChan, or a tick of the ticker
		select {
		// Executes, and so exits this function, if true (or, really anything) is written to the channel
		case <-physicsDoneChan:
			return
		// Executes on physicsTicker tick
		case <-physicsTicker.C:
			if paused {
				return
			} // Shouldn't be necessary but also doesn't hurt
			startPhysicsExecTime = time.Now()

			// Where all the magic happens
			mergeOccurred, mergeMultiple, mergeSource, mergedResult := physics.UpdateParticles()

			// Set status with merger info
			if mergeOccurred {
				statusText := fmt.Sprintf("Merging %s with %s", mergeSource.ShortString(),
					reflect.ValueOf(mergeSource.MergingWith).MapKeys()[0].Interface().(*physics.Particle).ShortString())
				if mergeMultiple {
					statusText += " (et. al.)"
				}
				statusText += ". Now: " + mergedResult.ShortString()
				GUI.SetStatusText(statusText, 1500)
			}

			GUI.DrawParticles(State.PhysicsEngine.Particles)

			// Increase State.PhysicsLoopSpeed if actual execution time is longer than the requested time.
			loopTime := int(time.Since(startPhysicsExecTime).Milliseconds())
			fmt.Println(loopTime)
			if loopTime > State.PhysicsLoopSpeed {
				loopTime = int(float64(loopTime) * 1.05)
				GUI.SetPhysicsLoopSpeed(loopTime)
				PhysicsLoopSpeedChangedEvent(loopTime)
			}
		}
	}
}

// GenerateParticles generates random physics.Engine.Particles within the environment.
func GenerateParticles() {
	State.PhysicsEngine.Particles = make([]*physics.Particle, State.NumberOfParticles, State.NumberOfParticles)

	var m, cc, fc, x, y float64
	for i := 0; i < len(State.PhysicsEngine.Particles); i++ {
		// Random mass, normally distributed around State.AverageMass
		m = math.Min(math.Max(
			rand.NormFloat64()*0.55*float64(State.AverageMass)+float64(State.AverageMass),
			math.Max(4, 0.2*float64(State.AverageMass))), 1.75*float64(State.AverageMass))
		// For the charges, we just want a random number across the range, not a normal distribution
		cc = rand.Float64()*2.0 - 1.0
		fc = rand.Float64()
		// Random position.
		x = rand.Float64() * float64(State.PhysicsEngine.EnvironmentSize)
		y = rand.Float64() * float64(State.PhysicsEngine.EnvironmentSize)
		State.PhysicsEngine.Particles[i] = physics.NewParticle(m, cc, fc, x, y)
	}
	// Initialize history trails (enable/disable them in particles & create their empty position history "lists").
	HistoryTrailChangedEvent(State.HistoryTrail)
	HistoryTrailLengthChangedEvent(State.HistoryLength)

	physics.SaveInitialParticleStates()
}

// initRandom seeds math.rand with crypto/rand (imported as cryptorand), such that future math.rand operations are more or less cryptographically
// secure. It falls back to seeding with current nanosecond time. Without either, the math/rand package will always
// initialize with the same seed (0, I think).
// See: https://stackoverflow.com/a/54491783/5061881
// Imports:
// cryptorand "crypto/rand"
// log "github.com/sirupsen/logrus"
// TODO: Move this to CCSL
func initRandom() {
	// Gets 8 bytes using the cryptographically secure random package, and casts them into a uint64 and then an int64
	// (if you use a random byte for the most significant byte of a signed int64 you aren't randomly assigning the sign
	// bit, thus the conversion to unsigned first). I believe it shouldn't matter whether you use LittleEndian or
	// BigEndian, but you need to use one or the other to get to the Uint64([]byte) method.
	var b [8]byte
	_, err := cryptorand.Read(b[:])
	if err != nil {
		log.Warnln("Cannot seed math/rand package with cryptographically secure RNG, using time seed.")
		rand.Seed(time.Now().UTC().UnixNano())
		return
	}
	rand.Seed(int64(binary.LittleEndian.Uint64(b[:])))
}
