package app

import (
	"encoding/json"
	"net/url"
	"strings"

	"github.com/flashdatadev/flashdata-cli/internal/state"
)

func (a *App) rememberSecrets(value any) {
	a.knownSecrets = append(a.knownSecrets, inputSecrets(value)...)
}

func inputSecrets(value any) []string {
	secrets := []string{}
	b, err := json.Marshal(value)
	if err != nil {
		return secrets
	}
	var v any
	_ = json.Unmarshal(b, &v)
	var visit func(any)
	visit = func(v any) {
		switch v := v.(type) {
		case map[string]any:
			for key, x := range v {
				switch strings.ToLower(strings.ReplaceAll(key, "_", "")) {
				case "apikey", "accesskey", "secretkey", "sastoken", "privatekey", "setupcredential", "handoffproof", "registrationproof", "callbackurl":
					if s, ok := x.(string); ok && s != "" {
						secrets = append(secrets, s)
					}
				}
				visit(x)
			}
		case []any:
			for _, x := range v {
				visit(x)
			}
		}
	}
	visit(v)
	return secrets
}

func requirePrivateInput(path string, value any) error {
	if path == "" || path == "-" || len(inputSecrets(value)) == 0 {
		return nil
	}
	if _, err := state.ReadPrivate(path, 8<<20); err != nil {
		return fail("private_config_required", "Query files containing credentials or callbacks must be private, or read from stdin.", "")
	}
	return nil
}
func (a *App) redacted(value any) any {
	v := sanitize(value)
	var visit func(any) any
	visit = func(v any) any {
		switch v := v.(type) {
		case map[string]any:
			for k, x := range v {
				if content, ok := x.(string); k == "content" && ok {
					v[k] = a.redactedResponseText(content)
				} else {
					v[k] = visit(x)
				}
			}
		case []any:
			for i, x := range v {
				v[i] = visit(x)
			}
		case string:
			for _, secret := range a.knownSecrets {
				if secret != "" {
					v = strings.ReplaceAll(v, secret, "[REDACTED]")
				}
			}
			if u, err := url.Parse(v); err == nil && u.Host != "" {
				query := u.Query()
				for key := range query {
					switch strings.ToLower(key) {
					case "x-amz-signature", "x-goog-signature", "sig", "signature", "sas_token":
						return "[REDACTED signed URL]"
					}
				}
			}
			return v
		}
		return v
	}
	return visit(v)
}
