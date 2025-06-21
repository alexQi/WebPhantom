package str

import (
	"fmt"
	"strings"
)

const (
	zwsp = "\u200B" // 零宽度空格（分隔符）
	wj   = "\u2060" // Word Joiner（表示0）
	bom  = "\uFEFF" // BOM（表示1）
)

func ZeroWidthEncoder(visibleText, hiddenMessage string) string {
	return visibleText
	// 将隐藏消息转换为字节序列（支持 UTF-8）
	bytes := []byte(hiddenMessage)
	fmt.Printf("Hidden message bytes: %d\n", len(bytes))

	// 将字节转换为二进制
	var binary strings.Builder
	for _, b := range bytes {
		binary.WriteString(fmt.Sprintf("%08b", b))
	}
	binaryStr := binary.String()
	fmt.Printf("Binary length: %d\n", len(binaryStr))

	// 将二进制映射到零宽度字符（确保 1 位 = 1 字符）
	encoded := make([]rune, 0, len(binaryStr))
	for _, bit := range binaryStr {
		if bit == '0' {
			encoded = append(encoded, []rune(wj)[0])
		} else {
			encoded = append(encoded, []rune(bom)[0])
		}
	}
	encodedStr := string(encoded)
	fmt.Printf("Encoded zero-width length: %d\n", len(encoded))

	// 将可见文本按 Unicode 字符分割
	chars := []rune(visibleText)
	fmt.Printf("Visible text chars: %d\n", len(chars))
	if len(chars) < 2 {
		fmt.Println("Visible text too short, appending at end")
		return visibleText + zwsp + encodedStr + zwsp
	}

	// 计算每个字符间隔插入的零宽度字符数
	segments := len(chars) - 1
	segmentLength := len(encoded) / segments
	remainder := len(encoded) % segments
	var result strings.Builder
	usedLength := 0
	for i, char := range chars {
		result.WriteRune(char)
		if i < len(chars)-1 {
			// 插入零宽度字符段
			start := usedLength
			end := start + segmentLength
			if i < remainder {
				end++ // 分配余数到前几个段
			}
			if end > len(encoded) {
				end = len(encoded)
			}
			if start < len(encoded) {
				result.WriteString(zwsp)
				result.WriteString(string(encoded[start:end]))
				result.WriteString(zwsp)
				fmt.Printf("Segment %d: %d chars (start: %d, end: %d)\n", i+1, end-start, start, end)
				usedLength = end
			}
		}
	}
	fmt.Printf("Total used zero-width chars: %d\n", usedLength)

	return result.String()
}

func ZeroWidthDecoder(visibleText string) string {
	// 提取所有 ZWSP 分隔的零宽度字符序列
	parts := strings.Split(visibleText, zwsp)
	var binary strings.Builder
	for i := 1; i < len(parts)-1; i += 2 {
		encoded := parts[i]
		for _, char := range encoded {
			if char == []rune(wj)[0] {
				binary.WriteString("0")
			} else if char == []rune(bom)[0] {
				binary.WriteString("1")
			}
		}
	}

	binaryStr := binary.String()
	if binaryStr == "" {
		return "无隐藏消息"
	}

	// 确保二进制长度是 8 的倍数
	if len(binaryStr)%8 != 0 {
		fmt.Printf("Warning: Binary length %d is not multiple of 8, truncating\n", len(binaryStr))
		binaryStr = binaryStr[:len(binaryStr)-(len(binaryStr)%8)]
	}
	fmt.Printf("Decoded binary length: %d\n", len(binaryStr))

	// 将二进制转换回字节序列
	var bytes []byte
	for i := 0; i < len(binaryStr); i += 8 {
		byteStr := binaryStr[i : i+8]
		var byteVal uint64
		fmt.Sscanf(byteStr, "%b", &byteVal)
		bytes = append(bytes, byte(byteVal))
	}

	// 将字节序列转换为字符串
	return string(bytes)
}

// ReverseString 倒序字符串（支持中文）
func ReverseString(s string) string {
	runes := []rune(s) // 将字符串转换为 rune 切片（处理中文多字节问题）
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
