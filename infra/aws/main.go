package main

import (
	"fmt"
	"os"

	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

func _(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

func main() {
	defer err2.Catch(err2.Err(func(err error) {
		fmt.Fprintln(os.Stderr, err)
	}))

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

	res := try.To1(client.S3ListBucketFiles(bucketName))

	try.To(client.S3DownloadBucketFiles(bucketName, subfolder, targetFolder, res))

	fmt.Printf("All done, files copied to %s\n", targetFolder)
}
