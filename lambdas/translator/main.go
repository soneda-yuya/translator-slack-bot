package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/slack-go/slack"
	"log"
	"runtime"
	"strings"
	"translator/pkg"
)

type reactionItem struct {
	Type        string `json:"type"`
	Channel     string `json:"channel,omitempty"`
	File        string `json:"file,omitempty"`
	FileComment string `json:"file_comment,omitempty"`
	Timestamp   string `json:"ts,omitempty"`
}

type reactionEvent struct {
	Challenge      string       `json:"challenge"`
	Type           string       `json:"type"`
	User           string       `json:"user"`
	ItemUser       string       `json:"item_user"`
	Item           reactionItem `json:"item"`
	Reaction       string       `json:"reaction"`
	EventTimestamp string       `json:"event_ts"`
}

type ReqBody struct {
	Challenge string        `json:"challenge"`
	Event     reactionEvent `json:"event"`
}

type ResBody struct {
	Challenge string `json:"challenge"`
}

func handleError(err error, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	b, _ := json.Marshal(req)
	log.Println(string(b))
	log.Println(err)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       err.Error(),
	}, nil
}

func Handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("panic recovered: %s", r)
			for depth := 0; ; depth++ {
				_, file, line, ok := runtime.Caller(depth)
				if !ok {
					break
				}
				log.Printf("======> %d: %v:%d", depth, file, line)
			}
			handleError(err, req)
		}
	}()

	if req.Headers["X-Slack-Retry-Num"] != "" {
		return handleError(errors.New("retry"), events.APIGatewayProxyRequest{})
	}

	if req.Body == "" {
		return handleError(errors.New("no body"), req)
	}

	reqBody := &ReqBody{}
	json.Unmarshal([]byte(req.Body), reqBody)
	if reqBody.Event.Type != "reaction_added" {
		return handleError(errors.New("no expected event type"), req)
	}

	translate(reqBody, req)

	resBody := &ResBody{
		Challenge: reqBody.Challenge,
	}
	body, _ := json.Marshal(resBody)

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(body),
	}, nil
}

func translate(reqBody *ReqBody, req events.APIGatewayProxyRequest) {
	slackClient := pkg.NewSlackClient()

	//err := slackClient.VerifySigningSecret(req.MultiValueHeaders, strings.NewReader(req.Body))
	//if err != nil {
	//	handleError(err, req)
	//	return
	//}

	message, err := slackClient.GetMessage(reqBody.Event.Item.Channel, reqBody.Event.Item.Timestamp)
	if err != nil {
		handleError(err, req)
		return
	}

	translatorLangCode, err := pkg.FlagEmojiToLangCode(reqBody.Event.Reaction)
	if err != nil {
		handleError(err, req)
		return
	}
	target := string(translatorLangCode)

	translator := pkg.NewTranslator()

	description := fmt.Sprintf("以下のメッセージを :%s: に翻訳しました", reqBody.Event.Reaction)
	translatedDescription, err := translator.Translate(string(pkg.TranslatorLangCodeEN), description)
	if err != nil {
		handleError(err, req)
		return
	}

	translatedText, err := translator.Translate(target, message.Text)
	if err != nil {
		handleError(err, req)
		return
	}

	options := make([]slack.MsgOption, 0, 10)
	options = append(options, slack.MsgOptionBlocks(
		&slack.SectionBlock{
			Type: slack.MBTSection,
			Text: &slack.TextBlockObject{
				Type:  "plain_text",
				Text:  translatedDescription,
				Emoji: true,
			},
		},

		slack.NewDividerBlock(),

		&slack.SectionBlock{
			Type: slack.MBTSection,
			Text: &slack.TextBlockObject{
				Type:  "plain_text",
				Text:  translatedText,
				Emoji: true,
			},
		},

		&slack.SectionBlock{
			Type: slack.MBTSection,
			Text: &slack.TextBlockObject{
				Type: "mrkdwn",
				Text: fmt.Sprintf("(posted by %s)", fmt.Sprintf("<https://umeboshiio.slack.com/team/%s| this person>", message.User)),
			},
		},

		&slack.ActionBlock{
			Type: slack.MBTAction,
			Elements: &slack.BlockElements{
				ElementSet: []slack.BlockElement{
					&slack.ButtonBlockElement{
						Type: slack.METButton,
						Text: &slack.TextBlockObject{
							Type: "plain_text",
							Text: "Original Message",
						},
						Value: "Link",
						URL:   fmt.Sprintf("https://umeboshiio.slack.com/archives/%s/%s", reqBody.Event.Item.Channel, strings.Replace(message.Timestamp, ".", "", -1)),
					},
				},
			},
		},
	))
	options = append(options, slack.MsgOptionText(translatedText, true))
	postedTimeStamp, err := slackClient.PostMessage(reqBody.Event.Item.Channel, options)
	if err != nil {
		handleError(err, req)
		return
	}

	postMessages := func(threads []slack.Message) error {
		for idx, m := range threads {
			if idx == 0 || m.Subscribed {
				continue
			}
			translatedText, err := translator.Translate(target, m.Text)
			if err != nil {
				return err
			}
			options := []slack.MsgOption{slack.MsgOptionBlocks(
				&slack.SectionBlock{
					Type: slack.MBTSection,
					Text: &slack.TextBlockObject{
						Type:  "plain_text",
						Text:  translatedText,
						Emoji: true,
					},
				},
				&slack.SectionBlock{
					Type: slack.MBTSection,
					Text: &slack.TextBlockObject{
						Type: "mrkdwn",
						Text: fmt.Sprintf("(posted by %s)", fmt.Sprintf("<https://umeboshiio.slack.com/team/%s| this person>", message.User)),
					},
				}),
			}
			options = append(options, slack.MsgOptionTS(postedTimeStamp))
			_, err = slackClient.PostMessage(reqBody.Event.Item.Channel, options)
			if err != nil {
				return err
			}
		}
		return nil
	}

	var threads []slack.Message
	var hasMore bool
	var cursol string
	for {
		threads, hasMore, cursol, err = slackClient.GetThreadMessages(reqBody.Event.Item.Channel, reqBody.Event.Item.Timestamp, cursol)
		postMessages(threads)
		if err != nil {
			handleError(err, req)
			return
		}
		if !hasMore {
			break
		}
	}

	if err != nil {
		handleError(err, req)
		return
	}
}

func main() {
	lambda.Start(Handler)
}
