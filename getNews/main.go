package main

import (
	"context"
	"encoding/json"
	"log"
	"sort"

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
    input := &dynamodb.ScanInput{
        TableName: aws.String("NewsTable"), // Adjust table name if needed
    }

    result, err := db.Scan(input)
    if err != nil {
        log.Println("Error scanning DynamoDB:", err)
        return events.APIGatewayProxyResponse{
            StatusCode: 500,
            Headers: map[string]string{
                "Access-Control-Allow-Origin":  "*",
                "Access-Control-Allow-Methods": "GET,POST,OPTIONS",
                "Access-Control-Allow-Headers": "Content-Type",
            },
            Body: "Error scanning DynamoDB: " + err.Error(),
        }, err
    }

    // Handle empty results gracefully
    if len(result.Items) == 0 {
        return events.APIGatewayProxyResponse{
            StatusCode: 200,
            Headers: map[string]string{
                "Access-Control-Allow-Origin":  "*",
                "Access-Control-Allow-Methods": "GET,POST,OPTIONS",
                "Access-Control-Allow-Headers": "Content-Type",
            },
            Body: "[]",
        }, nil
    }

    newsItems := make([]map[string]*dynamodb.AttributeValue, len(result.Items))
    copy(newsItems, result.Items)

    // Sort by news_id (reverse chronological order), checking for nil values
    sort.Slice(newsItems, func(i, j int) bool {
        idI, idJ := newsItems[i]["news_id"], newsItems[j]["news_id"]
        if idI == nil || idI.S == nil {
            return false
        }
        if idJ == nil || idJ.S == nil {
            return true
        }
        return *idI.S > *idJ.S
    })

    body, err := json.Marshal(newsItems)
    if err != nil {
        log.Println("Error marshaling news items:", err)
        return events.APIGatewayProxyResponse{
            StatusCode: 500,
            Headers: map[string]string{
                "Access-Control-Allow-Origin":  "*",
                "Access-Control-Allow-Methods": "GET,POST,OPTIONS",
                "Access-Control-Allow-Headers": "Content-Type",
            },
            Body: "Error marshaling news items: " + err.Error(),
        }, err
    }

    return events.APIGatewayProxyResponse{
        StatusCode: 200,
        Headers: map[string]string{
            "Access-Control-Allow-Origin":  "*",
            "Access-Control-Allow-Methods": "GET,POST,OPTIONS",
            "Access-Control-Allow-Headers": "Content-Type",
        },
        Body: string(body),
    }, nil
}

func main() {
    lambda.Start(Handler)
}
