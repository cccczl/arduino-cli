// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package lib

import (
	"context"
	"errors"

	"github.com/arduino/arduino-cli/arduino"
	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesmanager"
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

// LibraryUpgradeAll upgrades all the available libraries
func LibraryUpgradeAll(req *rpc.LibraryUpgradeAllRequest, downloadCB rpc.DownloadProgressCB, taskCB rpc.TaskProgressCB) error {
	lm := commands.GetLibraryManager(req)
	if lm == nil {
		return &arduino.InvalidInstanceError{}
	}

	if err := upgrade(lm, listLibraries(lm, true, false), downloadCB, taskCB); err != nil {
		return err
	}

	if err := commands.Init(&rpc.InitRequest{Instance: req.GetInstance()}, nil); err != nil {
		return err
	}

	return nil
}

// LibraryUpgrade upgrades a library
func LibraryUpgrade(ctx context.Context, req *rpc.LibraryUpgradeRequest, downloadCB rpc.DownloadProgressCB, taskCB rpc.TaskProgressCB) error {
	lm := commands.GetLibraryManager(req)

	// Get the library to upgrade
	name := req.GetName()
	lib := filterByName(listLibraries(lm, false, false), name)
	if lib == nil {
		// library not installed...
		return &arduino.LibraryNotFoundError{Library: name}
	}
	if lib.Available == nil {
		taskCB(&rpc.TaskProgress{Message: tr("Library %s is already at the latest version", name), Completed: true})
		return nil
	}

	// Install update
	return upgrade(lm, []*installedLib{lib}, downloadCB, taskCB)
}

func upgrade(lm *librariesmanager.LibrariesManager, libs []*installedLib, downloadCB rpc.DownloadProgressCB, taskCB rpc.TaskProgressCB) error {
	// Go through the list and download them
	for _, lib := range libs {
		if err := downloadLibrary(lm, lib.Available, downloadCB, taskCB); err != nil {
			return err
		}
	}

	// Go through the list and install them
	for _, lib := range libs {
		if err := installLibrary(lm, lib.Available, libraries.User, taskCB); err != nil {
			if !errors.Is(err, librariesmanager.ErrAlreadyInstalled) {
				return err
			}
		}
	}

	return nil
}

func filterByName(libs []*installedLib, name string) *installedLib {
	for _, lib := range libs {
		if lib.Library.RealName == name {
			return lib
		}
	}
	return nil
}
