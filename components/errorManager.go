package components

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"server/config"
	"server/util"
)

var messageMap map[int64][]string
var messageCache Cache[string, map[int64][]string]

func init() {
	messageMap, _ = getMessageMap(config.GetPath("errors"))
}

func getMessageMap(file string) (map[int64][]string, error) {
	result := make(map[int64][]string)
	buf, err := os.ReadFile(file)
	if nil != err {
		util.LogError(err)
		return nil, err
	}

	lines := util.Split(string(buf))
	for _, line := range lines {
		errorList := strings.Split(line, ":")

		if 2 > len(errorList) {
			continue
		}

		messages := strings.Split(errorList[1], ",")
		if 0 < len(messages) {
			key, err := strconv.ParseInt(errorList[0], 10, 64)
			if nil != err {
				continue
			}
			result[key] = messages
		}
	}
	return result, nil
}

func GetMessage(path string, no int64, param map[string]any) string {
	pathMessageMap, err := messageCache.TryGet(path, func(s *string) (map[int64][]string, error) {
		return getMessageMap(fmt.Sprintf("%s/errors", path))
	})
	if nil != err {
		return ""
	}
	var messages []string
	var ok bool
	if nil == pathMessageMap {
		if messages, ok = messageMap[no]; !ok {
			return ""
		}
	} else {
		if messages, ok = pathMessageMap[no]; !ok {
			return ""
		}
	}
	length := len(messages)
	if length < 1 {
		return ""
	} else if length < 2 {
		return messages[0]
	}
	messageParam := make([]any, length-1)
	for i := 1; i < length; i++ {
		messageParam[i-1] = param[messages[i]]
	}
	return fmt.Sprintf(messages[0], messageParam...)
}
