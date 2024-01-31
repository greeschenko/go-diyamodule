package diyamodule

import (
  "fmt"
  "github.com/go-rest-framework/core"
)

var App core.App

func Configure(a core.App) {
	App = a
}
