package main

import (
	"log"
	"os"

	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/s3"
)

func main() {
	auth, err := aws.EnvAuth()

	if err != nil {
		log.Fatal(err)
	}

	b := s3.New(auth, aws.USEast).Bucket("xyz")

	rc, err := b.GetReader(os.Args[0])
}
