package freebsd

// ParseJailBool interprets a boolean jail parameter as exported by
// jailparam_export(3), which renders booleans such as "persist" as "true" or
// "false". "1" is accepted too, for callers that pass raw integer values.
func ParseJailBool(s string) bool {
	return s == "true" || s == "1"
}
