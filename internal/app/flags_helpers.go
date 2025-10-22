// C:\_Projects_Go\AcousticLog\internal\app\flags_helpers.go

package app

import "os"

var _osArgs = func() []string { return os.Args }
var _setOsArgs = func(v []string) { os.Args = v }

func hasToken(tok string) bool {
	for _, a := range _osArgs() {
		if a == tok {
			return true
		}
	}
	return false
}

func stripToken(tok string) {
	in := _osArgs()
	out := make([]string, 0, len(in))
	for _, a := range in {
		if a == tok {
			continue
		}
		out = append(out, a)
	}
	_setOsArgs(out)
}
