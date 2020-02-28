package providers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"

	flaggerv1 "github.com/weaveworks/flagger/pkg/apis/flagger/v1beta1"
)

const (
	cloudWatchMaxRetries = 3
)

type cloudWatchProvider struct {
	client cloudWatchClient
}

// for the testing purpose
type cloudWatchClient interface {
	GetMetricData(input *cloudwatch.GetMetricDataInput) (*cloudwatch.GetMetricDataOutput, error)
}

func newCloudWatchProvider(provider flaggerv1.MetricTemplateProvider) (*cloudWatchProvider, error) {
	region := strings.TrimLeft(provider.Address, "monitoring.")
	region = strings.TrimRight(region, ".amazonaws.com")
	sess, err := session.NewSession(
		aws.NewConfig().
			WithRegion(region).
			WithMaxRetries(cloudWatchMaxRetries).
			WithEndpoint(provider.Address),
	)

	return &cloudWatchProvider{
		client: cloudwatch.New(sess),
	}, err
}

func (p *cloudWatchProvider) RunQuery(query string) (float64, error) {
	var cq []*cloudwatch.MetricDataQuery
	if err := json.Unmarshal([]byte(query), &cq); err != nil {
		return 0, fmt.Errorf("error unmarshaling query: %s", err.Error())
	}

	res, err := p.client.GetMetricData(&cloudwatch.GetMetricDataInput{
		EndTime:           nil,
		MaxDatapoints:     aws.Int64(1),
		StartTime:         nil,
		MetricDataQueries: cq,
	})

	if err != nil {
		return 0, fmt.Errorf("error requesting cloudwatch: %s", err.Error())
	}

	mr := res.MetricDataResults
	if len(mr) < 1 {
		return 0, fmt.Errorf("no values found in response: %s", res.String())
	}

	vs := res.MetricDataResults[0].Values
	if len(vs) < 1 {
		return 0, fmt.Errorf("no values found in response: %s", res.String())
	}

	return aws.Float64Value(vs[0]), nil
}

func (p *cloudWatchProvider) IsOnline() (bool, error) {
	_, err := p.client.GetMetricData(&cloudwatch.GetMetricDataInput{
		EndTime:           aws.Time(time.Time{}),
		MetricDataQueries: []*cloudwatch.MetricDataQuery{},
		StartTime:         aws.Time(time.Time{}),
	})

	if err == nil {
		return true, nil
	}

	ae, ok := err.(awserr.RequestFailure)
	if !ok {
		return false, fmt.Errorf("unexpected error: %v", err)
	} else if ae.StatusCode() != http.StatusBadRequest {
		return false, fmt.Errorf("unexpected status code: %v", ae)
	}
	return true, nil
}
