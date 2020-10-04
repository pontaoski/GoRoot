// SPDX-FileCopyrightText: 2020 Carson Black <uhhadd@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package simple

// RunAsRoot runs a command as root
// actionName = "com.github.pontaoski.RunAsRoot"
// authenticationMode = "admin"
// persistAuthentication = true
func RunAsRoot(Command string, Args []string) (Stdout, Stderr string) {
	return Command, Args[0]
}
