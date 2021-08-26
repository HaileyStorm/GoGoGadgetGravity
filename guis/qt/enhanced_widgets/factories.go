package eWidgets

import (
	"fmt"
	"math"
	"strconv"

	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
)

// NewESlider is a factory method for creating a new ESlider.
func NewESlider(min, max, interval, value int, scale float64) *ESlider {
	w := &ESlider{}

	pLayout := widgets.NewQGridLayout(nil)
	w.ParentLayout = pLayout

	// Prep the layout to place the widgets
	pLayout.SetColumnMinimumWidth(0, 50)
	pLayout.SetColumnStretch(0, 0)
	pLayout.SetColumnMinimumWidth(1, 50)
	pLayout.SetColumnStretch(1, 0)
	pLayout.SetColumnMinimumWidth(2, 20)
	pLayout.SetColumnStretch(2, 1)
	pLayout.SetRowMinimumHeight(0, 20)
	pLayout.SetRowStretch(0, 0)
	pLayout.SetRowMinimumHeight(1, 15)
	pLayout.SetRowStretch(1, 1)

	// Create the main slider widget
	tmpSlider := widgets.NewQSlider2(core.Qt__Horizontal, nil)
	tmpSlider.SetRange(min, max)
	tmpSlider.SetTickPosition(widgets.QSlider__TicksBelow)
	tmpSlider.SetTickInterval(interval)
	tmpSlider.SetValue(value)

	w.Scale = scale

	tmpSlider.ConnectValueChanged(w.triggerValueChangedEvent)
	// Add the slider to the layout and set it as the ESlider MainWidget
	pLayout.AddWidget3(tmpSlider, 0, 0, 1, 2, 0)
	w.MainWidget = tmpSlider

	// Create the value label, which appears to the right of the slider.
	tmpLabelValue := widgets.NewQLabel2("", nil, 0)
	if i, f := math.Modf(scale); f == 0 {
		tmpLabelValue.SetText(strconv.Itoa(value * int(i)))
	} else {
		tmpLabelValue.SetText(fmt.Sprintf("%.2f", float64(value)*w.Scale))
	}
	pLayout.AddWidget2(tmpLabelValue, 0, 2, core.Qt__AlignTop)
	w.ValueLabel = tmpLabelValue

	// Create the minimum value label, which appears beneath the slider on the left.
	tmpLabelMin := widgets.NewQLabel2("", nil, 0)
	if i, f := math.Modf(scale); f == 0 {
		tmpLabelMin.SetText(strconv.Itoa(min * int(i)))
	} else {
		tmpLabelMin.SetText(fmt.Sprintf("%.2f", float64(min)*w.Scale))
	}
	pLayout.AddWidget2(tmpLabelMin, 1, 0, core.Qt__AlignLeft|core.Qt__AlignTop)
	w.MinLabel = tmpLabelMin

	// Create the maximum value label, which appears beneath the slider on the right.
	tmpLabelMax := widgets.NewQLabel2("", nil, 0)
	if i, f := math.Modf(scale); f == 0 {
		tmpLabelMax.SetText(strconv.Itoa(max * int(i)))
	} else {
		tmpLabelMax.SetText(fmt.Sprintf("%.2f", float64(max)*w.Scale))
	}
	pLayout.AddWidget2(tmpLabelMax, 1, 1, core.Qt__AlignRight|core.Qt__AlignTop)
	w.MaxLabel = tmpLabelMax

	return w
}
