/*
 * Copyright 2022 The Furiko Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package prompt

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"

	execution "github.com/furiko-io/furiko/apis/execution/v1alpha1"
	stringsutils "github.com/furiko-io/furiko/pkg/utils/strings"
)

// Prompt knows how to prompt the user for input, and returns the formatted
// option value.
type Prompt interface {
	Run() (interface{}, error)
}

// MakePrompt returns a general Prompt based on the given Option.
func MakePrompt(option execution.Option) (Prompt, error) {
	switch option.Type {
	case execution.OptionTypeBool:
		return NewBoolSelect(option), nil
	case execution.OptionTypeString:
		return NewStringPrompt(option), nil
	case execution.OptionTypeSelect:
		return NewSelectPrompt(option), nil
	}

	return nil, fmt.Errorf("unhandled option type: %v", option.Type)
}

type boolPrompt struct {
	cfg      *execution.BoolOptionConfig
	promptui *promptui.Prompt
}

var _ Prompt = (*boolPrompt)(nil)

// NewBoolPrompt returns a new Prompt from a Bool option.
func NewBoolPrompt(option execution.Option) Prompt {
	cfg := option.Bool
	if cfg == nil {
		cfg = &execution.BoolOptionConfig{}
	}

	defaultVal := "N"
	if cfg.Default {
		defaultVal = "Y"
	}

	return &boolPrompt{
		cfg: cfg,
		promptui: &promptui.Prompt{
			Label:     MakeLabel(option),
			IsConfirm: true,
			Default:   defaultVal,
		},
	}
}

func (p *boolPrompt) Run() (interface{}, error) {
	return p.promptui.Run()
}

type boolSelect struct {
	cfg      *execution.BoolOptionConfig
	promptui *promptui.Select
}

var _ Prompt = (*boolSelect)(nil)

// NewBoolSelect returns a new Prompt from a Bool option, using a Select instead.
func NewBoolSelect(option execution.Option) Prompt {
	cfg := option.Bool
	if cfg == nil {
		cfg = &execution.BoolOptionConfig{}
	}

	cursorPos := 1
	if cfg.Default {
		cursorPos = 0
	}

	return &boolSelect{
		cfg: cfg,
		promptui: &promptui.Select{
			Label:     MakeLabel(option),
			Items:     []string{"Yes", "No"},
			CursorPos: cursorPos,
		},
	}
}

func (p *boolSelect) Run() (interface{}, error) {
	index, _, err := p.promptui.Run()
	if err != nil {
		return "", err
	}

	val := index == 0
	return val, nil
}

type stringPrompt struct {
	option   execution.Option
	cfg      *execution.StringOptionConfig
	promptui *promptui.Prompt
}

var _ Prompt = (*stringPrompt)(nil)

// NewStringPrompt returns a Prompt from a String option.
func NewStringPrompt(option execution.Option) Prompt {
	cfg := option.String
	if cfg == nil {
		cfg = &execution.StringOptionConfig{}
	}

	return &stringPrompt{
		option: option,
		cfg:    cfg,
		promptui: &promptui.Prompt{
			Label:     MakeLabel(option),
			Default:   cfg.Default,
			AllowEdit: true,
			Validate: func(s string) error {
				if s == "" {
					return errors.New("value is required")
				}
				return nil
			},
		},
	}
}

func (p *stringPrompt) Run() (interface{}, error) {
	return p.promptui.Run()
}

type selectPrompt struct {
	cfg      *execution.SelectOptionConfig
	promptui *promptui.Select
}

var _ Prompt = (*selectPrompt)(nil)

// NewSelectPrompt returns a new Prompt from a Select option.
func NewSelectPrompt(option execution.Option) Prompt {
	cfg := option.Select
	if cfg == nil {
		cfg = &execution.SelectOptionConfig{}
	}

	var cursorPos int
	idx, ok := stringsutils.IndexOf(cfg.Values, cfg.Default)
	if ok {
		cursorPos = idx
	}

	return &selectPrompt{
		cfg: cfg,
		promptui: &promptui.Select{
			Label:     MakeLabel(option),
			Items:     cfg.Values,
			CursorPos: cursorPos,
		},
	}
}

func (p *selectPrompt) Run() (interface{}, error) {
	index, _, err := p.promptui.Run()
	if err != nil {
		return "", err
	}

	val := index == 0
	return val, nil
}

func MakeLabel(option execution.Option) string {
	if option.Label != "" {
		return option.Label
	}
	return option.Name
}
