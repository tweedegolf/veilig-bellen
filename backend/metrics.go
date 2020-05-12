package main

import "github.com/aws/aws-sdk-go/aws"
import "github.com/aws/aws-sdk-go/aws/session"
import "github.com/aws/aws-sdk-go/aws/credentials"
import "github.com/aws/aws-sdk-go/service/connect"

func (cfg Configuration) getConnectSession() *connect.Connect {
	mySession := session.Must(session.NewSession())
	return connect.New(mySession, &aws.Config{
		Region:      aws.String(cfg.Connect.Region),
		Credentials: credentials.NewStaticCredentials(cfg.Connect.Id, cfg.Connect.Secret, ""),
	})
}

func (cfg Configuration) getConnectCurrentMetrics() (*connect.GetCurrentMetricDataOutput, error) {
	svc := cfg.getConnectSession()

	instanceId := cfg.Connect.InstanceId

	agentsOnline := "AGENTS_ONLINE"
	agentsAvailable := "AGENTS_AVAILABLE"
	agentsOnCall := "AGENTS_ON_CALL"
	contactsInQueue := "CONTACTS_IN_QUEUE"
	count := "COUNT"

	queue := cfg.Connect.Queue

	result, err := svc.GetCurrentMetricData(&connect.GetCurrentMetricDataInput{
		InstanceId: &instanceId,
		CurrentMetrics: []*connect.CurrentMetric{
			&connect.CurrentMetric{Name: &agentsOnline, Unit: &count},
			&connect.CurrentMetric{Name: &agentsAvailable, Unit: &count},
			&connect.CurrentMetric{Name: &agentsOnCall, Unit: &count},
			&connect.CurrentMetric{Name: &contactsInQueue, Unit: &count},
		},
		Filters:   &connect.Filters{Queues: []*string{&queue}},
		Groupings: []*string{},
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (cfg Configuration) listConnectQueues() (*connect.ListQueuesOutput, error) {
	svc := cfg.getConnectSession()

	instanceId := cfg.Connect.InstanceId

	result, err := svc.ListQueues(&connect.ListQueuesInput{
		InstanceId: &instanceId,
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
