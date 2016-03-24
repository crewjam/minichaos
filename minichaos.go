package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/crewjam/awsregion"
	"github.com/crewjam/ec2cluster"
)

func CurrentAutoScalingGroup(awsSession *session.Session) (*autoscaling.Group, error) {
	instanceID, err := ec2cluster.DiscoverInstanceID()
	if err != nil {
		return nil, err
	}

	c := ec2cluster.Cluster{
		AwsSession: awsSession,
		InstanceID: instanceID,
	}
	return c.AutoscalingGroup()
}

func GetAutoscalingGroupByName(awsSession *session.Session, autoscalingGroupName string) (*autoscaling.Group, error) {
	autoscalingService := autoscaling.New(awsSession)
	groupInfo, err := autoscalingService.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String(autoscalingGroupName)},
		MaxRecords:            aws.Int64(1),
	})
	if err != nil {
		return nil, err
	}
	if len(groupInfo.AutoScalingGroups) != 1 {
		return nil, fmt.Errorf("cannot find autoscaling group %s", autoscalingGroupName)
	}
	return groupInfo.AutoScalingGroups[0], nil
}

func TerminateRandomInstanceFromASG(awsSession *session.Session, asg *autoscaling.Group, dryRun bool) error {
	for _, instance := range asg.Instances {
		log.Printf("%s: %s %s", *instance.InstanceId, *instance.HealthStatus, *instance.LifecycleState)
		if *instance.HealthStatus != "Healthy" {
			return fmt.Errorf("chaos aborted because %s health is %s", *instance.InstanceId, *instance.HealthStatus)
		}
		if *instance.LifecycleState != autoscaling.LifecycleStateInService {
			return fmt.Errorf("chaos aborted because %s lifecycle state is %s",
				*instance.InstanceId, *instance.LifecycleState)
		}
	}

	// select an instance randomly
	i, err := rand.Int(rand.Reader, big.NewInt(int64(len(asg.Instances))))
	if err != nil {
		return err
	}
	instance := asg.Instances[int(i.Int64())]

	_, err = ec2.New(awsSession).TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: []*string{instance.InstanceId},
		DryRun:      &dryRun,
	})
	if dryRun && err != nil && strings.Contains(err.Error(), "DryRunOperation") {
		err = nil
	}
	if err != nil {

		return err
	}

	dryRunStr := ""
	if dryRun {
		dryRunStr = " (dry run)"
	}
	log.Printf("%s: terminated%s", *instance.InstanceId, dryRunStr)
	return nil
}

func main() {
	asgName := flag.String("asg", "", "The name of the autoscaling group. If not specified, the ASG of the currently running instance is used.")
	dryRun := flag.Bool("dry-run", false, "If true don't actually terminate anything, just pretend to")
	flag.Parse()

	awsSession := session.New()
	if region := os.Getenv("AWS_REGION"); region != "" {
		awsSession.Config.WithRegion(region)
	}
	awsregion.GuessRegion(awsSession.Config)

	var err error
	var asg *autoscaling.Group
	if *asgName == "" {
		asg, err = CurrentAutoScalingGroup(awsSession)
	} else {
		asg, err = GetAutoscalingGroupByName(awsSession, *asgName)
	}
	if err != nil {
		log.Fatalf("autoscaling: %s", err)
	}

	if err := TerminateRandomInstanceFromASG(awsSession, asg, *dryRun); err != nil {
		log.Fatalf("error: %s", err)
	}
}
