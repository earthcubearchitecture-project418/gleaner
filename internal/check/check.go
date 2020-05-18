package check

import (
	"fmt"
	"log"
	"net/http"

	"github.com/minio/minio-go"
)

var bl []string

func init() {
	// bl := []string{"gleaner", "gleaner-config", "gleaner-summoned", "gleaner-milled", "gleaner-shacl", "gleaner-voc"}
	// bl = []string{"gleaner", "gleaner-summoned", "gleaner-milled", "gleaner-assets"}
	bl = []string{"gleaner"}
}

// ConnCheck check the connections iwth a list buckets call
func ConnCheck(mc *minio.Client) error {
	_, err := mc.ListBuckets()
	return err
}

// Buckets checks the setup
func Buckets(mc *minio.Client) error {
	var err error

	for i := range bl {
		found, err := mc.BucketExists(bl[i])
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("Unable to locate required bucket.  Did you run with -setup the first time? Missing bucket: %s", bl[i])
		}
		if found {
			log.Printf("Verified Gleaner bucket: %s.\n", bl[i])
		}
	}

	return err
}

// MakeBuckets checks the setup
func MakeBuckets(mc *minio.Client) error {
	var err error

	for i := range bl {
		found, err := mc.BucketExists(bl[i])
		if err != nil {
			log.Printf("Existing bucket %s check:%v\n", bl[i], err)
		}
		if found {
			log.Printf("Gleaner Bucket %s found.\n", bl[i])
		} else {
			log.Printf("Gleaner Bucket %s not found, generating\n", bl[i])
			err = mc.MakeBucket(bl[i], "us-east-1") // location is kinda meaningless here
			if err != nil {
				log.Printf("Make bucket:%v\n", err)
			}
		}
	}

	return err
}

// need to check, tika, shacl, headless
// this is just a place holder for that work now
// TODO this is just a placeholder not function..  needs to be finished
// to loop on the services I need in place for gleaner
func urlCheck() bool {
	s := false

	resp, err := http.Get("http://localhost:7001") // 9998  7000 32772
	if err != nil {
		return s
	}

	if resp.Status == "200 OK" {
		s = true
	}
	return s
}

// syscheck is a place holder function for work to be done....
func syscheck(mc *minio.Client) {
	log.Println("System setup check placeholder")
	s := "NA"
	log.Printf("System check results: %s\n", s)
}
