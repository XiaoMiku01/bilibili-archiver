package internal

import (
	"fmt"
	"github.com/skip2/go-qrcode"
)

type QRCodeString string

func (q *QRCodeString) Print() {
	fmt.Print(*q)
}

type qrcodeTerminal struct {
	content string
}

func NewQR(content string) *qrcodeTerminal {
	return &qrcodeTerminal{content: content}
}

func (qt *qrcodeTerminal) Print() error {
	qr, err := qrcode.New(qt.content, qrcode.Medium)
	if err != nil {
		return err
	}
	data := qr.Bitmap()
	qrStr := qt.getQRCodeString(data)
	qrStr.Print()
	return nil
}

func (qt *qrcodeTerminal) getQRCodeString(data [][]bool) *QRCodeString {
	var output string
	length := len(data)
	for row := 2; row < length-3; row += 2 {
		topRow := data[row]
		botRow := data[row+1]
		for col := 2; col < length-3; col++ {
			topBlack := topRow[col]
			botBlack := botRow[col]
			switch {
			case topBlack && botBlack:
				output += "█"
			case topBlack && !botBlack:
				output += "▀"
			case !topBlack && botBlack:
				output += "▄"
			default:
				output += " "
			}
		}
		output += "\n"
	}
	qrStr := QRCodeString(output)
	return &qrStr
}
