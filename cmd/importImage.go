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
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"
)

var imageName string
var diskContainerName string
var url string

// importImageCmd represents the importImage command
var importImageCmd = &cobra.Command{
	Use:   "importImage",
	Short: `Cobra application to import disk image to AMI.`,
	Long:  "Import single or multi-volume disk images or EBS snapshots into an Amazon Machine (AMI)",
	Run: func(cmd *cobra.Command, args []string) {
		importImageImpl(cmd, args)
		//fmt.Println("importImage called")
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

	// Flag to pass in the disk-container json file
	//- Create a json file with your bucket name and the name of the image:

	// [
	//   {
	// 	  "Description": "STIG RHEL 7.2 Image",
	// 	  "Format": "vhd",
	// 	  "UserBucket": {
	// 		   "S3Bucket": "claudiol-bucket",
	// 		   "S3Key": "rhel-7.2.vhd"
	// 	   }
	//   }
	// ]
	//  7 - Now run the import command:
	// aws ec2 import-image --description "RHEL 7.2 Image" --disk-containers file://import-rhel.json
	importImageCmd.Flags().StringVarP(&diskContainerName, "", "d", "", "JSON file with Disk container description for AWS image")
	importImageCmd.MarkFlagRequired("d") // Mark as required flag

	importImageCmd.Flags().StringVarP(&url, "url", "u", "", "AWS image URL")
	importImageCmd.MarkFlagRequired("u") // Mark as required flag

}

func importImageImpl(cmd *cobra.Command, args []string) bool {

	// Structure for UserBucket to be passed in
	// All the TISC uploads will go into the /TISC/Uploads key
	// TODO: This will need to be variabilized
	userBucket := &ec2.UserBucket{
		S3Bucket: aws.String("claudiol-bucket"),
		S3Key:    aws.String("//TISC/Uploads//rhcos-4.4.3-x86_64-aws.x86_64.vmdk"),
	}

	imageSnapContainer := &ec2.SnapshotDiskContainer{
		Description: aws.String("Red Hat Provided RHCOS Image"),
		Format:      aws.String("VMDK"),
	}

	imageSnapContainer.SetUrl("https://claudiol-bucket.s3-us-gov-west-1.amazonaws.com/TISC/Uploads/rhcos-4.4.3-x86_64-aws.x86_64.vmdk")

	var buckets []*ec2.UserBucket
	buckets = append(buckets, userBucket)

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(awsCfg.awsRegion)},
	)

	// Create EC2 service client
	svc := ec2.New(sess)

	// Create the ImportSnapshotInput object
	importSnapshotInput := &ec2.ImportSnapshotInput{
		ClientToken: aws.String("tisc-rhcos-4.4.3-x86_64-snapshot"),
		Description: aws.String("Red Hat provided RHCOS image"),
	}

	// Add the ImageSnapshotDiskContainer to the list in ImportImageInput
	importSnapshotInput.SetDiskContainer(imageSnapContainer)

	fmt.Printf("Calling import Snapshot ...")
	result, err := svc.ImportSnapshot(importSnapshotInput) //Image(importImageInput)

	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	fmt.Printf("Import Task id\n: %v", *result.ImportTaskId)
	fmt.Println(result)

	//var snapshotIds []*string // Slice of Snapshot Ids
	//snapshotIds = append(snapshotIds, result.SnapshotTaskDetail.SnapshotId)

	//fmt.Println(snapshotIds)
	//describeSnapshotInput := &ec2.DescribeSnapshotsInput{}

	//describeSnapshotInput.SetSnapshotIds(snapshotIds)

	// err = svc.WaitUntilSnapshotCompleted(describeSnapshotInput)

	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return false
	// }
	// fmt.Println("creation complete!")

	// Let's call DescribeImportSnapshotTasks
	var taskIds []*string
	taskIds = append(taskIds, result.ImportTaskId)
	snapshotTasksInput := &ec2.DescribeImportSnapshotTasksInput{}
	snapshotTasksInput.SetImportTaskIds(taskIds)

	var snapshotID string
	var bar string

	for i := 0; i < 1000; {
		taskOutput, err := svc.DescribeImportSnapshotTasks(snapshotTasksInput)
		if err != nil {
			fmt.Println(err.Error())
		}
		bar = "="
		if *taskOutput.ImportSnapshotTasks[0].SnapshotTaskDetail.Status != "completed" {
			fmt.Printf("%v=%v%v", bar, *taskOutput.ImportSnapshotTasks[0].SnapshotTaskDetail.Progress, "%")
			// fmt.Printf("Description: %v \t Complete Percentage: %v\t Status: %v \n",
			// 	taskOutput.ImportSnapshotTasks[0].SnapshotTaskDetail.Description,
			// 	taskOutput.ImportSnapshotTasks[0].SnapshotTaskDetail.Progress,
			// 	taskOutput.ImportSnapshotTasks[0].SnapshotTaskDetail.Status)
			time.Sleep(10 * time.Second)
		} else {
			snapshotID = *taskOutput.ImportSnapshotTasks[0].SnapshotTaskDetail.SnapshotId
			fmt.Println("==100%")
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
		Name:               aws.String("tisc-rhcos-4.4.3-x86_64-ami"),
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
