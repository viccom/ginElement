package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/shirou/gopsutil/v3/cpu"
)

// GetHardwareID 生成一个基于CPU信息的唯一ID
func GetHardwareID() (string, error) {
	// 获取CPU信息
	info, err := cpu.Info()
	if err != nil {
		return "", fmt.Errorf("无法获取CPU信息: %v", err)
	}

	if len(info) == 0 {
		return "", fmt.Errorf("未找到CPU信息")
	}

	// 使用CPU的VendorID和ModelName生成哈希
	hash := sha256.New()
	hash.Write([]byte(info[0].VendorID + info[0].ModelName))
	hashSum := hash.Sum(nil)

	// 取哈希值的前8位作为ID
	id := hex.EncodeToString(hashSum)[:8]

	return id, nil
}
