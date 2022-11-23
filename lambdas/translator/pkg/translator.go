package pkg

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/translate"
)

type translator struct {
	client *translate.Translate
}

type Translator interface {
	Translate(to, message string) (string, error)
}

func NewTranslator() Translator {
	sess := session.Must(session.NewSession())
	return &translator{
		client: translate.New(sess),
	}
}

func (t *translator) Translate(to, message string) (string, error) {
	from := aws.String(string(TranslatorLangCodeJP))
	if to == string(TranslatorLangCodeJP) {
		from = aws.String(string(TranslatorLangCodeVN))
	}

	result, err := t.client.Text(&translate.TextInput{
		SourceLanguageCode: from,
		TargetLanguageCode: aws.String(to),
		Text:               aws.String(message),
	})

	text := message
	if result.TranslatedText != nil {
		text = *result.TranslatedText
	}

	return text, err
}
