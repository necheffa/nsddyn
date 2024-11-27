// Copyright (C) 2024 Alexander Necheff
// This program is licensed under the terms of the GPLv3.
// See the COPYING file that came packaged with this source code for the full terms.

package data

import (
	"encoding/base64"
)

type Key struct {
	Name   string
	Algo   string
	Secret string
}

func (k *Key) Base64() string {
	return base64.StdEncoding.EncodeToString([]byte(k.Secret))
}
