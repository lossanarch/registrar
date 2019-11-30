package registrar

import (
	"log"
	"net"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

func Register(host string) {

	r := newRoute53Client()

	ip := getHostIP()

	log.Printf("Got IP: %s\n", ip.String())

	zone := findZone(host, r)

	changeSet := newUpsert(newRecord(host, ip.String()))

	req, out := r.newChangeRequest(changeSet, zone.Id)

	err := req.Send()

	if err != nil {
		log.Println("[WARN] Change request failed: ", err)
	}

	log.Printf("ChangeResult: %#v", out)

}

func (r *r53Client) newChangeRequest(changeSet []*route53.Change, zoneID *string) (*request.Request, *route53.ChangeResourceRecordSetsOutput) {

	changeSetInput := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: changeSet,
		},
		HostedZoneId: zoneID,
	}

	return r.ChangeResourceRecordSetsRequest(changeSetInput)

}

func newUpsert(record *route53.ResourceRecordSet) []*route53.Change {

	return []*route53.Change{
		&route53.Change{
			Action:            aws.String("UPSERT"),
			ResourceRecordSet: record,
		},
	}

}

func newRecord(host, ip string) *route53.ResourceRecordSet {

	return &route53.ResourceRecordSet{
		Name: &host,
		Type: aws.String("A"),
		TTL:  aws.Int64(60),
		ResourceRecords: []*route53.ResourceRecord{
			&route53.ResourceRecord{
				Value: &ip,
			},
		},
	}

}

func getHostIP() net.IP {

	m := newEC2MetadataClient()

	var ip string

	if !m.Available() {
		log.Println("No metadata detected, using dummy ip")

		return net.ParseIP("127.0.0.1")
	}

	ip, err := m.GetMetadata("/public-ipv4")

	if err != nil || ip == "" {
		log.Println("Couldn't get public ip address from ec2 metadata service:", err)

		ip, err = m.GetMetadata("/local-ipv4")
		if err != nil {
			log.Fatal("Couldn't get a local ip address either")
		}
	}

	return net.ParseIP(ip)

}

type r53Client struct {
	*route53.Route53
}

func newRoute53Client() *r53Client {

	s := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	return &r53Client{route53.New(s)}

}

func newEC2MetadataClient() *ec2metadata.EC2Metadata {

	s := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	return ec2metadata.New(s)

}

func findZone(host string, r route53Client) *route53.HostedZone {

	parts := strings.Split(host, ".")

	zone := strings.Join(parts[1:len(parts)], ".")

	if zone == "" {
		panic("unable to find zone")
	}

	match := getZone(zone, r)

	if match != nil {
		return match
	}

	return findZone(zone, r)
}

type route53Client interface {
	ListHostedZonesByName(input *route53.ListHostedZonesByNameInput) (*route53.ListHostedZonesByNameOutput, error)
}

func getZone(zone string, r route53Client) *route53.HostedZone {

	zoneList, err := r.ListHostedZonesByName(
		&route53.ListHostedZonesByNameInput{
			DNSName:  &zone,
			MaxItems: aws.String("1"),
		})
	if err != nil {
		return nil
	}

	if len(zoneList.HostedZones) == 1 {
		return zoneList.HostedZones[0]
	}

	return nil

}
