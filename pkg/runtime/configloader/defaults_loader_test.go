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

package configloader_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"

	configv1alpha1 "github.com/furiko-io/furiko/apis/config/v1alpha1"
	"github.com/furiko-io/furiko/pkg/runtime/configloader"
)

func TestDefaultsLoader(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	loader := configloader.NewDefaultsLoader()
	loader.Defaults = map[configv1alpha1.ConfigName]runtime.Object{
		configv1alpha1.JobExecutionConfigName: &configv1alpha1.JobExecutionConfig{
			DefaultTTLSecondsAfterFinished: pointer.Int64(123),
			DefaultPendingTimeoutSeconds:   pointer.Int64(234),
		},
	}

	mgr := configloader.NewConfigManager()
	mgr.AddConfigLoaders(loader)
	err := mgr.Start(ctx)
	assert.NoError(t, err)

	cfg, err := loadJobControllerConfig(mgr)
	assert.NoError(t, err)
	assert.Equal(t, pointer.Int64(123), cfg.DefaultTTLSecondsAfterFinished)
	assert.Equal(t, pointer.Int64(234), cfg.DefaultPendingTimeoutSeconds)

	// Unset fields should be 0
	assert.Zero(t, cfg.DeleteKillingTasksTimeoutSeconds)
	assert.Zero(t, cfg.ForceDeleteKillingTasksTimeoutSeconds)

	// Empty configuration.
	cronCfg, err := loadCronControllerConfig(mgr)
	assert.NoError(t, err)
	assert.Empty(t, cronCfg.MaxMissedSchedules)
	assert.Empty(t, cronCfg.MaxDowntimeThresholdSeconds)
}

func TestDefaultsLoader_LoaderOverride(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	defaultLoader := configloader.NewDefaultsLoader()
	defaultLoader.Defaults = map[configv1alpha1.ConfigName]runtime.Object{
		configv1alpha1.JobExecutionConfigName: &configv1alpha1.JobExecutionConfig{
			DefaultTTLSecondsAfterFinished: pointer.Int64(123),
			DefaultPendingTimeoutSeconds:   pointer.Int64(234),
		},
	}
	mockLoader := newMockConfigLoader(MockConfig{
		configv1alpha1.CronExecutionConfigName: map[string]interface{}{
			"maxMissedSchedules": 100,
		},
	})

	mgr := configloader.NewConfigManager()
	mgr.AddConfigLoaders(defaultLoader, mockLoader)
	err := mgr.Start(ctx)
	assert.NoError(t, err)

	// Non-empty CronExecutionConfig overridden by mock loader.
	cronCfg, err := loadCronControllerConfig(mgr)
	assert.NoError(t, err)
	assert.Equal(t, pointer.Int64(100), cronCfg.MaxMissedSchedules)
	assert.Empty(t, cronCfg.MaxDowntimeThresholdSeconds)
}
