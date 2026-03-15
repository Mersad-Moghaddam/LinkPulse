package logger

import (
	"log"
	"os"
)

func New() *log.Logger { return log.New(os.Stdout, "linkpulse ", log.LstdFlags|log.LUTC) }
