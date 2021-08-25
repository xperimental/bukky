package testutil

// EqualErrorMessage returns true if the errors have the same error message.
func EqualErrorMessage(got, want error) bool {
	if got == want {
		return true
	}

	if got == nil && want != nil {
		return false
	}

	if got != nil && want == nil {
		return false
	}

	gotMsg := got.Error()
	wantMsg := want.Error()

	return gotMsg == wantMsg
}
