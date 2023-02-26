// Copyright 2022 of chainx.zh@gmail.com, All rights reserved.
// Use of this source code is governed by a MIT license.

// Package cbase 基础公共集，一些全局的变量、常量和函数等。
package cbase

import (
	"bytes"
	"crypto/ed25519"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/cxio/cbase/paddr"
)

// 每年的区块数量。
// 按恒星年（Sidereal year），每6分钟一个块计算。
const SY6BLOCKS = 87661

// 原始铸币终止线。
// 即每块低于 3币 后终止。
const MINTENDLINE = 3e8

// 公钥类型引用。
type PubKey = ed25519.PublicKey

// 构造脚本ID。
// 脚本ID用于唯一性地标识一段脚本。
// 构成：
// - h 理想块高度（4字节）
// - n 交易ID在区块中的序位（4字节）
// - i 脚本序位（2字节）
// 返回：ID的字节序列。
func KeyID(h, n, i int) []byte {
	var buf bytes.Buffer

	binary.Write(&buf, binary.BigEndian, uint32(h))
	binary.Write(&buf, binary.BigEndian, uint32(n))
	binary.Write(&buf, binary.BigEndian, uint16(i))

	return buf.Bytes()
}

/*
 * 基本工具
 ******************************************************************************
 */

// 浮点数相等比较。
// d 为 x 和 y 之间的误差值，不超过则视为相等。
// 注：
// 如果d为零，就是严格相等了。
func FloatEqual(x, y, d float64) bool {
	return math.Abs(x-y) <= d
}

// 提取多重签名的公钥集。
// 即去除脚本中提供的公钥上前置的序位标识。
func MulPubKeys(pbks [][]byte) []PubKey {
	buf := make([]PubKey, len(pbks))

	for i, pk := range pbks {
		// 直接引用，
		// 因为单脚本验证为顺序执行，无并发被修改的风险。
		buf[i] = PubKey(pk[1:])
	}
	return buf
}

// 兑奖检查。
// 返回合法兑奖的数量（聪）。
func CheckAward(h int) int {
	//...
	return 0
}

// 单签名验证。
// ver 为版本值。便于安全升级。
// 当前采用ed25519签名认证。
func CheckSig(ver int, pubkey PubKey, msg, sig []byte) bool {
	// ver: 1
	return ed25519.Verify(pubkey, msg, sig)
}

// 多签名验证。
// ver 为版本值。便于安全升级。
// 当前采用ed25519签名认证。
func CheckSigs(ver int, pubkeys []PubKey, msg []byte, sigs [][]byte) bool {
	// ver: 1
	for i, pk := range pubkeys {
		if !ed25519.Verify(pk, msg, sigs[i]) {
			return false
		}
	}
	return true
}

// 系统内置验证（单签名）。
// 解锁数据：
// - ver 为版本值。
// - pubkey 用户的签名公钥。
// - msg 签名的消息：脚本ID（4+4+2）。
// - sig 用户的签名数据。
// - pkaddr 付款者的公钥地址。
// 注记：
// 需要对比目标公钥地址和计算出来的是否相同。
func SingleCheck(ver int, pubkey PubKey, msg, sig, pkaddr []byte) bool {
	pka := paddr.Hash([]byte(pubkey), nil)

	if !bytes.Equal(pka, pkaddr) {
		return false
	}
	return CheckSig(ver, pubkey, msg, sig)
}

// 系统内置验证（多重签名）。
// 公钥条目和公钥地址条目都已前置1字节的序位值（在公钥地址清单中的位置）。
// 解锁数据：
// - ver 为版本值。
// - msg 签名消息：脚本ID（4+4+2）。
// - sigs 签名数据集。
// - pks 签名公钥集（与签名集成员一一对应）。
// - pkhs 未签名公钥地址集。
// - pkaddr 多重签名公钥地址（付款者）。
// 注记：
// 需要先对比两个来源的公钥地址是否相同。
func MultiCheck(ver int, msg []byte, sigs, pks, pkhs [][]byte, pkaddr []byte) (bool, error) {
	pka, err := paddr.MulHash(pks, pkhs)

	if err != nil {
		return false, err
	}
	// 已含前置n/T配比对比。
	if !bytes.Equal(pka, pkaddr) {
		return false, nil
	}
	return CheckSigs(ver, MulPubKeys(pks), msg, sigs), nil
}

//
// 私有辅助
///////////////////////////////////////////////////////////////////////////////

// 奖励总量计算&打印。
// base 初始每块币量（单位：币）。
// rate 前阶比率（千分值），如 900 表示 90%。
// 返回：累计总量（单位：聪）。
func AwardTotal(base, rate int64) int64 {
	if rate >= 1000 {
		panic("比率设置错误")
	}
	var sum int64
	y := 0
	// 1币 = 1亿聪
	base *= 1e8

	fmt.Println("年次\t累计\t\t\t（年计）\t\t币量/块")
	fmt.Println("----------------------------------------------------------------------")

	for {
		// 低于 3币/块 时止
		if base < MINTENDLINE {
			break
		}
		ysum := base * SY6BLOCKS
		sum += ysum
		y++
		fmt.Printf("%d\t%d \t(%d)\t%d\n", y, sum, ysum, base)

		base = base * rate / 1000
	}
	return sum
}
