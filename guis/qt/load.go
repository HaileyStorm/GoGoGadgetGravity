// Package qt is a Qt-based implementation of guis.GUIEnabler
package qt

import (
	"image"
	"math"
	"os"
	"sync"

	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"

	"GoGoGadgetGravity/guis"
	eWidgets "GoGoGadgetGravity/guis/qt/enhanced_widgets"
	"GoGoGadgetGravity/physics"
)

// Qt is the struct containing GUI control handles and state data
type Qt struct {
	// View is the Qt graphics view object where particles are displayed. It is a container.
	View *widgets.QGraphicsView
	// Scene is the Qt graphics scene object where particles are displayed. Qt scene objects (e.g. shapes) can also be
	// added to the scene, but the way this is implemented, it contains only Pixmap.
	Scene *widgets.QGraphicsScene
	// Pixmap is the pixel-array image where the particles are drawn. The "pixels" that can be individually addressed
	// are determined by the EnvironmentSize. Each "pixel" may be drawn on the screen as multiple pixels, or less than
	// one pixel, depending on the size of the window and therefore the size of the View, Scene, and this object.
	// It is created from the Canvas.
	Pixmap *widgets.QGraphicsPixmapItem
	// statusbar is the status text control at the bottom of the window which is updated with the SetStatusText method.
	statusbar *widgets.QStatusBar

	// GridLayout is the main window layout.
	GridLayout *widgets.QGridLayout
	// FormLayout is the layout of the right hand side of the window which contains all the controls (other than the
	// scene objects).
	FormLayout *widgets.QFormLayout
	// FormItems is a map of eWidgets.EWidgeter instances, with descriptive names as keys. It is used to interact with
	// these widgets throughout the package (e.g. to connect an event handler or set a slider value).
	FormItems map[string]eWidgets.EWidgeter
	// SaveStateButton is the button which the user clicks to save the current simulation state to file
	SaveStateButton *widgets.QPushButton
	// LoadStateButton is the button which the user clicks to load the current simulation state from file
	LoadStateButton *widgets.QPushButton
	// ResetButton is the button which the user clicks to revert particles to their original (generated/loaded) state
	ResetButton *widgets.QPushButton
	// RegenButton is the button which the user clicks to generate a new set of particles
	RegenButton *widgets.QPushButton
	// PauseButton is the button which the user clicks to pause and resume the simulation
	PauseButton *widgets.QPushButton

	// Canvas is used to do pixel work on our Scene. It's bg is transparent. Like everything in the Scene, the
	// visibility of non-transparent pixels will depend on when the Canvas (as a whole) was updated vs when Items in the
	// Scene, if any, were updated.
	Canvas *gui.QImage
	// tempImage is used to go between Canvas & a temporary file (yes, file, because I'm dumb and can't sort out the
	// back-buffer), so we can do quick work w/ the canvas (Canvas.SetPixel, e.g., is horrifically slow)
	tempImage *image.NRGBA
	// imgLock is used to ensure thread-sfe access of tempImage
	imgLock sync.Mutex
	// im2qim indicates whether the im2qim mode (Canvas <-> file <-> standard library image) is currently active,
	// as set by StartIm2Qim / StopIm2Qim.
	im2qim bool

	//NoPen					*gui.QPen
	//TestEllipse			*widgets.QGraphicsEllipseItem

	// AllowMergeCheck is the checkbox the user (un)checks to indicate whether particle mergers should be enabled
	AllowMergeCheck *widgets.QCheckBox
	// WallBounceCheck is the checkbox the user (un)checks to indicate whether particles bounce off the "walls"
	// (environment bounds).
	WallBounceCheck *widgets.QCheckBox
	// HistoryTrailCheck is the checkbox the user (un)checks to indicate whether to track&display particle position
	// history trails.
	HistoryTrailCheck *widgets.QCheckBox

	// EnvironmentSize is kept in sync with state.Data.PhysicsEngine.EnvironmentSize and is used to (re)size the canvas,
	// determine whether pixels are in bounds when drawing particles, etc.
	EnvironmentSize int

	// loadingState indicates whether the simulation state is currently being loaded. Primarily used to disable
	// triggering connected main app event handlers during GUI control updates.
	loadingState bool

	// EventSystem holds the main app functions which have been connected to this GUI, which are triggered during GUI
	// interactions
	EventSystem EventSystemData
}

