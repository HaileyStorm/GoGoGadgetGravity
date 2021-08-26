// Package eWidgets provides tools to create/interact with/manage Qt widgets composed of multiple child widgets
// arranged within a parent layout, surrounding a primary (main) widget.
package eWidgets

import (
	"fmt"
	"math"
	"strconv"

	"github.com/therecipe/qt/widgets"
)

// EWidget is a type (intended to be extended) for building Qt widgets composed of multiple child widgets arranged
// within a parent layout, surrounding a primary (main) widget.
type EWidget struct {
	// ParentLayout is the container widget used for placing the EWidget in the (external) layout.
	// It is parent to the MainWidget and any others such as associated QLabels.
	ParentLayout widgets.QLayout_ITF
	// MainWidget is the widget which the EWidget is built around, such as a QSlider
	MainWidget widgets.QWidget_ITF
}

// Enabled indicates whether the MainWidget is enabled (greyed out).
func (w *EWidget) Enabled() bool {
	return w.MainWidget.QWidget_PTR().IsEnabled()
}

// SetEnabled enables/disables (greys out) the MainWidget. It is meant to be overridden by EWidgeters to enable/disable
// all component widgets.
func (w *EWidget) SetEnabled(enable bool) {
	w.MainWidget.QWidget_PTR().SetEnabled(enable)
}

//region Interface

// EWidgeter is an interface for the EWidget struct, allowing any child struct (e.g. ESlider) to be represented,
// essentially, as an EWidget. It allows for collections of EWidgets.
type EWidgeter interface {
	// AsEWidget returns any EWidget / child instance as an EWidget.
	// The alternative to this provider pattern would be to have Getters&Setters
	// for each Widget (parent) field that might be used in generic collections
	// or elsewhere that it doesn't make sense to type assert to a child struct.
	// Note that we can (and do, above) still have getters/setters/other methods on the EWidget struct,
	// but they don't need to be part of this interface - the object in question would be used like
	// o.AsEWidget().OtherMethod()
	AsEWidget() *EWidget
}

// AsEWidget returns any EWidget (including child struct instances) as an EWidget.
func (w *EWidget) AsEWidget() *EWidget {
	return w
}

//endregion Interface

//region ESlider

// ESlider is an EWidget with a QSlider central widget and ticker & current value labels
type ESlider struct {
	EWidget
	// ValueLabel shows the currently selected value of the QSlider
	ValueLabel *widgets.QLabel
	// MinLabel shows the minimum value of the QSlider
	MinLabel *widgets.QLabel
	// MaxLabel shows the maximum value of the QSlider
	MaxLabel *widgets.QLabel

	// Scale is the scale factor applied to convert the slider value (which must be an integer) to the user/engine scale
	Scale float64

	// valueChangedEventHandlers is a slice of functions to be called when the slider value is changed. Appended to
	// using ConnectValueChangedEvent
	valueChangedEventHandlers []func(value int)
}

// Slider returns the EWidget.MainWidget as a *QSlider.
// Since this is implemented on ESlider (as opposed to EWidget), we assume there will not be an error,
// as this enables convenient one-liners.
func (w *ESlider) Slider() *widgets.QSlider { //(slider *widgets.QSlider, e error) {
	//slider, ok := w.MainWidget.(*widgets.QSlider)
	// Type-assert the main widget as a QSlider, which we know it is.
	slider, _ := w.MainWidget.(*widgets.QSlider)
	/*if ok {
		e = errors.New("MainWidget is not a *QSlider")
	} else {
		e = nil
	}
	return*/
	return slider
}

// SetEnabled enabled/disables (greys out) all child widgets of the ESlider (shadows EWidget.SetEnabled).
func (w *ESlider) SetEnabled(enable bool) {
	w.MainWidget.QWidget_PTR().SetEnabled(enable)

	// QLabels do not have a SetEnabled method, so we use their stylesheet to control their (foreground) color
	if enable {
		w.ValueLabel.SetStyleSheet("QLabel { color : black; }")
		w.MinLabel.SetStyleSheet("QLabel { color : black; }")
		w.MaxLabel.SetStyleSheet("QLabel { color : black; }")
	} else {
		w.ValueLabel.SetStyleSheet("QLabel { color : grey; }")
		w.MinLabel.SetStyleSheet("QLabel { color : grey; }")
		w.MaxLabel.SetStyleSheet("QLabel { color : grey; }")
	}
}

// GetValue is a convenience method to get the current value of the MainWidget slider
func (w *ESlider) GetValue() int {
	return w.Slider().Value()
}

// GetScaledValue is a convenience method to get the current value of the MainWidget slider, scaled by the Scale field
// (user units)
func (w *ESlider) GetScaledValue() float64 {
	return float64(w.Slider().Value()) * w.Scale
}

// SetValue is a convenience method to set the current value of the MainWidget slider
func (w *ESlider) SetValue(value int) {
	w.Slider().SetValue(value)
}

// SetValueFromScaled is a convenience method to set the current (displayed) value of the MainWidget slider from the
// supplied value, which is scaled by the Scale field (user units)
func (w *ESlider) SetValueFromScaled(value float64) {
	w.Slider().SetValue(int(math.Round(value / w.Scale)))
}

// ConnectValueChangedEvent connects a function so it will be triggered when triggerValueChangedEvent is called
// (that is, when the user changes the value of the slider).
func (w *ESlider) ConnectValueChangedEvent(f func(value int)) {
	w.valueChangedEventHandlers = append(w.valueChangedEventHandlers, f)
}

// triggerValueChangedEvent is the method connected to the MainWidget (Qt library) value changed event.
func (w *ESlider) triggerValueChangedEvent(value int) {
	// Value is integer
	if i, f := math.Modf(w.Scale); f == 0 {
		w.ValueLabel.SetText(strconv.Itoa(value * int(i)))
		// Value is float
	} else {
		w.ValueLabel.SetText(fmt.Sprintf("%.2f", float64(value)*w.Scale))
	}

	// Call all the subscribed event handlers
	for _, handler := range w.valueChangedEventHandlers {
		handler(value)
	}
}

//endregion ESlider
