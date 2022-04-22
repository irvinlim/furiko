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

package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	configv1alpha1 "github.com/furiko-io/furiko/apis/config/v1alpha1"
	"github.com/furiko-io/furiko/pkg/runtime/controllercontext"
)

var (
	ctrlContext controllercontext.Context
)

// NewContext returns a common context from the cobra command.
func NewContext(cmd *cobra.Command) (controllercontext.Context, error) {
	kubeconfig, err := ctrl.GetConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get kubeconfi")
	}
	return controllercontext.NewForConfig(kubeconfig, &configv1alpha1.BootstrapConfigSpec{})
}

// PrerunWithKubeconfig is a pre-run function that will set up the common context when kubeconfig is needed.
// TODO(irvinlim): We currently reuse controllercontext, but most of it is unusable for CLI interfaces.
//  We should create a new common context as needed.
func PrerunWithKubeconfig(cmd *cobra.Command, _ []string) error {
	newContext, err := NewContext(cmd)
	if err != nil {
		return err
	}
	ctrlContext = newContext
	return nil
}

// GetNamespace returns the namespace to use depending on what was defined in the flags.
func GetNamespace(cmd *cobra.Command) (string, error) {
	namespace, err := cmd.Flags().GetString("namespace")
	if err != nil {
		return "", err
	}
	if namespace == "" {
		namespace = metav1.NamespaceDefault
	}
	return namespace, nil
}
