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

	execution "github.com/furiko-io/furiko/apis/execution/v1alpha1"
)

func NewListCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List multiple resources.",
	}

	cmd.AddCommand(NewListJobCommand(ctx))

	return cmd
}

func NewListJobCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "job",
		Aliases: []string{"jobs"},
		Short:   "Displays information about multiple Jobs.",
		PreRunE: PrerunWithKubeconfig,
		RunE: func(cmd *cobra.Command, args []string) error {
			return ExecuteListJob(ctx, cmd, args)
		},
	}

	return cmd
}

func ExecuteListJob(ctx context.Context, cmd *cobra.Command, _ []string) error {
	client := ctrlContext.Clientsets().Furiko().ExecutionV1alpha1()
	namespace, err := cmd.Flags().GetString("namespace")
	if err != nil {
		return err
	}

	jobList, err := client.Jobs(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrapf(err, "cannot list jobs")
	}

	PrintTable([]string{
		"NAME",
		"PHASE",
		"START TIME",
		"RUN TIME",
		"FINISH TIME",
	}, makeJobRows(jobList))

	return nil
}

func makeJobRows(jobList *execution.JobList) [][]string {
	rows := make([][]string, 0, len(jobList.Items))
	for _, item := range jobList.Items {
		item := item
		rows = append(rows, makeJobRow(&item))
	}
	return rows
}

func makeJobRow(job *execution.Job) []string {
	startTime := FormatTimeAgo(job.Status.StartTime)
	runTime := ""
	finishTime := ""

	if condition := job.Status.Condition.Running; condition != nil {
		runTime = FormatTimeAgo(&condition.StartedAt)
	}
	if condition := job.Status.Condition.Finished; condition != nil {
		finishTime = FormatTimeAgo(&condition.FinishedAt)
	}

	return []string{
		job.Name,
		string(job.Status.Phase),
		startTime,
		runTime,
		finishTime,
	}
}
