package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type config struct {
	Session     *session.Session
	EC2Metadata *ec2metadata.EC2Metadata
	EC2         *ec2.EC2
	Listen      *string
	IID         ec2metadata.EC2InstanceIdentityDocument
}

func configure() *config {
	s := session.Must(session.NewSession())
	// a somewhat ugly dance here. To avoid needing a region argument passed in
	// (when the operator could reasonably expect it to be autodetected) we setup
	// an initial session to find the region, and then another to create a region
	// -specific EC2 client
	c := &config{
		Session:     s,
		EC2:         nil,
		EC2Metadata: ec2metadata.New(s),
		Listen:      flag.String("listen", ":9981", "port and optional address to listen on"),
	}
	flag.Parse()
	var err error
	c.IID, err = c.EC2Metadata.GetInstanceIdentityDocument()
	if err != nil {
		log.Fatalf("unable to get instance identity document: %v", err)
	}
	c.Session = session.Must(session.NewSession(&aws.Config{
		Region: aws.String(c.IID.Region),
	}))
	c.EC2 = ec2.New(c.Session)
	return c
}

type metricset struct {
	ImageCreatedAt prometheus.Gauge
}

func registerMetrics() *metricset {
	m := &metricset{
		ImageCreatedAt: prometheus.NewGauge(prometheus.GaugeOpts{
			Subsystem: "ec2metadata",
			Name:      "image_created_at",
			Help:      "Unix-format creation timestamp of the EC2 AMI used to launch this instance",
		}),
	}
	prometheus.MustRegister(m.ImageCreatedAt)
	return m
}

func imageCreationDate(client *ec2.EC2, ami string) time.Time {
	dii := &ec2.DescribeImagesInput{ImageIds: []*string{aws.String(ami)}}
	dio, err := client.DescribeImages(dii)
	if err != nil {
		log.Fatalf("unable to describe image: %v", err)
	}
	if n := len(dio.Images); n != 1 {
		log.Fatalf("did not find exactly one image with ID %s: found %d", ami, n)
	}
	image := dio.Images[0]
	ctime, err := time.Parse(time.RFC3339, *image.CreationDate)
	if err != nil {
		log.Fatalf("unparseable creation date: %s: %v", *image.CreationDate, err)
	}
	return ctime
}

func main() {
	cfg := configure()
	metrics := registerMetrics()
	metrics.ImageCreatedAt.Set(float64(imageCreationDate(cfg.EC2, cfg.IID.ImageID).Unix()))
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*cfg.Listen, nil))
}
