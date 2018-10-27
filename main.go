package main

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
	"v4e.io/compliance/agent/tasks"
	"v4e.io/compliance/agent/types"
)

var (
	wait            time.Duration
	trendMicroToken *string
	s3Bucket        *string
	targetAddress   *string
	targetCmd       *string
	targetFile      *string
	targetPrefix    *string
	targetUser      *string
	keyFile         *string
	taskSlice       []tasks.Task
)

func init() {
	targetAddress = flag.String("target", "", "target host")
	targetCmd = flag.String("cmd", "", "command to run through ssh")
	s3Bucket = flag.String("s3Bucket", "", "S3 bucket to store the results")
	targetFile = flag.String("file", "", "file to store the output of the ssh command")
	targetPrefix = flag.String("prefix", "mpm", "prefix to store the output of the commands")
	targetUser = flag.String("username", "", "ssh username, must be passed when cmd is entered")
	keyFile = flag.String("keyFile", "", "ssh private key file, must be passed when cmd is entered")
	trendMicroToken = flag.String("trendMicroToken", "", "token to access TrendMicro API")
	flag.Parse()
	flag.Lookup("logtostderr").Value.Set("true")
	if *targetAddress == "" {
		glog.Fatal("Must pass in a target")
	}
	if *targetCmd != "" {
		if *targetUser == "" || *keyFile == "" {
			glog.Fatal("Cannot launch remote command without username and keyfile")
		}
		if *targetFile == "" {
			glog.Fatal("Cannot launch remote command without naming a file for output")
		}
		remoteTask := &tasks.RemoteAccessTask{User: *targetUser, KeyFile: *keyFile, Host: *targetAddress, FileName: *targetFile}
		taskSlice = append(taskSlice, remoteTask)
	}
}

func main() {

	if len(taskSlice) > 0 {
		for _, task := range taskSlice {
			result, err := task.Build(*targetCmd)
			if err != nil {
				glog.Fatalf("error in running task %v", err)
			}
			storeInS3(result, task.GetFileName())
		}
		return
	}

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
	req.Header.Set("api-secret-key", *trendMicroToken)
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

	storeInS3(bodyString, "trendMicro")

}

func storeInS3(payload, filename string) {
	// The session the S3 Uploader will use
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String("us-west-2")}))

	// Create an uploader with the session and default options
	uploader := s3manager.NewUploader(sess)

	hash := sha256.New()
	hash.Write([]byte(payload))
	hashValue := hash.Sum(nil)

	s3Directory := *targetPrefix
	s3Directory += "/"
	s3Directory += filename

	glog.Infof("sha256 hash value for upload: %x", hashValue)
	// Upload the file to S3.
	metaMap := make(map[string]*string)
	hashValueStr := fmt.Sprintf("%x", hashValue)
	hashValueStr = url.QueryEscape(hashValueStr)
	metaMap["sha256"] = &hashValueStr
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:   aws.String(*s3Bucket),
		Key:      aws.String(s3Directory),
		Body:     bytes.NewBufferString(payload),
		Metadata: metaMap,
	})
	if err != nil {
		glog.Fatalf("failed to upload file, %v", err)
	}
	glog.Infof("file uploaded to, %s\n", result.Location)
}
