package client

// Config is the dedicated fleetlock client config.
type Config struct {
	// Group of the instance. Defaults to "default"
	Group string
	// ID of the instance, must be unique and should persist across reboot.
	// Required.
	ID string
	// HTTP client to use - can be used to implement authentication logic.
	// Defaults to `http.DefaultClient`
	HTTP HTTPClient
	// URL of the FleetLock server implementation.
	// Required.
	URL string
}
