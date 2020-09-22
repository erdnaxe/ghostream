package bypass

// ByPass authentification backend
// By pass password check, open your streaming server to everyone!
type ByPass struct {
}

// Login always return success
func (a ByPass) Login(username string, password string) (bool, error) {
	return true, nil
}

// Close has no connection to close
func (a ByPass) Close() {
}

// New instanciates a new Basic authentification backend
func New() (ByPass, error) {
	backend := ByPass{}
	return backend, nil
}
