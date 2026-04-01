package data

import (
	"fmt"
	"os"
	"time"
)

var fileNameBase = time.Now().Format("2006-01-02 15:04:05") + ".txt"
var isOutputMap = map[string]bool{
    "A": false,
    "B": true,
}

func baseLog(t string,text  string) {
    isOutput := isOutputMap[t]
    if isOutput {
        print(text + "\n")
    }
    filePath := "logs/" + t + fileNameBase
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
func LogA(text  string) {
    baseLog("A",text)
}

func LogB(text  string) {
    baseLog("B",text)
}