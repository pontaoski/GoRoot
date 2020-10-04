// SPDX-FileCopyrightText: 2020 Carson Black <uhhadd@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"errors"
	"go/ast"
	"strings"

	"github.com/BurntSushi/toml"
)

type ActionConfig struct {
	ActionName            string `toml:"actionName"`
	AuthenticationMode    string `toml:"authenticationMode"`
	PersistAuthentication bool   `toml:"persistAuthentication"`
}

func ReadComments(ast *ast.CommentGroup) (ActionConfig, error) {
	if ast == nil {
		return ActionConfig{}, errors.New("missing comment")
	}

	data := strings.Join(strings.Split(ast.Text(), "\n")[1:], "\n")
	var ac ActionConfig
	err := toml.Unmarshal([]byte(data), &ac)
	if err != nil {
		return ActionConfig{}, err
	}

	return ac, nil
}
