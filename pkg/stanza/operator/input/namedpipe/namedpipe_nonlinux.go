// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build !linux

package namedpipe // import "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/input/namedpipe"

import (
	"errors"

	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
)

func (c *Config) Build(_ *zap.SugaredLogger, meter metric.Meter) (operator.Operator, error) {
	return nil, errors.New("namedpipe input operator is only supported on linux")
}
