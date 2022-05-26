//go:build windows
// +build windows

package slackclient

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/billgraziano/dpapi"
)

type EncryptedKey struct {
	EncryptedKey string `json:"encrypted_key"`
}
type LocalState struct {
	OSCrypt EncryptedKey `json:"os_crypt"`
}

func cookiePassword() ([]byte, error) {
	bs, err := os.ReadFile(path.Join(slackConfigDir(), "Local State"))
	if err != nil {
		return nil, err
	}

	var localState LocalState
	err = json.Unmarshal(bs, &localState)
	if err != nil {
		return nil, err
	}

	encryptedKey, err := base64.StdEncoding.DecodeString(localState.OSCrypt.EncryptedKey)
	if err != nil {
		return nil, err
	}

	encryptionMethod := encryptedKey[:5]
	if string(encryptionMethod) != "DPAPI" {
		return nil, fmt.Errorf("encryption method %q is not supported", encryptionMethod)
	}

	encryptedKey = encryptedKey[5:]
	decryptedKey, err := dpapi.DecryptBytes(encryptedKey)
	if err != nil {
		return nil, err
	}
	return decryptedKey, nil
}
