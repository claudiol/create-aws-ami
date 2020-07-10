# create-aws-ami

Utility to create AMI on AWS from an image

````
Usage:
  create-aws-ami [command]

Available Commands:
  help        Help about any command
  uploadToS3  Uploads a file to a specified S3 bucket.

Flags:
      --config string   config file (default is $HOME/.create-aws-ami.yaml)
  -h, --help            help for create-aws-ami

Use "create-aws-ami [command] --help" for more information about a command.
````
# Using Configuration file or Environment variables

For now we are using a file called *.create-aws-ami* as our configuration file. This file will need to exist in your $HOME directory. The file format is:

````
[AWSGovCloud]
aws_access_key_id: <REDACTED>	
aws_secret_access_key: <REDACTED>
aws_default_region: "us-gov-west-1"
````

This file will probably change in the future but you can also use environment variables by going the following:
export AWS_ACCESS_KEY_ID=<REDACTED>	
export AWS_SECRET_ACCESS_KEY=<REDACTED>
export AWS_DEFAULT_REGION="us-gov-west-1"

Once these variables are set you can run the command: 

````
$ create-aws-ami uploadToS3 -b claudiol-bucket -f ./main.go -c
````

# How to build

Set your GOPATH variable to point where you have your Go source files. You can find out the current location but running the **go env** command on the command line.

Or you can just set it in your environment:

````
$ export GOPATH="/home/claudiol/go"
````

## Clone the project

````
$ cd $GOPATH/src
$ git clone https://github.com/claudiol/create-aws-ami.git
$ cd create-aws-ami
$ git install create-aws-ami
````
# Current branch 

claudiol-upload-s3 - Has the latest code for uploading to S3 bucket.

