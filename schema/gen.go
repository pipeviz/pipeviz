package schema

//go:generate go-bindata -pkg $GOPACKAGE -o bindata_schema.go schema.json

// MasterSchema returns the master JSON schema as a byte array.
//
// This is a simple, static convenience function that wraps the go-bindata interface.
func Master() ([]byte, error) {
	return Asset("schema.json")
}
