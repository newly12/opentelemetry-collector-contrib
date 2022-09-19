// Copyright 2020, OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kubeletstatsreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver"

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"go.opentelemetry.io/collector/service/featuregate"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver/internal/kubelet"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver/internal/metadata"
)

const (
	emitMetricsWithDirectionAttributeFeatureGateID    = "receiver.kubeletstatsreceiver.emitMetricsWithDirectionAttribute"
	emitMetricsWithoutDirectionAttributeFeatureGateID = "receiver.kubeletstatsreceiver.emitMetricsWithoutDirectionAttribute"
)

var (
	emitMetricsWithDirectionAttributeFeatureGate = featuregate.Gate{
		ID:      emitMetricsWithDirectionAttributeFeatureGateID,
		Enabled: true,
		Description: "Some kubeletstats metrics reported are transitioning from being reported with a direction " +
			"attribute to being reported with the direction included in the metric name to adhere to the " +
			"OpenTelemetry specification. This feature gate controls emitting the old metrics with the direction " +
			"attribute. For more details, see: " +
			"https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/kubeletstatsreceiver/README.md#feature-gate-configurations",
	}

	emitMetricsWithoutDirectionAttributeFeatureGate = featuregate.Gate{
		ID:      emitMetricsWithoutDirectionAttributeFeatureGateID,
		Enabled: false,
		Description: "Some kubeletstats metrics reported are transitioning from being reported with a direction " +
			"attribute to being reported with the direction included in the metric name to adhere to the " +
			"OpenTelemetry specification. This feature gate controls emitting the new metrics without the direction " +
			"attribute. For more details, see: " +
			"https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/kubeletstatsreceiver/README.md#feature-gate-configurations",
	}
)

func init() {
	featuregate.GetRegistry().MustRegister(emitMetricsWithDirectionAttributeFeatureGate)
	featuregate.GetRegistry().MustRegister(emitMetricsWithoutDirectionAttributeFeatureGate)
}

type scraperOptions struct {
	id                    config.ComponentID
	collectionInterval    time.Duration
	extraMetadataLabels   []kubelet.MetadataLabel
	metricGroupsToCollect map[kubelet.MetricGroup]bool
	k8sAPIClient          kubernetes.Interface
}

type kubletScraper struct {
	statsProvider                        *kubelet.StatsProvider
	metadataProvider                     *kubelet.MetadataProvider
	logger                               *zap.Logger
	extraMetadataLabels                  []kubelet.MetadataLabel
	metricGroupsToCollect                map[kubelet.MetricGroup]bool
	k8sAPIClient                         kubernetes.Interface
	cachedVolumeLabels                   map[string][]metadata.ResourceMetricsOption
	mbs                                  *metadata.MetricsBuilders
	emitMetricsWithDirectionAttribute    bool
	emitMetricsWithoutDirectionAttribute bool
}

func logDeprecatedFeatureGateForDirection(log *zap.Logger, gate featuregate.Gate) {
	log.Warn("WARNING: The " + gate.ID + " feature gate is deprecated and will be removed in the next release. The change to remove " +
		"the direction attribute has been reverted in the specification. See https://github.com/open-telemetry/opentelemetry-specification/issues/2726 " +
		"for additional details.")
}

func newKubletScraper(
	restClient kubelet.RestClient,
	set component.ReceiverCreateSettings,
	rOptions *scraperOptions,
	metricsConfig metadata.MetricsSettings,
) (scraperhelper.Scraper, error) {
	ks := &kubletScraper{
		statsProvider:         kubelet.NewStatsProvider(restClient),
		metadataProvider:      kubelet.NewMetadataProvider(restClient),
		logger:                set.Logger,
		extraMetadataLabels:   rOptions.extraMetadataLabels,
		metricGroupsToCollect: rOptions.metricGroupsToCollect,
		k8sAPIClient:          rOptions.k8sAPIClient,
		cachedVolumeLabels:    make(map[string][]metadata.ResourceMetricsOption),
		mbs: &metadata.MetricsBuilders{
			NodeMetricsBuilder:      metadata.NewMetricsBuilder(metricsConfig, set.BuildInfo),
			PodMetricsBuilder:       metadata.NewMetricsBuilder(metricsConfig, set.BuildInfo),
			ContainerMetricsBuilder: metadata.NewMetricsBuilder(metricsConfig, set.BuildInfo),
			OtherMetricsBuilder:     metadata.NewMetricsBuilder(metricsConfig, set.BuildInfo),
		},
		emitMetricsWithDirectionAttribute:    featuregate.GetRegistry().IsEnabled(emitMetricsWithDirectionAttributeFeatureGateID),
		emitMetricsWithoutDirectionAttribute: featuregate.GetRegistry().IsEnabled(emitMetricsWithoutDirectionAttributeFeatureGateID),
	}
	if !ks.emitMetricsWithDirectionAttribute {
		logDeprecatedFeatureGateForDirection(ks.logger, emitMetricsWithDirectionAttributeFeatureGate)
	}

	if ks.emitMetricsWithoutDirectionAttribute {
		logDeprecatedFeatureGateForDirection(ks.logger, emitMetricsWithoutDirectionAttributeFeatureGate)
	}
	return scraperhelper.NewScraper(typeStr, ks.scrape)
}

func (r *kubletScraper) scrape(context.Context) (pmetric.Metrics, error) {
	summary, err := r.statsProvider.StatsSummary()
	if err != nil {
		r.logger.Error("call to /stats/summary endpoint failed", zap.Error(err))
		return pmetric.Metrics{}, err
	}

	var podsMetadata *v1.PodList
	// fetch metadata only when extra metadata labels are needed
	if len(r.extraMetadataLabels) > 0 {
		podsMetadata, err = r.metadataProvider.Pods()
		if err != nil {
			r.logger.Error("call to /pods endpoint failed", zap.Error(err))
			return pmetric.Metrics{}, err
		}
	}

	metadata := kubelet.NewMetadata(r.extraMetadataLabels, podsMetadata, r.detailedPVCLabelsSetter())
	mds := kubelet.MetricsData(r.logger, summary, metadata, r.metricGroupsToCollect, r.mbs, r.emitMetricsWithDirectionAttribute, r.emitMetricsWithoutDirectionAttribute)
	md := pmetric.NewMetrics()
	for i := range mds {
		mds[i].ResourceMetrics().MoveAndAppendTo(md.ResourceMetrics())
	}
	return md, nil
}

func (r *kubletScraper) detailedPVCLabelsSetter() func(volCacheID, volumeClaim, namespace string) ([]metadata.ResourceMetricsOption, error) {
	return func(volCacheID, volumeClaim, namespace string) ([]metadata.ResourceMetricsOption, error) {
		if r.k8sAPIClient == nil {
			return nil, nil
		}

		if r.cachedVolumeLabels[volCacheID] == nil {
			ctx := context.Background()
			pvc, err := r.k8sAPIClient.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, volumeClaim, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}

			volName := pvc.Spec.VolumeName
			if volName == "" {
				return nil, fmt.Errorf("PersistentVolumeClaim %s does not have a volume name", pvc.Name)
			}

			pv, err := r.k8sAPIClient.CoreV1().PersistentVolumes().Get(ctx, volName, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}

			ro := kubelet.GetPersistentVolumeLabels(pv.Spec.PersistentVolumeSource)

			// Cache collected labels.
			r.cachedVolumeLabels[volCacheID] = ro
		}
		return r.cachedVolumeLabels[volCacheID], nil
	}
}
