// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/sirupsen/logrus"
)

func main() {
	log := logrus.StandardLogger()
	cmd := app.NewCommand(log)
	if err := cmd.Execute(); err != nil {
		return
	}
	err := cmd.Execute()
	if err != nil {
		log.Error("error executing main command: %s", err)
		os.Exit(1)
	}
}
