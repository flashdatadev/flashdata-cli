package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"unicode/utf8"
)

func decodeObject(b []byte, out any, fields ...string) error {
	var object map[string]json.RawMessage
	if err := decodeStrict(b, &object); err != nil {
		return err
	}
	if object == nil {
		return fail("invalid_json_object", "Provide a JSON object.", "")
	}
	allowed := map[string]bool{}
	for _, field := range fields {
		allowed[field] = true
	}
	for field := range object {
		if !allowed[field] {
			return fail("unsupported_field", "The object contains an unsupported field; field names are case sensitive.", "")
		}
	}
	return decodeStrict(b, out)
}

func uniqueJSON(b []byte) error {
	if !utf8.Valid(b) {
		return errors.New("JSON must be UTF-8")
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	var value func(int) error
	value = func(depth int) error {
		if depth > 64 {
			return errors.New("JSON nesting limit")
		}
		token, err := d.Token()
		if err != nil {
			return err
		}
		delimiter, ok := token.(json.Delim)
		if !ok {
			return nil
		}
		switch delimiter {
		case '{':
			seen := map[string]bool{}
			for d.More() {
				keyToken, err := d.Token()
				if err != nil {
					return err
				}
				key, ok := keyToken.(string)
				if !ok || seen[key] {
					return errors.New("duplicate JSON field")
				}
				seen[key] = true
				if err = value(depth + 1); err != nil {
					return err
				}
			}
			end, err := d.Token()
			if err != nil || end != json.Delim('}') {
				return errors.New("invalid object")
			}
		case '[':
			for d.More() {
				if err = value(depth + 1); err != nil {
					return err
				}
			}
			end, err := d.Token()
			if err != nil || end != json.Delim(']') {
				return errors.New("invalid array")
			}
		default:
			return errors.New("unexpected delimiter")
		}
		return nil
	}
	if err := value(0); err != nil {
		return err
	}
	if _, err := d.Token(); err != io.EOF {
		return errors.New("extra JSON content")
	}
	return nil
}
