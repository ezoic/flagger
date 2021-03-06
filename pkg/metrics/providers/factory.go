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

	switch {
	case provider.Type == "prometheus":
		return NewPrometheusProvider(provider, credentials)
	case provider.Type == "datadog":
		return NewDatadogProvider(metricInterval, provider, credentials)
	default:
		return NewPrometheusProvider(provider, credentials)
	}
}
