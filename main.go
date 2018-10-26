package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	wait       time.Duration
	configFile *string
)

func init() {
	//configFile = flag.String("configurationFile", "/usr/local/compliance-agent/app.toml", "configuration file")
	flag.Parse()
	flag.Lookup("logtostderr").Value.Set("true")
	//config := readConfig()
}

func main() {
	var bodyString string
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	client := &http.Client{Transport: tr}
	req, _ := http.NewRequest("GET", "https://10.71.6.95/api/computers/2283", nil)
	req.Header.Set("api-secret-key", "3:u7Dhsbs66vhSwRrNmWu3RCOWUyBVuZEBAJcj0DSTg5k=")
	req.Header.Set("api-version", "v1")
	resp, err := client.Do(req)
	if err != nil {
		glog.Errorf("Error talking to remote host %x", err)
	}
	glog.V(4).Infof("Response: %s", resp)
	if resp.StatusCode == http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		bodyString = string(bodyBytes)
		glog.Infof("Response Body: %s", bodyString)
	}

	// The session the S3 Uploader will use
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String("us-west-2")}))

	// Create an uploader with the session and default options
	uploader := s3manager.NewUploader(sess)

	// Upload the file to S3.
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String("hackathontest22"),
		Key:    aws.String("test"),
		Body:   bytes.NewBufferString(bodyString),
	})
	if err != nil {
		glog.Fatalf("failed to upload file, %v", err)
	}
	glog.Infof("file uploaded to, %s\n", result.Location)
}
