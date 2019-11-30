package registrar

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/aws/aws-sdk-go/service/route53"
)

type MockRoute53 struct {
	*route53.Route53
}

func (m *MockRoute53) ListHostedZonesByName(input *route53.ListHostedZonesByNameInput) (*route53.ListHostedZonesByNameOutput, error) {

	success := &route53.ListHostedZonesByNameOutput{
		HostedZones: []*route53.HostedZone{
			&route53.HostedZone{
				Name: input.DNSName,
			},
		},
	}

	switch *input.DNSName {
	case "bar.com":
		return success, nil
	case "foo.bar.com":
		return success, nil
	case "beep.florp.com":
		return success, nil
	}

	return nil, fmt.Errorf("not found, got %s", *input.DNSName)

}

func TestZoneFromHost(t *testing.T) {

	m := &MockRoute53{}

	host := "foo.bar.com"

	assert.Equal(t, *findZone(host, m).Name, "bar.com", "")

	host = "baz.foo.bar.com"

	assert.Equal(t, *findZone(host, m).Name, "foo.bar.com", "")

	host = "bloop.quux.beep.florp.com"

	assert.Equal(t, *findZone(host, m).Name, "beep.florp.com", "")

}
