package module

import (
	"net"
	"strconv"
	"strings"
	"fmt"

	"web_crawler/errors"
)

// 组件ID模板
var midTemplate = "%s%d|%s"

// 组件ID   类型字母、序列号和组件网络地址
type MID string

// 计算组件评分函数
type CalculateScore func(counts Counts) uint64

// LegalMID 用于判断给定的组件ID是否合法。
func LegalMID(mid MID) bool {
	if _, err := SplitMID(mid); err == nil {
		return true
	}
	return false
}

// SplitMID 用于分解组件ID。
// 第一个结果值表示分解是否成功。
// 若分解成功，则第二个结果值长度为3，
// 并依次包含组件类型字母、序列号和组件网络地址（如果有的话）。
func SplitMID(mid MID) ([]string, error) {
	var ok bool
	var letter string
	var snStr string
	var addr string
	midStr := string(mid)
	if len(midStr) < 1 {
		return nil, errors.NewIllegalParameterError("insufficient MID")
	}

	letter = midStr[:1]
	if _, ok = legalLetterTypeMap[letter]; !ok {
		return nil, errors.NewIllegalParameterError(
			fmt.Sprintf("illegal module type letter: %s", letter))
	}

	snAndAddr := midStr[1:]
	index := strings.LastIndex(snAndAddr, "|")
	if index < 0 {
		snStr := snAndAddr
		if !legalSN(snStr) {
			return nil, errors.NewIllegalParameterError(
				fmt.Sprintf("illegal module SN: %s", snStr))
		}
	} else {
		snStr = snAndAddr[:index]
		if !legalSN(snStr) {
			return nil, errors.NewIllegalParameterError(
				fmt.Sprintf("illegal module SN: %s", snStr))
		}
		addr = snAndAddr[index+1:]
		index = strings.LastIndex(addr, ":")
		if index <= 0 {
			return nil, errors.NewIllegalParameterError(
				fmt.Sprintf("illegal module address: %s", addr))
		}
		ipStr := addr[:index]
		if ip := net.ParseIP(ipStr); ip == nil {
			return nil, errors.NewIllegalParameterError(
				fmt.Sprintf("illegal module IP: %s", ip))
		}
		portStr := addr[index+1:]
		if _, err := strconv.ParseUint(portStr, 10, 64); err != nil {
			return nil, errors.NewIllegalParameterError(
				fmt.Sprintf("illegal module port: %s", portStr))
		}

		return []string{letter, snStr, addr}, nil
	}
}

// legalSN 用于判断序列号的合法性
func legalSN(snStr string) bool {
	_, err := strconv.ParseUint(snStr, 10, 64)
	if err != nil {
		return false
	}
	return true
}