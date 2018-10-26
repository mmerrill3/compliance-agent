package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"os"
	"time"
	"v4e.io/compliance/agent/types"
)

var (
	wait            time.Duration
	trendMicroToken string
	s3Prefix        string
	s3Bucket        string
	targetAddress   *string
)

func init() {
	targetAddress = flag.String("target", "127.0.0.1", "target host")
	flag.Parse()
	flag.Lookup("logtostderr").Value.Set("true")
	trendMicroToken = os.Getenv("TREND_TOKEN")
	s3Prefix = os.Getenv("S3_PREFIX")
	s3Bucket = os.Getenv("S3_BUCKET")
	if s3Prefix == "" {
		s3Prefix = "test"
	}
}

func main() {
	var bodyString string
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	searchType := &types.SearchComputers{MaxItems: 1, SortByObjectID: true, SearchCriteria: []types.SearchCriteria{types.SearchCriteria{FieldName: "hostName", StringTest: "equal", StringValue: *targetAddress}}}
	client := &http.Client{Transport: tr}
	byteBuffer := new(bytes.Buffer)

	if err := json.NewEncoder(byteBuffer).Encode(searchType); err != nil {
		glog.Fatalf("Error creating json request to trend: %v", err)
	}

	req, _ := http.NewRequest("POST", "https://10.71.6.95/api/computers/search", byteBuffer)
	req.Header.Set("api-secret-key", trendMicroToken)
	req.Header.Set("api-version", "v1")
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		glog.Errorf("Error talking to remote host %v", err)
	}
	glog.V(4).Infof("Response: %s", resp)
	if resp.StatusCode == http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		bodyString = string(bodyBytes)
		glog.Infof("Response Body: %s", bodyString)
	} else {
		glog.Fatalf("Found bad response code from trend server")
	}

	// The session the S3 Uploader will use
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String("us-west-2")}))

	// Create an uploader with the session and default options
	uploader := s3manager.NewUploader(sess)

	s3Directory := s3Prefix
	s3Directory += "/trendmicro"

	// Upload the file to S3.
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(s3Directory),
		Body:   bytes.NewBufferString(bodyString),
	})
	if err != nil {
		glog.Fatalf("failed to upload file, %v", err)
	}
	glog.Infof("file uploaded to, %s\n", result.Location)
}