// CreateGUI implements guis.GUIEnabler.CreateGUI.
func (q *Qt) CreateGUI(initialValues guis.GUIInitializationData) {
	q.EnvironmentSize = initialValues.PhysicsEngine.EnvironmentSize

	widgets.NewQApplication(len(os.Args), os.Args)

	// Main Window
	var window = widgets.NewQMainWindow(nil, 0)
	window.SetWindowTitle("GoGo Gadget Gravity")
	// Minimum / initial size
	window.SetMinimumSize2(initialValues.WinMinWidth, initialValues.WinMinHeight)

	// Statusbar
	q.statusbar = widgets.NewQStatusBar(window)
	window.SetStatusBar(q.statusbar)

	// Canvas -> Pixmap -> Scene -> View
	q.Scene = widgets.NewQGraphicsScene(nil)
	q.View = widgets.NewQGraphicsView(nil)

	// When window is resized, View will be resized, and we need to scale View so that Scene fits
	q.View.ConnectResizeEvent(q.resizeEvent)

	// mainWidget contains the primary window layout, GridLayout
	mainWidget := widgets.NewQWidget(nil, 0)
	//window.SetCentralWidget(View)
	window.SetCentralWidget(mainWidget)

	//region Layouts
	q.loadingState = true
	// Since its parent is mainWidget and the Window's Central Widget is mainWidget, this is the main layout.
	// Specifying the parent here is all that's needed.
	q.GridLayout = widgets.NewQGridLayout(mainWidget)
	q.GridLayout.SetContentsMargins(11, 20, 11, 11)
	// Set up a grid, 2colX1row, with the first column by far the largest (to hold the View)
	q.GridLayout.SetColumnMinimumWidth(0, int(math.Round(float64(initialValues.WinMinWidth)*(2/3))))
	// Stretch factor is relative to other columns (a ratio, I think, but maybe a priority order or something, tbd).
	// A value of 0 means it won't stretch unless all columns have stretch factor 0 or are otherwise restricted from
	// expanding.
	q.GridLayout.SetColumnStretch(0, 4)
	q.GridLayout.SetColumnMinimumWidth(1, int(math.Round(float64(initialValues.WinMinWidth)*(1/3))))
	q.GridLayout.SetColumnStretch(1, 1)
	q.GridLayout.SetRowMinimumHeight(0, initialValues.WinMinHeight)
	q.GridLayout.SetRowStretch(0, 0)
	// Add the widgets to layout
	// Alignments:
	// 	Horizontal: 0=fill, 1=left, 2=right, 4=hcenter, 8=justify.
	// 	Vertical: 0=fill, 20=top, 40=bottom, 80=vcenter, 100="aligns with the baseline."
	//	Alignment values are OR'd ( | ). Special Qt__AlignCenter = hcenter | vcenter = 4|80
	// The View
	q.GridLayout.AddWidget2(q.View, 0, 0, core.Qt__AlignCenter)
	// A FormLayout for the controls (basically, a VBox w/ two column, label & widget)
	q.FormLayout = widgets.NewQFormLayout(nil)
	q.GridLayout.AddLayout(q.FormLayout, 0, 1, 0)

	q.FormItems = make(map[string]eWidgets.EWidgeter)
	q.SaveStateButton = widgets.NewQPushButton2("Save State To File", nil)
	q.SaveStateButton.ConnectClicked(q.SaveButtonClickEvent)
	q.FormLayout.AddWidget(q.SaveStateButton)
	q.LoadStateButton = widgets.NewQPushButton2("Load State From File", nil)
	q.LoadStateButton.ConnectClicked(q.LoadButtonClickEvent)
	q.FormLayout.AddWidget(q.LoadStateButton)
	q.FormLayout.AddItem(widgets.NewQSpacerItem(0, 20, 1|4|8, 1|4))
	q.FormItems["Environment Size (units*units)"] =
		eWidgets.NewESlider(400, 2500, 191, q.EnvironmentSize, 1)
	q.FormItems["Environment Size (units*units)"].(*eWidgets.ESlider).
		ConnectValueChangedEvent(q.EnvironmentSizeSliderChangedEvent)
	q.FormLayout.AddRow4("Environment Size (units*units)",
		q.FormItems["Environment Size (units*units)"].AsEWidget().ParentLayout)
	q.FormItems["Number of Particles"] =
		eWidgets.NewESlider(2, 1000, 91, initialValues.NumberOfParticles, 1)
	q.FormItems["Number of Particles"].(*eWidgets.ESlider).ConnectValueChangedEvent(q.NumParticlesSliderChangedEvent)
	q.FormLayout.AddRow4("Number of Particles", q.FormItems["Number of Particles"].AsEWidget().ParentLayout)
	q.FormItems["Average Mass"] =
		eWidgets.NewESlider(15, 1500, 135, initialValues.AverageMass, 1)
	q.FormItems["Average Mass"].(*eWidgets.ESlider).ConnectValueChangedEvent(q.AverageMassSliderChangedEvent)
	q.FormLayout.AddRow4("Average Mass", q.FormItems["Average Mass"].AsEWidget().ParentLayout)
	q.FormLayout.AddItem(widgets.NewQSpacerItem(0, 20, 1|4|8, 1|4))
	q.RegenButton = widgets.NewQPushButton2("Generate New Particles", nil)
	q.RegenButton.ConnectClicked(q.RegenButtonClickEvent)
	q.FormLayout.AddWidget(q.RegenButton)
	q.FormLayout.AddItem(widgets.NewQSpacerItem(0, 40, 1|4|8, 1|4))
	q.FormItems["Gravity Strength"] = eWidgets.NewESlider(0, 5000, 455,
		int(initialValues.PhysicsEngine.GravityStrength/0.1), 0.1)
	q.FormItems["Gravity Strength"].(*eWidgets.ESlider).ConnectValueChangedEvent(q.GravityStrengthSliderChangedEvent)
	q.FormLayout.AddRow4("Gravity Strength", q.FormItems["Gravity Strength"].AsEWidget().ParentLayout)
	q.FormItems["Close Charge Strength"] = eWidgets.NewESlider(0, 25000, 2273,
		int(initialValues.PhysicsEngine.CloseChargeStrength/10000), 10000)
	q.FormItems["Close Charge Strength"].(*eWidgets.ESlider).
		ConnectValueChangedEvent(q.CloseChargeStrengthSliderChangedEvent)
	q.FormLayout.AddRow4("Close Charge Strength", q.FormItems["Close Charge Strength"].AsEWidget().ParentLayout)
	q.FormItems["Far Charge Strength"] = eWidgets.NewESlider(0, 2000, 182,
		int(initialValues.PhysicsEngine.FarChargeStrength/0.01), 0.01)
	q.FormItems["Far Charge Strength"].(*eWidgets.ESlider).
		ConnectValueChangedEvent(q.FarChargeStrengthSliderChangedEvent)
	q.FormLayout.AddRow4("Far Charge Strength", q.FormItems["Far Charge Strength"].AsEWidget().ParentLayout)
	q.AllowMergeCheck = widgets.NewQCheckBox(nil)
	q.AllowMergeCheck.SetChecked(initialValues.PhysicsEngine.AllowMerge)
	q.AllowMergeCheck.ConnectClicked(q.AllowMergeClickEvent)
	q.FormLayout.AddRow3("Particles Can Merge", q.AllowMergeCheck)
	q.WallBounceCheck = widgets.NewQCheckBox(nil)
	q.WallBounceCheck.SetChecked(initialValues.PhysicsEngine.WallBounce)
	q.WallBounceCheck.ConnectClicked(q.WallBounceClickEvent)
	q.FormLayout.AddRow3("Wall Bounce", q.WallBounceCheck)
	q.HistoryTrailCheck = widgets.NewQCheckBox(nil)
	q.HistoryTrailCheck.ConnectClicked(q.HistoryTrailClickEvent)
	q.HistoryTrailCheck.SetChecked(true)
	q.FormLayout.AddRow3("Show History Trail", q.HistoryTrailCheck)
	q.FormItems["History Trail Length"] =
		eWidgets.NewESlider(3, 100, 5, initialValues.HistoryLength, 1)
	q.FormItems["History Trail Length"].(*eWidgets.ESlider).
		ConnectValueChangedEvent(q.HistoryTrailLengthSliderChangedEvent)
	q.FormLayout.AddRow4("History Trail Length", q.FormItems["History Trail Length"].AsEWidget().ParentLayout)
	q.FormItems["Physics Loop (ms)"] =
		eWidgets.NewESlider(75, 1500, 142, initialValues.PhysicsLoopSpeed, 1)
	q.FormItems["Physics Loop (ms)"].(*eWidgets.ESlider).ConnectValueChangedEvent(q.PhysicsLoopSliderChangedEvent)
	q.FormLayout.AddRow4("Physics Loop (ms)", q.FormItems["Physics Loop (ms)"].AsEWidget().ParentLayout)
	q.FormLayout.AddItem(widgets.NewQSpacerItem(0, 20, 1|4|8, 1|4))
	q.ResetButton = widgets.NewQPushButton2("Reset Particles", nil)
	q.ResetButton.ConnectClicked(q.ResetButtonClickEvent)
	q.FormLayout.AddWidget(q.ResetButton)
	q.FormLayout.AddItem(widgets.NewQSpacerItem(0, 20, 1|4|8, 1|4))
	q.PauseButton = widgets.NewQPushButton2("Start", nil)
	q.PauseButton.ConnectClicked(q.PauseButtonClickEvent)
	q.FormLayout.AddWidget(q.PauseButton)

	q.loadingState = false

	//endregion Layouts

	//region Canvas

	//Conveniently, this also sets up the bounds on the Scene (though we can overwrite that later with SetSceneRect()
	// if we want to zoom in/out)
	q.Canvas = gui.NewQImage().ConvertToFormat(gui.QImage__Format_ARGB32, core.Qt__AutoColor).
		Scaled2(q.EnvironmentSize, q.EnvironmentSize, core.Qt__KeepAspectRatio, core.Qt__FastTransformation)
	q.Pixmap = widgets.NewQGraphicsPixmapItem2(gui.NewQPixmap().FromImage(q.Canvas, 0), nil)

	q.DrawParticles(initialValues.PhysicsEngine.Particles)

	q.Scene.AddItem(q.Pixmap)
	//endregion Canvas

	//We can also use Items, e.g. Ellipses. But while more than fast enough for a relatively small number (50 or 100),
	// they're much slower than drawing pixels with the Canvas. On the other hand, there's a lot of convenient stuff
	// like getting all items in a bounding box, getting a list of all colliding items, etc...
	/*NoPen = gui.NewQPen2(core.Qt__NoPen)
	tempBrush := gui.NewQBrush3(gui.NewQColor3(0, 255, 0, 255), core.Qt__SolidPattern)
	TestEllipse = widgets.NewQGraphicsEllipseItem3(100, 100, 30, 20, nil)
	TestEllipse.SetPen(NoPen)
	TestEllipse.SetBrush(tempBrush)
	Scene.AddItem(TestEllipse)

	//This is called when a Scene Item moves outside the current Scene bounding Rect. I have no earthly clue what to do
	// with the Canvas in this is situation, which is what I intend to handle here. Do we scale it up and reset its
	// position 0,0? (if so I can't figure out how). Do we blow it away and recreate it at the new size - and if so,
	// do we get all the pixels & remap them or just wipe it or?
	//Or, maybe we should disallow this. Not sure the best way to do that, but probably handling each Item moving?
	Scene.ConnectSceneRectChanged(func(rect *core.QRectF) {
		//statusbar.ShowMessage("New Scene dims: " + strconv.FormatFloat(rect.Width(), 'f', 2, 64) + ", " +
			strconv.FormatFloat(rect.Height(), 'f', 2, 64), 0)
		//statusbar.ShowMessage("Scale: " + strconv.FormatFloat((rect.Width()*rect.Height())/
		//	(float64(Canvas.Width())*float64(Canvas.Height())), 'f', 2, 64), 0)
	})*/

	q.View.SetScene(q.Scene)

	//if EnvironmentSize < 800 {
	//	View.Scale(909 / float64(EnvironmentSize), 909 / float64(EnvironmentSize))
	//}
	q.View.Show()

	// Run App
	widgets.QApplication_SetStyle2("fusion")
	window.Show()
	widgets.QApplication_Exec()
}

