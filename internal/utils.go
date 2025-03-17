package internal

import (
	"bufio"
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
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

// ExecCommand 执行命令行，接受输入字符串作为标准输入，并实时输出结果
// command: 要执行的命令和参数，如 "python main.py"
// stdin: 要传递给命令的标准输入
// ctx: 上下文，可用于取消命令执行
func ExecCommand(command string, stdin string) {
	// 分割命令和参数
	cmdFields := strings.Fields(command)
	if len(cmdFields) == 0 {
		log.Error().Msg("命令不能为空")
		return
	}

	// 创建命令
	cmd := exec.CommandContext(context.Background(), cmdFields[0], cmdFields[1:]...)

	// 设置标准输入
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		log.Error().Msgf("自定义命令[ %s ]-无法获取标准输入管道: %s", command, err)
		return
	}

	// 设置标准输出和标准错误
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Error().Msgf("自定义命令[ %s ]-无法获取标准输出管道: %s", command, err)
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Error().Msgf("自定义命令[ %s ]-无法获取标准错误管道: %s", command, err)
		return
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		log.Error().Msgf("自定义命令[ %s ]-启动命令失败: %s", command, err)
		return
	}

	// 写入标准输入
	go func() {
		defer stdinPipe.Close()
		// 确保输入以换行符结尾
		if stdin != "" && !strings.HasSuffix(stdin, "\n") {
			stdin += "\n"
		}
		io.WriteString(stdinPipe, stdin)
	}()

	// 使用 WaitGroup 等待所有输出读取完毕
	var wg sync.WaitGroup
	wg.Add(2)

	// 读取并打印标准输出
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			log.Warn().Msgf("自定义命令[ %s ]-输出: %s", command, scanner.Text())
		}
	}()

	// 读取并打印标准错误
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			log.Error().Msgf("自定义命令[ %s ]-错误: %s", command, scanner.Text())
		}
	}()

	// 等待所有输出读取完毕
	wg.Wait()

	// 等待命令执行完成
	if err := cmd.Wait(); err != nil {
		log.Error().Msgf("自定义命令[ %s ]-执行失败: %s", command, err)
		return
	}
}
