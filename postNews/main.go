package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

var db dynamodbiface.DynamoDBAPI

func init() {
	sess := session.Must(session.NewSession())
	db = dynamodb.New(sess, &aws.Config{
		Region: aws.String("us-east-1"), // Adjust region if needed
	})
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("Received request: %v", req.Body)

	var body map[string]string
	err := json.Unmarshal([]byte(req.Body), &body)
	if err != nil {
		log.Println("Error unmarshaling request body:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Invalid request format",
			Headers: map[string]string{
				"Access-Control-Allow-Origin": "*",
				"Access-Control-Allow-Methods": "POST",
				"Access-Control-Allow-Headers": "Content-Type",
			},
		}, err
	}

	newsText, ok := body["news"]
	if !ok {
		err := fmt.Errorf("missing 'news' field")
		log.Println(err)
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Missing 'news' field",
			Headers: map[string]string{
				"Access-Control-Allow-Origin": "*",
				"Access-Control-Allow-Methods": "POST,OPTIONS,GET",
				"Access-Control-Allow-Headers": "Content-Type",
			},
		}, err
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String("NewsTable"), // Adjust table name if needed
		Item: map[string]*dynamodb.AttributeValue{
			"news_id": {
				S: aws.String(fmt.Sprintf("%d", time.Now().UnixNano())),
			},
			"timestamp": {
				S: aws.String(fmt.Sprintf("%d", time.Now().UnixNano())),
			},
			"text": {
				S: aws.String(newsText),
			},
		},
	}

	_, err = db.PutItem(input)
	if err != nil {
		log.Println("Error adding item to DynamoDB:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "Failed to post news",
			Headers: map[string]string{
				"Access-Control-Allow-Origin": "*",
				"Access-Control-Allow-Methods": "POST,OPTIONS,GET",
				"Access-Control-Allow-Headers": "Content-Type",
			},
		}, err
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "News posted successfully",
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
			"Access-Control-Allow-Methods": "POST,OPTIONS,GET",
			"Access-Control-Allow-Headers": "Content-Type",
		},
	}, nil
}


func main() {
	lambda.Start(Handler)
}