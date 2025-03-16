package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func SanitizeFilename(name string) string {
	// 定义Windows非法字符集合
	illegalChars := `\/:*?"<>|`
	cleaned := make([]rune, 0, len(name))
	for _, r := range name {
		if strings.ContainsRune(illegalChars, r) {
			cleaned = append(cleaned, '_')
		} else {
			cleaned = append(cleaned, r)
		}
	}

	// 处理路径分隔符（防止生成目录）
	cleanedStr := string(cleaned)
	cleanedStr = strings.Trim(cleanedStr, " .") // 去除首尾空格和点
	cleanedStr = strings.ReplaceAll(cleanedStr, string(os.PathSeparator), "_")

	// 如果名称为空则用默认名
	if cleanedStr == "" {
		return fmt.Sprintf("file_%d", time.Now().Unix())
	}

	// 检查文件名长度
	if len(cleanedStr) > 255 {
		cleanedStr = cleanedStr[:255]
	}

	return cleanedStr
}

// 格式化文件大小为易读格式
// func formatSize(bytes int64) string {
// 	const (
// 		KB = 1024
// 		MB = KB * 1024
// 		GB = MB * 1024
// 	)

// 	switch {
// 	case bytes >= GB:
// 		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
// 	case bytes >= MB:
// 		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
// 	case bytes >= KB:
// 		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
// 	default:
// 		return fmt.Sprintf("%d B", bytes)
// 	}
// }

func FillTemplatePath(template string, values map[string]string) string {
	// 首先替换所有模板变量
	re := regexp.MustCompile(`{{\s*([a-zA-Z_]+)\s*}}`)

	// 查找所有匹配并替换
	result := re.ReplaceAllStringFunc(template, func(match string) string {
		// 从匹配中提取键名
		submatch := re.FindStringSubmatch(match)
		if len(submatch) < 2 {
			return match // 如果没有捕获到键，保持原样
		}

		key := submatch[1]
		if value, exists := values[key]; exists {
			// 确保路径中的非法字符被处理（简单示例，实际使用可能需要更复杂的替换）
			sanitized := SanitizeFilename(value)
			return sanitized
		}

		return match // 如果映射中没有对应的键，保持原样
	})

	// 将所有正斜杠转换为操作系统特定的路径分隔符
	// 这样可以确保在 Windows 使用反斜杠，在 Unix/Linux/Mac 使用正斜杠
	result = filepath.FromSlash(result)

	return result
}

func FormatTime(t int) string {
	return time.Unix(int64(t), 0).Format("2006-01-02 15:04:05")
}

func FormatDate(t int) string {
	return time.Unix(int64(t), 0).Format("2006-01-02")
}

func DM2XmlD(d []*DanmakuStruct) []XmlD {
	var xd XmlD
	var xds []XmlD
	for _, v := range d {
		xd.P = fmt.Sprintf("%.3f,%d,%d,%d,%d,%d,%s,%s", float64(v.Progress)/1000, v.Mode, v.Fontsize, v.Color, v.Ctime, 0, v.MidHash, v.IdStr)
		xd.Text = v.Content
		xds = append(xds, xd)
	}
	return xds
}

func MergeDMList(odms, ldms []XmlD) []XmlD {
	var xds []XmlD
	// 根据 xd.P 去重后合并
	dmMap := make(map[string]XmlD)
	for _, dm := range odms {
		dmMap[dm.P] = dm
	}
	for _, dm := range ldms {
		dmMap[dm.P] = dm
	}
	for _, dm := range dmMap {
		xds = append(xds, dm)
	}
	return xds
}
