// Package main runs the multi-arm-calibration Viam module.
package main

import (
	"go.viam.com/rdk/module"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/services/generic"

	"github.com/viam-labs/multi-arm-calibration/touch"
)

func main() {
	module.ModularMain(
		resource.APIModel{API: generic.API, Model: touch.Model},
	)
}
