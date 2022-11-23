package pkg

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/slack-go/slack"
	"io"
	"log"
	"net/http"
)

type appSlack struct {
	client        *slack.Client
	SigningSecret string
}

type AppSlackClient interface {
	VerifySigningSecret(header http.Header, reader io.Reader) error
	GetMessage(channelName, oldestTimestamp string) (slack.Message, error)
	GetThreadMessages(channelName, threadTimestamp, cursor string) (msgs []slack.Message, hasMore bool, nextCursor string, err error)
	PostMessage(channelName string, msgOptions []slack.MsgOption) (timeStamp string, err error)
}

func NewSlackClient() AppSlackClient {
	tokens := getSlackAppSecrets()
	api := slack.New(tokens.OAuthToken)
	return &appSlack{
		client:        api,
		SigningSecret: tokens.SigningSecret,
	}
}

func (s *appSlack) VerifySigningSecret(header http.Header, body io.Reader) error {
	verifier, err := slack.NewSecretsVerifier(header, s.SigningSecret)
	if err != nil {
		return err
	}

	bodyReader := io.TeeReader(body, &verifier)
	_, err = io.ReadAll(bodyReader)
	if err != nil {
		return err

	}

	err = verifier.Ensure()
	if err != nil {
		return err
	}

	return nil
}

func (s *appSlack) GetMessage(channelName, oldestTimestamp string) (slack.Message, error) {
	mainMessageRes, err := s.client.GetConversationHistory(
		&slack.GetConversationHistoryParameters{
			Limit:              1,
			ChannelID:          channelName,
			Inclusive:          true,
			Oldest:             oldestTimestamp,
			IncludeAllMetadata: true,
		},
	)
	if err != nil {
		return slack.Message{}, err
	}

	if len(mainMessageRes.Messages) < 0 {
		return slack.Message{}, errors.New("not found message")
	}

	return mainMessageRes.Messages[0], nil
}

func (s *appSlack) GetThreadMessages(channelName, threadTimestamp, cursor string) (msgs []slack.Message, hasMore bool, nextCursor string, err error) {
	return s.client.GetConversationReplies(
		&slack.GetConversationRepliesParameters{
			ChannelID: channelName,
			Cursor:    cursor,
			Inclusive: true,
			Timestamp: threadTimestamp,
		},
	)
}

func (s *appSlack) PostMessage(channelName string, msgOptions []slack.MsgOption) (string, error) {
	_, timeStamp, err := s.client.PostMessage(
		channelName,
		msgOptions...,
	)
	return timeStamp, err
}

type slackAppToken struct {
	OAuthToken    string `json:"app_translator_oauth_token"`
	SigningSecret string `json:"app_translator_signing_secret"`
}

func getSlackAppSecrets() *slackAppToken {
	secretName := "slack-app-token"
	region := "ap-northeast-1"

	config, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Fatal(err)
	}

	svc := secretsmanager.NewFromConfig(config)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"),
	}

	result, err := svc.GetSecretValue(context.TODO(), input)
	if err != nil {
		log.Fatal(err.Error())
	}

	tokens := &slackAppToken{}
	json.Unmarshal([]byte(*result.SecretString), tokens)

	return tokens
}