// LoadState implements guis.GUIEnabler.LoadState
func (q *Qt) LoadState(initialValues guis.GUIInitializationData) {
	q.loadingState = true

	q.EnvironmentSize = initialValues.PhysicsEngine.EnvironmentSize
	q.FormItems["Environment Size (units*units)"].(*eWidgets.ESlider).
		SetValue(initialValues.PhysicsEngine.EnvironmentSize)
	q.FormItems["Number of Particles"].(*eWidgets.ESlider).SetValue(initialValues.NumberOfParticles)
	q.FormItems["Average Mass"].(*eWidgets.ESlider).SetValue(initialValues.AverageMass)
	q.FormItems["Gravity Strength"].(*eWidgets.ESlider).
		SetValueFromScaled(initialValues.PhysicsEngine.GravityStrength)
	q.FormItems["Close Charge Strength"].(*eWidgets.ESlider).
		SetValueFromScaled(initialValues.PhysicsEngine.CloseChargeStrength)
	q.FormItems["Far Charge Strength"].(*eWidgets.ESlider).
		SetValueFromScaled(initialValues.PhysicsEngine.FarChargeStrength)
	q.AllowMergeCheck.SetChecked(initialValues.PhysicsEngine.AllowMerge)
	q.WallBounceCheck.SetChecked(initialValues.PhysicsEngine.WallBounce)
	q.HistoryTrailCheck.SetChecked(initialValues.HistoryTrail)
	q.FormItems["History Trail Length"].(*eWidgets.ESlider).SetValue(initialValues.HistoryLength)
	q.FormItems["Physics Loop (ms)"].(*eWidgets.ESlider).SetValue(initialValues.PhysicsLoopSpeed)

	q.loadingState = false

	q.UpdateView(initialValues.PhysicsEngine.Particles)
}

