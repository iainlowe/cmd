package main

import (
	"fmt"

	"log"

	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/ec2"
)

func main() {
	a, _ := aws.EnvAuth()
	c := ec2.New(a, aws.USEast)

	f := ec2.NewFilter()
	f.Add("state", "running")

	ilist, err := c.Instances(nil, nil)

	if err != nil {
		log.Fatalln(err)
	}

	for _, res := range ilist.Reservations {
		for _, ins := range res.Instances {
			t := ins.Tags
			name := "unknown"
			for _, tag := range t {
				if tag.Key == "Name" {
					name = tag.Value
					break
				}
			}

			if ins.State.Name != "running" {
				continue
			}

			fmt.Printf("%15s %s\n", ins.IPAddress, name)
		}
	}
}
