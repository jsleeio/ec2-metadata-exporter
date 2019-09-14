# overview

The purpose of this little (one metric right now!) Prometheus exporter is to expose
the age of the AWS EC2 cloud image (AMI) that was used to launch an EC2 instance.

This then allows alerting on things like _tell me if I have any instances
launched with old images_, which is useful when your patching strategy involves
regularly building new cloud images and continually performing a rolling
replacement of the compute fleet. An instance launched with a too-old cloud
image probably means either a build failure somewhere, or the configuration
hasn't been updated to use the latest image.

# deployment

Pretty uncomplicated. Just run it. You don't even need to tell it an AWS
region, because it can autodiscover that from instance metadata. It listens on
`tcp/9981` by default. To operate, it needs two things:

1. access to the EC2 metadata API endpoint (http://169.254.169.254/...)
2. access to the EC2 API endpoint for the region it is running in (https://...)
2. AWS API credentials that will allow the `ec2:DescribeImages` API call.
   Provide these credentials via an IAM instance profile or equivalent.

# docker

Docker image can be found at

    gcr.io/jsleeio-containers/ec2-metadata-exporter:v1.0.0