// UpdateView implements guis.GUIEnabler.UpdateView
func (q *Qt) UpdateView(particles []*physics.Particle) {
	q.View.Hide()
	q.View.SetScene(nil)
	q.Scene.RemoveItem(q.Pixmap)
	q.Canvas = gui.NewQImage().ConvertToFormat(gui.QImage__Format_ARGB32, core.Qt__AutoColor).
		Scaled2(q.EnvironmentSize, q.EnvironmentSize, core.Qt__KeepAspectRatio, core.Qt__FastTransformation)
	q.Pixmap = widgets.NewQGraphicsPixmapItem2(gui.NewQPixmap().FromImage(q.Canvas, 0), nil)
	q.Scene.SetSceneRect2(0, 0, float64(q.EnvironmentSize), float64(q.EnvironmentSize))
	q.View.SetSceneRect2(0, 0, float64(q.EnvironmentSize), float64(q.EnvironmentSize))
	q.DrawParticles(particles)
	q.Scene.AddItem(q.Pixmap)
	q.View.SetScene(q.Scene)
	// Magic. Certain scales fit the View nicely, others leave big bezels, this makes it more likely to be the former
	q.View.Scale(909/float64(q.View.Width()), 909/float64(q.View.Height()))
	q.View.Show()
}

// SetPhysicsLoopSpeed implements guis.GUIEnabler.SetPhysicsLoopSpeed
func (q *Qt) SetPhysicsLoopSpeed(loopTime int) {
	// We know there's no need to scale / use SetValueFRomScaled
	q.FormItems["Physics Loop (ms)"].(*eWidgets.ESlider).SetValue(loopTime)
}

// SetStatusText implements guis.GUIEnabler.SetStatusText
func (q *Qt) SetStatusText(text string, timeout int) {
	q.statusbar.ShowMessage(text, timeout)
}
