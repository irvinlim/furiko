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
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	execution "github.com/furiko-io/furiko/apis/execution/v1alpha1"
	"github.com/furiko-io/furiko/pkg/furictl/util/prompt"
)

func NewRunCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "run",
		Short:   "Run a new Job.",
		Long:    "Runs a new Job from a JobConfig, prompting for option values.",
		Args:    cobra.ExactArgs(1),
		PreRunE: PrerunWithKubeconfig,
		RunE: func(cmd *cobra.Command, args []string) error {
			return ExecuteRunJob(ctx, cmd, args)
		},
	}

	cmd.Flags().Bool("no-prompt", false, "If specified, will not show a prompt. This may result "+
		"in an error when an interactive prompt is required, such as for option values.")
	cmd.Flags().Bool("default-options", false, "If specified, will use default values where available. "+
		"Any options without defaults will still show a prompt.")
	cmd.Flags().String("at", "", "RFC3339-formatted datetime to specify the time to run the job at. "+
		"Implies --concurrency-policy=Enqueue unless explicitly specified.")
	cmd.Flags().String("concurrency-policy", "", "Specify a concurrency policy to use for the job, "+
		"which overrides the JobConfig's concurrency policy.")

	return cmd
}

func ExecuteRunJob(ctx context.Context, cmd *cobra.Command, args []string) error {
	client := ctrlContext.Clientsets().Furiko().ExecutionV1alpha1()
	namespace, err := GetNamespace(cmd)
	if err != nil {
		return err
	}

	// TODO(irvinlim): Support running independent job from file.
	if len(args) == 0 {
		return errors.New("job config name must be specified")
	}
	name := args[0]

	jobConfig, err := client.JobConfigs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "cannot get job config")
	}

	// Prepare fields.
	optionValues, err := makeJobOptionValues(jobConfig)
	if err != nil {
		return errors.Wrapf(err, "cannot prepare job option values")
	}
	startPolicy, err := makeJobStartPolicy(cmd)
	if err != nil {
		return errors.Wrapf(err, "cannot prepare start policy")
	}

	// Create a new job using configName.
	newJob := &execution.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: jobConfig.Name + "-",
			Namespace:    jobConfig.Namespace,
		},
		Spec: execution.JobSpec{
			StartPolicy:  startPolicy,
			OptionValues: optionValues,
		},
	}
	createdJob, err := client.Jobs(namespace).Create(ctx, newJob, metav1.CreateOptions{})
	if err != nil {
		return errors.Wrapf(err, "cannot create job")
	}

	key, err := cache.MetaNamespaceKeyFunc(createdJob)
	if err != nil {
		return errors.Wrapf(err, "key func error")
	}

	fmt.Printf("Job %v created\n", key)
	return nil
}

func makeJobStartPolicy(cmd *cobra.Command) (*execution.StartPolicySpec, error) {
	startPolicy := &execution.StartPolicySpec{}
	if concurrencyPolicy, err := cmd.Flags().GetString("concurrency-policy"); err != nil {
		return nil, err
	} else if concurrencyPolicy != "" {
		startPolicy.ConcurrencyPolicy = execution.ConcurrencyPolicy(concurrencyPolicy)
	}
	if startAfter, err := cmd.Flags().GetString("at"); err != nil {
		return nil, err
	} else if startAfter != "" {
		parsed, err := time.Parse(time.RFC3339, startAfter)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid time: %v", startAfter)
		}
		startAfterTime := metav1.NewTime(parsed)
		startPolicy.StartAfter = &startAfterTime
	}
	return startPolicy, nil
}

func makeJobOptionValues(jobConfig *execution.JobConfig) (string, error) {
	if jobConfig.Spec.Option == nil {
		return "", nil
	}

	values := make(map[string]interface{}, len(jobConfig.Spec.Option.Options))
	for _, option := range jobConfig.Spec.Option.Options {
		// Use default values if desired

		// Run the prompt if required.
		prompter, err := prompt.MakePrompt(option)
		if err != nil {
			return "", errors.Wrapf(err, "cannot make prompt for option: %v", option.Name)
		}
		value, err := prompter.Run()
		if err != nil {
			return "", errors.Wrapf(err, "prompt error")
		}

		klog.V(4).Infof(`evaluated option "%v" value as "%v"`, option.Name, value)

		values[option.Name] = value
	}

	marshaled, err := json.Marshal(values)
	if err != nil {
		return "", errors.Wrapf(err, "marshal error")
	}

	klog.V(2).Infof(`evaluated option values: %v`, string(marshaled))

	return string(marshaled), nil
}
