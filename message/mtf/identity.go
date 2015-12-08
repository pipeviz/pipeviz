package mtf

func init() {
	Register("identity", tfIdentity)
}

// tfIdentity passes the entire contents of the input along unchanged.
func tfIdentity(in []byte) ([]byte, bool, error) {
	return in, true, nil
}
