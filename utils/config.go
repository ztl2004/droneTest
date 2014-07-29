package utils

import (
	"bufio"
	"io"
	"os"
	"strings"
)

func LoadConfig(path string) (map[string]string, error) {
	configMap := make(map[string]string)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buff := bufio.NewReader(f)

	for {
		line, err := buff.ReadString('\n')
		if err != nil || io.EOF == err {
			break
		}
		if strings.HasPrefix(line, "LOG_LEVEL") {
			configMap["LOG_LEVEL"] = strings.TrimRight(strings.Split(line, "=")[1], "\n")
		}
	}
	return configMap, nil
}
