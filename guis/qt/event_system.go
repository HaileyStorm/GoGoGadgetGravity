package qt

import (
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"

	eWidgets "GoGoGadgetGravity/guis/qt/enhanced_widgets"
)

// EventSystemData holds the main app event handlers which are passed to the GUI using the Connect*Event methods,
// and which are called when GUI events are triggered to inform the main app of the changes/requests.
type EventSystemData struct {
	// See Qt.ConnectSaveStateEvent
	saveStateEventHandler func(value string)
	// See Qt.ConnectLoadStateEvent
	loadStateEventHandler func(value string)
	// See Qt.ConnectEnvironmentSizeChangedEvent
	environmentSizeChangedEventHandler func(value int)
	// See Qt.ConnectNumParticlesChangedEvent
	numParticlesChangedEventHandler func(value int)
	// See Qt.ConnectAverageMassChangedEvent
	averageMassChangedEventHandler func(value int)
	// See Qt.ConnectRegenParticlesEvent
	regenParticlesEventHandler func()
	// See Qt.ConnectGravityStrengthChangedEvent
	gravityStrengthChangedEventHandler func(value float64)
	// See Qt.ConnectCloseChargeStrengthChangedEvent
	closeChargeStrengthChangedEventHandler func(value float64)
	// See Qt.ConnectFarChargeStrengthChangedEvent
	farChargeStrengthChangedEventHandler func(value float64)
	// See Qt.ConnectAllowMergeChangedEvent
	allowMergeChangedEventHandler func(enabled bool)
	// See Qt.ConnectWallBounceChangedEvent
	wallBounceChangedEventHandler func(enabled bool)
	// See Qt.ConnectHistoryTrailChangedEvent
	historyTrailChangedEventHandler func(enabled bool)
	// See Qt.ConnectHistoryTrailLengthChangedEvent
	historyTrailLengthChangedEventHandler func(value int)
	// See Qt.ConnectPhysicsLoopSpeedChangedEvent
	physicsLoopSpeedChangedEventHandler func(value int)
	// See Qt.ConnectResetEnvironmentEvent
	resetEnvironmentEventHandler func()
	// See Qt.ConnectPauseResumeEvent
	pauseResumeEventHandler func() (paused bool)
}

// SaveButtonClickEvent is triggered when the user clicks the SaveStateButton. It presents a file picker and passes the
// selected file back to the main app using the provided event handler.
func (q *Qt) SaveButtonClickEvent(checked bool) {
	path, err := os.Getwd()
	// Path will be ""
	if err != nil {
		log.Warnln("Unable to get current directory: " + err.Error())
	}
	dlg := widgets.NewQFileDialog2(nil, "Select File", path, "*.json")
	dlg.SetAcceptMode(widgets.QFileDialog__AcceptSave)
	// Anonymous function called on selection of valid file / clicking Save
	dlg.ConnectFileSelected(func(file string) {
		if !strings.HasSuffix(file, ".json") {
			file += ".json"
		}
		// Tell the main app the selected file
		q.EventSystem.saveStateEventHandler(file)
	})
	// Show the dialog (waits for save / cancel)
	dlg.Show()
}

// ConnectSaveStateEvent implements guis.GUIEnabler.ConnectSaveStateEvent
func (q *Qt) ConnectSaveStateEvent(f func(file string)) {
	q.EventSystem.saveStateEventHandler = f
}

// LoadButtonClickEvent is triggered when the user clicks the LoadStateButton. It presents a file picker and passes the
// selected file back to the main app using the provided event handler.
func (q *Qt) LoadButtonClickEvent(checked bool) {
	path, err := os.Getwd()
	// Path will be ""
	if err != nil {
		log.Warnln("Unable to get current directory: " + err.Error())
	}
	dlg := widgets.NewQFileDialog2(nil, "Select File", path, "*.json")
	dlg.SetAcceptMode(widgets.QFileDialog__AcceptOpen)
	// Anonymous function called on selection of valid file / clicking Open
	dlg.ConnectFileSelected(func(file string) {
		//Tell the main app the selected file
		q.EventSystem.loadStateEventHandler(file)
	})
	// Show the dialog (waits for open / cancel)
	dlg.Show()
}

// ConnectLoadStateEvent implements guis.GUIEnabler.ConnectLoadStateEvent
func (q *Qt) ConnectLoadStateEvent(f func(file string)) {
	q.EventSystem.loadStateEventHandler = f
}

// EnvironmentSizeSliderChangedEvent is triggered when the user changes the value of the Environment Size slider and
// passes that value back to the main app using the provided event handler.
func (q *Qt) EnvironmentSizeSliderChangedEvent(value int) {
	q.EnvironmentSize = value
	if !q.loadingState {
		q.EventSystem.environmentSizeChangedEventHandler(value)
	} // We know this isn't scaled
}

// ConnectEnvironmentSizeChangedEvent implements guis.GUIEnabler.ConnectEnvironmentSizeChangedEvent
func (q *Qt) ConnectEnvironmentSizeChangedEvent(f func(value int)) {
	q.EventSystem.environmentSizeChangedEventHandler = f
}

