package pkg

import (
	"errors"
	"strings"
)

type EmojiFlag string
type TranslatorLangCode string

const (
	EmojiFlagUnknown = EmojiFlag("unknown")
	EmojiFlagJP      = EmojiFlag("jp")
	EmojiFlagUS      = EmojiFlag("us")
	EmojiFlagEN      = EmojiFlag("england")
	EmojiFlagVN      = EmojiFlag("vn")
)

const (
	TranslatorLangCodeUnknown = TranslatorLangCode("unknown")
	TranslatorLangCodeJP      = TranslatorLangCode("ja")
	TranslatorLangCodeEN      = TranslatorLangCode("en")
	TranslatorLangCodeVN      = TranslatorLangCode("vi")
)

var EmojiFlagList = []EmojiFlag{EmojiFlagJP, EmojiFlagUS, EmojiFlagEN, EmojiFlagVN}

func FlagEmojiToLangCode(flagEmoji string) (TranslatorLangCode, error) {
	separated := strings.Split(flagEmoji, "-")
	emojiFlag := flagEmoji
	if len(separated) == 2 {
		emojiFlag = separated[1]
	}

	if !SliceContains(EmojiFlagList, emojiFlag) {
		return TranslatorLangCodeUnknown, errors.New("unknown emoji")
	}

	translatorLangCode := TranslatorLangCodeEN
	switch EmojiFlag(emojiFlag) {
	case EmojiFlagJP:
		translatorLangCode = TranslatorLangCodeJP
	case EmojiFlagVN:
		translatorLangCode = TranslatorLangCodeVN
	}

	return translatorLangCode, nil
}
