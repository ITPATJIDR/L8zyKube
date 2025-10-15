package main

import (
	"l8zykube/components"
)

// Example demonstrating how to use the reusable Modal component
func main() {
	modal := components.NewModal()

	// Example 1: Simple error modal
	modal.ShowError("Connection Failed", "Could not connect to the server.", "Q")

	// Example 2: Warning modal
	modal.ShowWarning("Disk Space Low", "You have less than 1GB of free space remaining.")

	// Example 3: Success modal
	modal.ShowSuccess("Operation Complete", "Your data has been saved successfully.")

	// Example 4: Info modal
	modal.ShowInfo("New Feature", "A new feature has been added to the application.")

	// Example 5: Confirmation modal with callbacks
	modal.ShowConfirm("Delete Item", "Are you sure you want to delete this item?",
		func() {
			// This function is called when user selects "Yes"
			println("Item deleted!")
		},
		func() {
			// This function is called when user selects "No"
			println("Deletion cancelled.")
		},
	)

	// Example 6: Custom modal with multiple buttons
	modal.ShowWithButtons("Choose Action", "What would you like to do?",
		components.ModalInfo,
		[]string{"Save", "Save As", "Cancel"})

	// Example 7: Modal with custom callbacks
	modal.OnConfirm = func() {
		println("Confirmed!")
	}
	modal.OnCancel = func() {
		println("Cancelled!")
	}

	// Navigation examples:
	// - Use left/right arrow keys or h/l to navigate between buttons
	// - Use Enter to select a button
	// - Use q or Ctrl+C to close the modal

	// Note: This is just an example file showing usage patterns.
	// In a real application, you would integrate this with your TUI framework.
}
