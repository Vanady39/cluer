package auth

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

// EncodeGob serialises a value into an opaque string suitable for gin's
// string-valued context storage.
func EncodeGob(v interface{}) (string, error) {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	err := enc.Encode(v)
	if err != nil {
		return "", fmt.Errorf("failed to encode object: %w", err)
	}
	return buff.String(), nil
}

// DecodeGob reverses EncodeGob into target, which must be a pointer.
func DecodeGob(encoded string, target interface{}) error {
	var buff bytes.Buffer
	buff.WriteString(encoded)
	dec := gob.NewDecoder(&buff)
	if err := dec.Decode(target); err != nil {
		return fmt.Errorf("failed to decode object: %w", err)
	}
	return nil
}
