package util

import (
	"github.com/op/go-logging"
)

var logger *logging.Logger

func LoggerCreate() error {
	format := logging.MustStringFormatter("[%{level}] %{shortfile} %{shortfunc}(): %{message}")
	logging.SetFormatter(format)
	logger = logging.MustGetLogger("comentario")

	return nil
}
