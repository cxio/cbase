// Copyright 2022 of chainx.zh@gmail.com, All rights reserved.
// Use of this source code is governed by a MIT license.

package cbase

import (
	"crypto/ed25519"
)

const (
	// 表达式结束标志
	ExprEnd = -1

	// 加密公钥长度
	PubKeySize = ed25519.PublicKeySize
)
