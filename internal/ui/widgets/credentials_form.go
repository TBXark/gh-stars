package widgets

import (
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

// NewCredentialsForm creates a reusable form for GitHub credentials input.
// It provides three fields:
//   - Username: Required GitHub username
//   - Token: Optional GitHub personal access token (password entry)
//   - PerPage: Results per page (1-100)
//
// All fields are bound to provided data bindings for automatic synchronization
// with the underlying ViewModel.
func NewCredentialsForm(username, token, perPage binding.String) *widget.Form {
	usernameEntry := widget.NewEntryWithData(username)
	usernameEntry.SetPlaceHolder("octocat")

	tokenEntry := widget.NewPasswordEntry()
	tokenEntry.Bind(token)
	tokenEntry.SetPlaceHolder("optional")

	perPageEntry := widget.NewEntryWithData(perPage)
	perPageEntry.SetPlaceHolder("1-100")

	return widget.NewForm(
		widget.NewFormItem("Username", usernameEntry),
		widget.NewFormItem("Token", tokenEntry),
		widget.NewFormItem("per_page", perPageEntry),
	)
}