// NumParticlesSliderChangedEvent is triggered when the user changes the value of the Number of Particles slider and
// passes that value back to the main app using the provided event handler.
func (q *Qt) NumParticlesSliderChangedEvent(value int) {
	if !q.loadingState {
		q.EventSystem.numParticlesChangedEventHandler(value)
	} // We know this isn't scaled
}

// ConnectNumParticlesChangedEvent implements guis.GUIEnabler.ConnectNumParticlesChangedEvent
func (q *Qt) ConnectNumParticlesChangedEvent(f func(value int)) {
	q.EventSystem.numParticlesChangedEventHandler = f
}

// AverageMassSliderChangedEvent is triggered when the user changes the value of the Average Mass slider and
// passes that value back to the main app using the provided event handler.
func (q *Qt) AverageMassSliderChangedEvent(value int) {
	if !q.loadingState {
		q.EventSystem.averageMassChangedEventHandler(value)
	} // We know this isn't scaled
}

// ConnectAverageMassChangedEvent implements guis.GUIEnabler.ConnectAverageMassChangedEvent
func (q *Qt) ConnectAverageMassChangedEvent(f func(value int)) {
	q.EventSystem.averageMassChangedEventHandler = f
}

// RegenButtonClickEvent is triggered when the user clicks the RegenButton. It informs the main app of this request by
// calling the provided event handler.
func (q *Qt) RegenButtonClickEvent(checked bool) {
	if !q.loadingState {
		q.EventSystem.regenParticlesEventHandler()
	}
}

// ConnectRegenParticlesEvent implements guis.GUIEnabler.ConnectRegenParticlesEvent
func (q *Qt) ConnectRegenParticlesEvent(f func()) {
	q.EventSystem.regenParticlesEventHandler = f
}

// GravityStrengthSliderChangedEvent is triggered when the user changes the value of the Gravity Strength slider and
// passes that value (scaled from slider to engine units) back to the main app using the provided event handler.
func (q *Qt) GravityStrengthSliderChangedEvent(value int) {
	if !q.loadingState {
		q.EventSystem.gravityStrengthChangedEventHandler(float64(value) *
			q.FormItems["Gravity Strength"].(*eWidgets.ESlider).Scale)
	}
}

// ConnectGravityStrengthChangedEvent implements guis.GUIEnabler.ConnectGravityStrengthChangedEvent
func (q *Qt) ConnectGravityStrengthChangedEvent(f func(value float64)) {
	q.EventSystem.gravityStrengthChangedEventHandler = f
}

// CloseChargeStrengthSliderChangedEvent is triggered when the user changes the value of the Close Charge Strength
// slider and passes that value (scaled from slider to engine units) back to the main app using the provided event handler.
func (q *Qt) CloseChargeStrengthSliderChangedEvent(value int) {
	if !q.loadingState {
		q.EventSystem.closeChargeStrengthChangedEventHandler(float64(value) *
			q.FormItems["Close Charge Strength"].(*eWidgets.ESlider).Scale)
	}
}

// ConnectCloseChargeStrengthChangedEvent implements guis.GUIEnabler.ConnectCloseChargeStrengthChangedEvent
func (q *Qt) ConnectCloseChargeStrengthChangedEvent(f func(value float64)) {
	q.EventSystem.closeChargeStrengthChangedEventHandler = f
}

// FarChargeStrengthSliderChangedEvent is triggered when the user changes the value of the Far Charge Strength slider
// and passes that value (scaled from slider to engine units) back to the main app using the provided event handler.
func (q *Qt) FarChargeStrengthSliderChangedEvent(value int) {
	if !q.loadingState {
		q.EventSystem.farChargeStrengthChangedEventHandler(float64(value) *
			q.FormItems["Far Charge Strength"].(*eWidgets.ESlider).Scale)
	}
}

// ConnectFarChargeStrengthChangedEvent implements guis.GUIEnabler.ConnectFarChargeStrengthChangedEvent
func (q *Qt) ConnectFarChargeStrengthChangedEvent(f func(value float64)) {
	q.EventSystem.farChargeStrengthChangedEventHandler = f
}

// AllowMergeClickEvent is triggered when the user clicks the AllowMergeCheck. It passes the current checked state back
// to the main app using the provided handler.
func (q *Qt) AllowMergeClickEvent(checked bool) {
	if !q.loadingState {
		q.EventSystem.allowMergeChangedEventHandler(checked)
	}
}

// ConnectAllowMergeChangedEvent  implements guis.GUIEnabler.ConnectAllowMergeChangedEvent
func (q *Qt) ConnectAllowMergeChangedEvent(f func(enabled bool)) {
	q.EventSystem.allowMergeChangedEventHandler = f
}

// WallBounceClickEvent is triggered when the user clicks the WallBounceCheck. It passes the current checked state back
// to the main app using the provided handler.
func (q *Qt) WallBounceClickEvent(checked bool) {
	if !q.loadingState {
		q.EventSystem.wallBounceChangedEventHandler(checked)
	}
}

