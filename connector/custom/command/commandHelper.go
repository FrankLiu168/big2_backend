package command

import (
	"encoding/json"
	"errors"
)

// 假设你的命令结构体定义如下
type SayAllCommand struct {
	CommandName string `json:"command_name"`
	Message     string `json:"message"`
}

type SayToCommand struct {
	CommandName string `json:"command_name"`
	TargetID    string `json:"target_id"`
	Message     string `json:"message"`
}

// GetCommandName 保持不变
func GetCommandName(message []byte) string {
	var cmd struct {
		CommandName string `json:"command_name"`
	}
	if err := json.Unmarshal(message, &cmd); err != nil {
		return ""
	}
	return cmd.CommandName
}

var commandMap = map[string]any{
	"say_all": &SayAllCommand{},
	"say_to":  &SayToCommand{},
}

// GetCommand 重构版
// 思路：根据 name 创建对应的具体类型指针，然后 Unmarshal 进去，最后断言转为 T
func GetCommand[T any](name string, message []byte) (*T, error) {
	var target any

	target, ok := commandMap[name]
	if !ok {
		return nil, ErrUnknownCommand
	}

	if err := json.Unmarshal(message, target); err != nil {
		return nil, err
	}
	result, ok := target.(*T)
	if !ok {
		return nil, ErrTypeMismatch
	}

	return result, nil
}

var (
	ErrUnknownCommand = errors.New("unknown command name")
	ErrTypeMismatch   = errors.New("command type mismatch")
)
