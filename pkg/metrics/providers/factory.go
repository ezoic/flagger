package providers

import (
	flaggerv1 "github.com/weaveworks/flagger/pkg/apis/flagger/v1beta1"
)

type Factory struct{}

func (factory Factory) Provider(
	metricInterval string,
	provider flaggerv1.MetricTemplateProvider,
	credentials map[string][]byte,
) (Interface, error) {
	switch provider.Type {
	case "prometheus":
		return newPrometheusProvider(provider, credentials)
	case "datadog":
		return newDatadogProvider(metricInterval, provider, credentials)
	case "cloudwatch":
		return newCloudWatchProvider(provider)
	default:
		return newPrometheusProvider(provider, credentials)
	}
}
