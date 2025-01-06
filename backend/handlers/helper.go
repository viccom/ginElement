package handlers

import (
	"github.com/google/uuid"
	"math/big"
)

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// 将大整数转换为 Base62 编码
func toBase62(n *big.Int) string {
	if n.Cmp(big.NewInt(0)) == 0 {
		return string(base62Chars[0])
	}
	var result []byte
	base := big.NewInt(62)
	zero := big.NewInt(0)
	for n.Cmp(zero) > 0 {
		remainder := new(big.Int)
		n.DivMod(n, base, remainder)
		result = append(result, base62Chars[remainder.Int64()])
	}
	// 反转结果
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return string(result)
}

// 生成 16 位 Base62 ID
func gen16ID() string {
	// 生成 UUID
	id := uuid.New()
	// 将 UUID 转换为大整数
	bigInt := new(big.Int)
	bigInt.SetBytes(id[:])
	// 转换为 Base62
	base62ID := toBase62(bigInt)
	// 取前 16 位
	if len(base62ID) > 16 {
		base62ID = base62ID[:16]
	}
	return base62ID
}
