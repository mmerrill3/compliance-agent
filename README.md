# compliance-agent
Proof of concept for a compliance agent that publishes findings to s3

The environment variables to set for this script are as follows

* S3_BUCKET -> The S3 bucket to publish information to
* TREND_TOKEN -> This is the secret API key to invoke Trend Micro
* S3_PREFIX -> This is the S3 key that will be used when publishing Trend information to S3


The script expects the -target option to be passed to it.  This can be an IP address or a host.  For example:

**./compliance-agent -target 10.205.48.114 -v 4**

If you are invoking a remote SSH command, you need to pass in a few more arguments.  Pass in the file to save on S3, the username and password to use for SSH access, and the command to run.

./compliance-agent -target 10.205.48.114 -v 4 -username mmerrill -password 888433j3j! -cmd whoami -file whoami
