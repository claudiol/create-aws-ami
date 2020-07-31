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
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"
)

var imageName string
var diskContainerName string
var s3src string
var s3bucket string
var amiName string
var rhcosSrc string
var format string

// importImageCmd represents the importImage command
var importImageCmd = &cobra.Command{
	Use:   "importImage",
	Short: `Cobra application to import disk image to AMI.`,
	Long:  "Import single or multi-volume disk images or EBS snapshots into an Amazon Machine (AMI)",
	Run: func(cmd *cobra.Command, args []string) {
		importImageImpl(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(importImageCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// importImageCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// importImageCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	// importImageCmd.Flags().StringVarP(&imageName, "image", "i", "", "Specify the disk image to convert to AMI")
	// importImageCmd.MarkFlagRequired("image") // Mark as required flag


	importImageCmd.Flags().StringVarP(&s3bucket, "s3bucket", "", "", "AWS S3 Bucket Name e.g. claudiol-bucket")
	importImageCmd.MarkFlagRequired("s3bucket") // Mark as required flag
	importImageCmd.Flags().StringVarP(&s3src, "s3src", "", "", "AWS S3 RHCOS image source e.g. /TISC/Uploads")
	importImageCmd.MarkFlagRequired("s3src") // Mark as required flag	
	importImageCmd.Flags().StringVarP(&rhcosSrc, "rhcosSrc", "", "", "RHCOS image name e.g. rhcos-4.5.2-x86_64-aws.x86_64.vmdk")
	importImageCmd.MarkFlagRequired("rhcosSrc") // Mark as required flag
	importImageCmd.Flags().StringVarP(&format, "format", "", "", "Image format e.g. vmdk")
	importImageCmd.MarkFlagRequired("format") // Mark as required flag
	importImageCmd.Flags().StringVarP(&amiName, "amiName", "", "", "AWS AMI Name e.g. tisc-ami")
	importImageCmd.MarkFlagRequired("amiName") // Mark as required flag
}

func importImageImpl(cmd *cobra.Command, args []string) bool {

	// Structure for UserBucket to be passed in
	// All the TISC uploads will go into the /TISC/Uploads key
	// TODO: This will need to be variabilized

	source := s3src + "/" + rhcosSrc
	fmt.Println(source)
	userBucket := &ec2.UserBucket{
		S3Bucket: aws.String(s3bucket),
		S3Key:    aws.String(source),
	}

	imageSnapContainer := &ec2.SnapshotDiskContainer{
		Description: aws.String("Red Hat Provided RHCOS Image"),
		Format:      aws.String(format),
	}

	//fmt.Println(s3src)
	//imageSnapContainer.SetUrl(s3src) // This is the URL passed in from the command line flag -u <URL>

	var buckets []*ec2.UserBucket
	buckets = append(buckets, userBucket)

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(awsCfg.awsRegion)},
	)

	// Create EC2 service client
	svc := ec2.New(sess)

	// Create the ImportSnapshotInput object
	importSnapshotInput := &ec2.ImportSnapshotInput{
		Description: aws.String("Red Hat provided RHCOS image"),
	}

	// Add the UserBucket
	imageSnapContainer.SetUserBucket(userBucket)

	// Add the ImageSnapshotDiskContainer to the list in ImportImageInput
	importSnapshotInput.SetDiskContainer(imageSnapContainer)

	fmt.Printf("Calling import Snapshot ...")
	result, err := svc.ImportSnapshot(importSnapshotInput) 

	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	fmt.Printf("Import Task id: %v\n", *result.ImportTaskId)
	fmt.Println(result)

	// Let's call DescribeImportSnapshotTasks
	var taskIds []*string
	taskIds = append(taskIds, result.ImportTaskId)
	snapshotTasksInput := &ec2.DescribeImportSnapshotTasksInput{}
	snapshotTasksInput.SetImportTaskIds(taskIds)

	// Wait until we get the snapshot id and the task is complete
	var snapshotID string

	// Create the Progress bar ...
	bar := &Progbar{total: 100}

	// for is Go's while
	for i := 0; i < 1000; {
		taskOutput, err := svc.DescribeImportSnapshotTasks(snapshotTasksInput)
		if err != nil {
			fmt.Println(err.Error())
		}
		if *taskOutput.ImportSnapshotTasks[0].SnapshotTaskDetail.Status != "completed" {
			if taskOutput.ImportSnapshotTasks[0].SnapshotTaskDetail.Progress != nil {
			  perc, _ := strconv.Atoi(*taskOutput.ImportSnapshotTasks[0].SnapshotTaskDetail.Progress)
			  bar.PrintProg(perc)
			  time.Sleep(2 * time.Second)
                        } else {
  			  fmt.Println(taskOutput.ImportSnapshotTasks)
                          i=1000
                        }
		} else {
			snapshotID = *taskOutput.ImportSnapshotTasks[0].SnapshotTaskDetail.SnapshotId
			bar.PrintComplete()
			fmt.Println(*(taskOutput.ImportSnapshotTasks[0].SnapshotTaskDetail.SnapshotId))
			i = 1000
		}
	}

	//
	// We no Register the newly created
	fmt.Println("Registering Image ...")

	// We need an EbsBlockDevice object
	ebsBlockDevice := &ec2.EbsBlockDevice{}
	ebsBlockDevice.SetSnapshotId(snapshotID)
	ebsBlockDevice.SetDeleteOnTermination(true)

	// We need a BlockDeviceMapping object
	blockDeviceMapping := &ec2.BlockDeviceMapping{}
	blockDeviceMapping.SetDeviceName("/dev/sda1")
	blockDeviceMapping.SetEbs(ebsBlockDevice)

	// We need a Slice of BlockDeviceMappings
	var blockDeviceMappings []*ec2.BlockDeviceMapping
	blockDeviceMappings = append(blockDeviceMappings, blockDeviceMapping)

	// Create a RegisterImageInput object
	registerImageInput := &ec2.RegisterImageInput{
		Name:               aws.String(amiName),
		Architecture:       aws.String("x86_64"),
		Description:        aws.String("Red Hat Provided RHCOS Image"),
		RootDeviceName:     aws.String("/dev/sda1"),
		EnaSupport:         aws.Bool(true),
		VirtualizationType: aws.String("hvm"),
	}
	registerImageInput.SetBlockDeviceMappings(blockDeviceMappings)

	// Make the call to register the AMI with the snapshot created
	registerResult, registerErr := svc.RegisterImage(registerImageInput)
	if registerErr != nil {
		fmt.Println(registerErr.Error())
		return false
	}
	fmt.Printf("AMI ID is: %v\n", *registerResult.ImageId)
	return true
}
