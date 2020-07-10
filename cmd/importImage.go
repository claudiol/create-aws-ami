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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"
)

var imageName string
var diskContainerName string

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
	importImageCmd.MarkFlagRequired("dc") // Mark as required flag

}

func importImageImpl(cmd *cobra.Command, args []string) {

	userBucket := &ec2.UserBucket{
		S3Bucket: aws.String("claudiol-bucket"),
		S3Key:    aws.String("//TISC/Uploads//rhcos-4.4.3-x86_64-aws.x86_64.vmdk"),
	}

	imageDiskContainer := &ec2.ImageDiskContainer{
		Description: aws.String("Red Hat Provided RHCOS Image"),
		Format:      aws.String("VMDK"),
	}

	var buckets []*ec2.UserBucket
	buckets = append(buckets, userBucket)

	var containers []*ec2.ImageDiskContainer
	containers = append(containers, imageDiskContainer)

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(awsCfg.awsRegion)},
	)

	// Create S3 service client
	svc := ec2.New(sess)

	// Create the ImportImageInput object
	importImageInput := &ec2.ImportImageInput{
		Architecture: aws.String("x86_64"),
		Platform:     aws.String("Linux"),
		Description:  aws.String("Red Hat provided RHCOS image"),
	}

	// Ad the ImageDiskContainer to the list in ImportImageInput
	importImageInput.SetDiskContainers(containers)

	result, err := svc.ImportImage(importImageInput)

	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(result)

}
