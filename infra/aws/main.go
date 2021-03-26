package main

import (
	"fmt"
	"os"

	"github.com/lainio/err2"
)

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s bucketName subfolder targetFolder\n", os.Args[0])
		fmt.Print("Example (copy all bucket contents to current folder):")
		fmt.Printf(" %s website-bucket \"\" .\n", os.Args[0])
		os.Exit(1)
	}

	bucketName := os.Args[1]
	subfolder := os.Args[2]
	targetFolder := os.Args[3]

	client := NewS3Client()

	// TODO: on error we just panic now
	res, err := client.S3ListBucketFiles(bucketName)
	err2.Check(err)

	err = client.S3DownloadBucketFiles(bucketName, subfolder, targetFolder, res)
	err2.Check(err)

	fmt.Printf("All done, files copied to %s\n", targetFolder)
}
