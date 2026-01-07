package widgets

import (
	"image/color"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// StatusPanel creates a compact status bar with a colored dot and message.
// The dot reflects state: green for success, blinking yellow for loading, red for errors.
func NewStatusPanel(status, errorMsg binding.String, loading binding.Bool) fyne.CanvasObject {
	message := widget.NewLabel("")
	message.Wrapping = fyne.TextWrapWord

	dot := canvas.NewCircle(theme.SuccessColor())
	dot.StrokeWidth = 0
	dotHolder := container.NewGridWrap(fyne.NewSize(8, 8), dot)
	bar := container.NewBorder(nil, nil, container.NewCenter(dotHolder), nil, message)

	warningOpaque := dotColor(theme.WarningColor(), true)
	warningTransparent := dotColor(theme.WarningColor(), false)
	blinkAnim := canvas.NewColorRGBAAnimation(warningTransparent, warningOpaque, 650*time.Millisecond, func(c color.Color) {
		dot.FillColor = c
		dot.Refresh()
	})
	blinkAnim.AutoReverse = true
	blinkAnim.Curve = fyne.AnimationLinear
	blinkAnim.RepeatCount = fyne.AnimationRepeatForever
	blinking := false

	update := func() {
		statusText, _ := status.Get()
		errText, _ := errorMsg.Get()
		loadingVal, _ := loading.Get()

		statusText = strings.TrimSpace(statusText)
		errText = strings.TrimSpace(errText)
		if errText != "" {
			if statusText != "" {
				statusText = statusText + ": " + errText
			} else {
				statusText = errText
			}
		}
		message.SetText(statusText)

		if loadingVal {
			if !blinking {
				dot.FillColor = warningOpaque
				dot.Refresh()
				blinkAnim.Start()
				blinking = true
			}
			return
		}

		if blinking {
			blinkAnim.Stop()
			blinking = false
		}

		var base color.Color
		if errText != "" {
			base = theme.ErrorColor()
		} else {
			base = theme.SuccessColor()
		}
		dot.FillColor = dotColor(base, true)
		dot.Refresh()
	}

	updateFromBinding := func() {
		update()
	}
	status.AddListener(binding.NewDataListener(updateFromBinding))
	errorMsg.AddListener(binding.NewDataListener(updateFromBinding))
	loading.AddListener(binding.NewDataListener(updateFromBinding))
	update()

	return bar
}

func dotColor(base color.Color, visible bool) color.Color {
	c := color.NRGBAModel.Convert(base).(color.NRGBA)
	if !visible {
		c.A = 0
	}
	return c
}
