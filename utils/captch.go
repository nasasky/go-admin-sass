package utils

import (
	"bytes"
	"fmt"
	"math/rand"
	"time"

	svg "github.com/ajstarks/svgo"
)

// 生成指定长度的随机数字字符串
func generateRandomDigits(length int) string {
	rand.Seed(time.Now().UnixNano())
	digits := make([]byte, length)
	for i := 0; i < length; i++ {
		digits[i] = byte('0' + rand.Intn(10))
	}
	return string(digits)
}

// 生成随机颜色
func randomColor() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("#%02x%02x%02x",
		rand.Intn(156)+100, // 避免太深的颜色
		rand.Intn(156)+100,
		rand.Intn(156)+100)
}

// GenerateSVG 生成SVG验证码
// width和height是图片尺寸
func GenerateSVG(width, height int) ([]byte, string) {
	// 生成4位随机数字
	code := generateRandomDigits(4)

	var svgContent bytes.Buffer
	canvas := svg.New(&svgContent)
	canvas.Start(width, height)

	// 绘制背景
	canvas.Rect(0, 0, width, height, "fill:white")

	// 添加干扰线
	for i := 0; i < 6; i++ {
		x1 := rand.Intn(width)
		y1 := rand.Intn(height)
		x2 := rand.Intn(width)
		y2 := rand.Intn(height)
		canvas.Line(x1, y1, x2, y2,
			fmt.Sprintf("stroke:%s;stroke-width:1", randomColor()))
	}

	// 添加干扰点
	for i := 0; i < 30; i++ {
		x := rand.Intn(width)
		y := rand.Intn(height)
		canvas.Circle(x, y, 1, fmt.Sprintf("fill:%s", randomColor()))
	}

	// 计算每个字符的位置
	charWidth := width / 5
	for i, char := range code {
		// 随机调整字符位置，增加难度
		x := charWidth * (i + 1)
		y := height/2 + rand.Intn(10) - 5
		// 随机旋转角度
		rotate := rand.Intn(30) - 15
		// 使用transform实现旋转效果
		canvas.Text(x, y, string(char),
			fmt.Sprintf("text-anchor:middle;font-size:%dpx;fill:%s;transform:rotate(%d,%d,%d)",
				height/2, randomColor(), rotate, x, y))
	}

	canvas.End()
	return svgContent.Bytes(), code
}
