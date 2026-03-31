package data

import (
	"fmt"
	"os"
	"time"
)

var fileName = time.Now().Format("2006-01-02 15:04:05") + ".txt"

func LogA(text  string) {
	print(text + "\n")
	filePath := "logs/" + fileName
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
    if err != nil {
        fmt.Printf("無法開啟日誌檔案: %v\n", err)
        return
    }
    defer file.Close()

    // 寫入文字（建議加上換行）
    if _, err := file.WriteString(text + "\n"); err != nil {
        fmt.Printf("寫入日誌失敗: %v\n", err)
    }
	
}