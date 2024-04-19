/**
 * TODO:文件描述
 *
 * @title logger
 * @projectName grpcDemo
 * @author niuk
 * @date 2024/4/18 8:35
 */

package discovery

import (
	"fmt"
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
	"runtime"
	"strings"
)

func customFormatter(caller *runtime.Frame) string {
	file := caller.File
	index := strings.LastIndex(file, "/")
	if index > 0 {
		file = file[index+1:]
	}
	return fmt.Sprintf(" %s:%d", file, caller.Line)
}

func loggerInit() {
	logger = logrus.New()
	formatter := &nested.Formatter{
		FieldsOrder:           nil,
		TimestampFormat:       "2006-01-02 15:04:05",
		NoColors:              true,
		CallerFirst:           true,
		CustomCallerFormatter: customFormatter,
	}
	logger.SetFormatter(formatter)
}
