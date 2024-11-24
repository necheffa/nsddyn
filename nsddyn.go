// Copyright (C) 2024 Alexander Necheff
// This program is licensed under the terms of the GPLv3.
// See the COPYING file that came packaged with this source code for the full terms.

package nsddyn

import (
	_ "embed"
)

//go:embed VERSION
var version string

func Version() string {
	return version
}
