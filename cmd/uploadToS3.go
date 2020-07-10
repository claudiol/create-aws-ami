/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	// AWS Go SDK
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var bucketName string
var fileName string
var createBucket bool

// uploadToS3Cmd represents the uploadToS3 command
var uploadToS3Cmd = &cobra.Command{
	Use:   "uploadToS3",
	Short: "Uploads a file to a specified S3 bucket.",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Print: " + strings.Join(args, " "))
		uploadToS3Impl(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(uploadToS3Cmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// uploadToS3Cmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	uploadToS3Cmd.Flags().StringVarP(&bucketName, "bucket", "b", "", "Specify the S3 bucket name to upload to")
	uploadToS3Cmd.MarkFlagRequired("bucket") // Mark as required flag

	// Add a boolean flag to create the S3 bucket if it does not exist
	uploadToS3Cmd.Flags().BoolVarP(&createBucket, "create", "c", false, "Create S3 bucket if it does not exist")

	uploadToS3Cmd.Flags().StringVarP(&fileName, "file", "f", "", "File to upload to S3 bucket")
	uploadToS3Cmd.MarkFlagRequired("file") // Mark as required flag
}

func uploadToS3Impl(cmd *cobra.Command, args []string) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(awsCfg.awsRegion)},
	)

	bucketInput := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	}

	headBucketInput := &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	}

	fExists := false

	// Create S3 service client
	svc := s3.New(sess)

	fmt.Printf("Checking to see if S3 bucket [%v] exists ...", bucketName)
	result, err := svc.HeadBucket(headBucketInput)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				fmt.Println(s3.ErrCodeNoSuchBucket, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
	} else {
		fExists = true
		fmt.Println(fExists, result)
		bucketLocationInput := &s3.GetBucketLocationInput{
			Bucket: aws.String(bucketName),
		}

		locationresult, _ := svc.GetBucketLocation(bucketLocationInput)

		fmt.Printf("S3 Bucket location: %v\n", locationresult)
	}

	if !fExists && createBucket {
		fmt.Printf("Creating bucket [%v]: ", bucketName)
		result, err := svc.CreateBucket(bucketInput)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case s3.ErrCodeBucketAlreadyExists:
					fmt.Println(s3.ErrCodeBucketAlreadyExists, aerr.Error())
				case s3.ErrCodeBucketAlreadyOwnedByYou:
					fmt.Println(s3.ErrCodeBucketAlreadyOwnedByYou, aerr.Error())
				default:
					fmt.Println(aerr.Error())
				}
			} else {
				// Print the error, cast err to awserr.Error to get the Code and
				// Message from an error.
				fmt.Println(err.Error())
			}
		}
		fmt.Printf("done: %v\n", result)
		bucketLocationInput := &s3.GetBucketLocationInput{
			Bucket: aws.String(bucketName),
		}

		locationresult, _ := svc.GetBucketLocation(bucketLocationInput)

		fmt.Printf("S3 Bucket location: %v\n", locationresult)
	}

	// Create an uploader with the session and default options
	uploader := s3manager.NewUploader(sess)

	f, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("failed to open file %q, %v", fileName, err)
	}

	// Upload the file to S3.
	uploadResult, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String("//TISC//Uploads//" + filepath.Base(fileName)),
		Body:   f,
	})
	if err != nil {
		fmt.Printf("failed to upload file, %v", err)
	}
	fmt.Printf("file uploaded to, %v\n", uploadResult)
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
