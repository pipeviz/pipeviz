package mtf

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/pipeviz/pipeviz/version"
)

func init() {
	Register("ensure-client", tfEnsureClient)
}

// tfEnsureClient ensures a valid 'client' block is present in the message.
func tfEnsureClient(in []byte) (out []byte, changed bool, err error) {
	var v interface{}
	err = json.Unmarshal(in, &v)
	if err != nil {
		return nil, false, err
	}

	jmap := v.(map[string]interface{})

	if clienti, exists := jmap["client"]; !exists {
		jmap["client"] = map[string]interface{}{
			"name":      "autogen by EnsureClient",
			"version":   version.Version(),
			"microtime": time.Now().UnixNano() / 1000,
		}
		changed = true
	} else if client, ok := clienti.(map[string]interface{}); ok {
		if _, exists := client["name"]; !exists {
			client["name"] = "autogen by EnsureClient"
			changed = true
		}
		if _, exists := client["version"]; !exists {
			client["version"] = version.Version()
			changed = true
		}
		if _, exists := client["version"]; !exists {
			client["microtime"] = time.Now().UnixNano() / 1000
			changed = true
		}

		if changed {
			jmap["client"] = client
		}
	} else {
		return nil, false, errors.New("'client' key was not of expected type, not safe to proceed")
	}

	if changed {
		out, err = json.Marshal(&jmap)
		if err != nil {
			return nil, false, err
		}
	} else {
		out = in
	}

	return
}
