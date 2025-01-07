package handlers

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/pelletier/go-toml/v2"
	"log"
	"math/big"
	"os"
	"strings"
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

func ConvertToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int, int64, float64, bool:
		return fmt.Sprintf("%v", v)
	case map[string]interface{}:
		return convertMapToString(v)
	case []interface{}:
		return convertSliceToString(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func convertMapToString(m map[string]interface{}) string {
	result := "{"
	for key, value := range m {
		result += fmt.Sprintf("%s: %s, ", key, ConvertToString(value))
	}
	result = strings.TrimSuffix(result, ", ") + "}"
	return result
}

func convertSliceToString(s []interface{}) string {
	result := "["
	for _, value := range s {
		result += fmt.Sprintf("%s, ", ConvertToString(value))
	}
	result = strings.TrimSuffix(result, ", ") + "]"
	return result
}

// ReadTOMLToMap 读取 TOML 文件并将其内容解析为 map[string]interface{}
func ReadTOMLToMap(filePath string) (map[string]interface{}, error) {
	// 读取 TOML 文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read TOML file: %w", err)
	}

	// 将 TOML 文件内容解析为 map
	var result map[string]interface{}
	err = toml.Unmarshal(data, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal TOML: %w", err)
	}

	return result, nil
}

func EnsureDirExists(dirName string) error {
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		// 使用 MkdirAll 创建多层文件夹
		err := os.MkdirAll(dirName, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory '%s': %w", dirName, err)
		}
		log.Printf("Directory '%s' created successfully.\n", dirName)
	}
	return nil
}
