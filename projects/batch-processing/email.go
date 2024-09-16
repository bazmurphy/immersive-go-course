package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

// THIS WAS AN ADDITIONAL TASK SET BY DANIEL
func SendEmailWithSES(recipient string, generatedDataMap GeneratedDataMap) error {
	log.Println("🔵 attempting: to email a report using AWS SES...")

	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		return fmt.Errorf("🔴 error: cannot get the AWS_REGION environment variable")
	}

	awsAccessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	if awsAccessKeyID == "" {
		return fmt.Errorf("🔴 error: cannot get the AWS_ACCESS_KEY_ID environment variable")
	}

	awsSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if awsSecretAccessKey == "" {
		return fmt.Errorf("🔴 error: cannot get the AWS_SECRET_ACCESS_KEY environment variable")
	}

	awsSession := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region:      aws.String(awsRegion),
			Credentials: credentials.NewStaticCredentials(awsAccessKeyID, awsSecretAccessKey, ""),
		},
	}))

	awsSESClient := ses.New(awsSession)

	emailCharset := "UTF-8"
	emailSubject := "Batch Processing Email Report"
	emailSender := "bazmurphy@gmail.com"
	emailHTMLBody := constructEmailHTMLBody(generatedDataMap)
	emailTextBody := "Batch Processing - Report - No Text Body (Need HTML Body)"

	customEmailParameters := &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: []*string{
				aws.String(recipient),
			},
		},
		Message: &ses.Message{
			Subject: &ses.Content{
				Charset: aws.String(emailCharset),
				Data:    aws.String(emailSubject),
			},
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(emailCharset),
					Data:    aws.String(emailHTMLBody),
				},
				Text: &ses.Content{
					Charset: aws.String(emailCharset),
					Data:    aws.String(emailTextBody),
				},
			},
		},
		Source: aws.String(emailSender),
	}

	_, err := awsSESClient.SendEmail(customEmailParameters)
	if err != nil {
		// https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/ses-example-send-email.html
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				fmt.Println(ses.ErrCodeMessageRejected, aerr.Error())
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				fmt.Println(ses.ErrCodeMailFromDomainNotVerifiedException, aerr.Error())
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				fmt.Println(ses.ErrCodeConfigurationSetDoesNotExistException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and Message from an error
			fmt.Println(err.Error())
		}
		return fmt.Errorf("🔴 error: failed to send the email report using AWS SES: %v", err)
	}

	log.Printf("🟢 success: emailed a report to %s using AWS SES", recipient)

	return nil
}

func constructEmailHTMLBody(generatedDataMap GeneratedDataMap) string {

	var emailHTMLTableRows string

	for _, data := range generatedDataMap {
		emailHTMLTableRows += fmt.Sprintf(`
			<tr>
					<td><a href="%s">%s</a></td>
					<td>%s</td>
					<td>%s</td>
					<td><a href="%s">%s</a></td>
			</tr>
		`, data[0], data[0], data[1], data[2], data[3], data[3])
	}

	emailHTMLTable := fmt.Sprintf(`
		<table>
				<thead>
						<tr>
								<th>Original URL</th>
								<th>Downloaded Filepath</th>
								<th>Grayscale Filepath</th>
								<th>S3 URL</th>
						</tr>
				</thead>
				<tbody>
						%s
				</tbody>
		</table>
    `, emailHTMLTableRows)

	emailHTMLBody := fmt.Sprintf(`
		<h1>Batch Processing</h1>
		<h2>Report</h2>
		%s
	`, emailHTMLTable)

	return emailHTMLBody
}
