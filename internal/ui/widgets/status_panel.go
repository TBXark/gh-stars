package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

// StatusPanel creates a reusable status display panel.
// It displays two lines:
//   - Status: Current operation status (e.g., "Loading...", "Loaded 50 repos")
//   - Error: Error message if operation failed (empty on success)
//
// Both labels are bound to provided data bindings for automatic updates
// when the underlying ViewModel changes state.
//
// Usage:
//
//	panel := widgets.NewStatusPanel(vm.Status, vm.Error)
func NewStatusPanel(status, errorMsg binding.String) fyne.CanvasObject {
	statusLabel := widget.NewLabelWithData(status)
	errorLabel := widget.NewLabelWithData(errorMsg)
	return container.NewVBox(statusLabel, errorLabel)
}