// ConnectWallBounceChangedEvent  implements guis.GUIEnabler.ConnectWallBounceChangedEvent
func (q *Qt) ConnectWallBounceChangedEvent(f func(enabled bool)) {
	q.EventSystem.wallBounceChangedEventHandler = f
}

// HistoryTrailClickEvent is triggered when the user clicks the HistoryTrailCheck. It passes the current checked state
// back to the main app using the provided handler.
func (q *Qt) HistoryTrailClickEvent(checked bool) {
	if !q.loadingState {
		q.EventSystem.historyTrailChangedEventHandler(checked)
	}
}

// ConnectHistoryTrailChangedEvent implements guis.GUIEnabler.ConnectHistoryTrailChangedEvent
func (q *Qt) ConnectHistoryTrailChangedEvent(f func(enabled bool)) {
	q.EventSystem.historyTrailChangedEventHandler = f
}

// HistoryTrailLengthSliderChangedEvent is triggered when the user changes the value of the History Trail Length slider
// and passes that value back to the main app using the provided event handler.
func (q *Qt) HistoryTrailLengthSliderChangedEvent(value int) {
	if !q.loadingState {
		q.EventSystem.historyTrailLengthChangedEventHandler(value)
	} // We know this isn't scaled
}

// ConnectHistoryTrailLengthChangedEvent implements guis.GUIEnabler.ConnectHistoryTrailLengthChangedEvent
func (q *Qt) ConnectHistoryTrailLengthChangedEvent(f func(value int)) {
	q.EventSystem.historyTrailLengthChangedEventHandler = f
}

// PhysicsLoopSliderChangedEvent is triggered when the user changes the value of the Physics Loop Speed slider
// and passes that value back to the main app using the provided event handler.
func (q *Qt) PhysicsLoopSliderChangedEvent(value int) {
	if !q.loadingState {
		q.EventSystem.physicsLoopSpeedChangedEventHandler(value)
	} // We know this isn't scaled
}

// ConnectPhysicsLoopSpeedChangedEvent implements guis.GUIEnabler.ConnectPhysicsLoopSpeedChangedEvent
func (q *Qt) ConnectPhysicsLoopSpeedChangedEvent(f func(value int)) {
	q.EventSystem.physicsLoopSpeedChangedEventHandler = f
}

// ResetButtonClickEvent is triggered when the user clicks the ResetButton. It informs the main app of this request by
// calling the provided event handler.
func (q *Qt) ResetButtonClickEvent(checked bool) {
	q.EventSystem.resetEnvironmentEventHandler()
}

// ConnectResetEnvironmentEvent implements guis.GUIEnabler.ConnectResetEnvironmentEvent
func (q *Qt) ConnectResetEnvironmentEvent(f func()) {
	q.EventSystem.resetEnvironmentEventHandler = f
}

// PauseButtonClickEvent is triggered when the user clicks the PauseButton. It informs the main app of this request by
// calling the provided event handler, which returns whether the simulation is currently paused, which is used to
// enable/disable GUI elements and update the PauseButton text.
func (q *Qt) PauseButtonClickEvent(checked bool) {
	paused := q.EventSystem.pauseResumeEventHandler()

	// Now pausing
	if paused {
		q.PauseButton.SetText("Resume")

		q.SaveStateButton.SetEnabled(true)
		q.LoadStateButton.SetEnabled(true)
		q.FormItems["Environment Size (units*units)"].(*eWidgets.ESlider).SetEnabled(true)
		q.FormItems["Number of Particles"].(*eWidgets.ESlider).SetEnabled(true)
		q.FormItems["Average Mass"].(*eWidgets.ESlider).SetEnabled(true)
		q.RegenButton.SetEnabled(true)
		q.ResetButton.SetEnabled(true)
		// Now resuming
	} else {
		q.PauseButton.SetText("Pause")

		q.SaveStateButton.SetEnabled(false)
		q.LoadStateButton.SetEnabled(false)
		q.FormItems["Environment Size (units*units)"].(*eWidgets.ESlider).SetEnabled(false)
		q.FormItems["Number of Particles"].(*eWidgets.ESlider).SetEnabled(false)
		q.FormItems["Average Mass"].(*eWidgets.ESlider).SetEnabled(false)
		q.RegenButton.SetEnabled(false)
		q.ResetButton.SetEnabled(false)
	}
}

// ConnectPauseResumeEvent implements guis.GUIEnabler.ConnectPauseResumeEvent
func (q *Qt) ConnectPauseResumeEvent(f func() (paused bool)) {
	q.EventSystem.pauseResumeEventHandler = f
}

// resizeEvent is triggered when the window (and therefore View) is resized. It scales View such that Scene will
// fit in it.
func (q *Qt) resizeEvent(e *gui.QResizeEvent) {
	//This doesn't control what's included in the scene or whether scene items are cut off (they're not) - it makes the
	// Scene fit in the View (scales it) so that the View doesn't have scrollbars to move around the Scene.
	q.View.FitInView(q.Scene.ItemsBoundingRect(), core.Qt__KeepAspectRatio)
}
