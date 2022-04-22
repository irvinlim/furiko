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
	"context"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewGetCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a single resource.",
	}

	cmd.AddCommand(NewGetJobCommand(ctx))

	return cmd
}

func NewGetJobCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "job",
		Aliases: []string{"jobs"},
		Short:   "Displays information about a single Job.",
		PreRunE: PrerunWithKubeconfig,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return ExecuteGetJob(ctx, cmd, args)
		},
	}

	return cmd
}

func ExecuteGetJob(ctx context.Context, cmd *cobra.Command, args []string) error {
	client := ctrlContext.Clientsets().Furiko().ExecutionV1alpha1()
	namespace, err := cmd.Flags().GetString("namespace")
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return errors.New("name is required")
	}
	name := args[0]

	job, err := client.Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "cannot list jobs")
	}

	// TODO(irvinlim): We simply print as a table for now, but the idea is that we
	//  want to show more information for a single job.
	PrintTable([]string{
		"NAME",
		"PHASE",
		"START TIME",
		"RUN TIME",
		"FINISH TIME",
	}, [][]string{makeJobRow(job)})

	return nil
}
