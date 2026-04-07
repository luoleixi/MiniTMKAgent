package utils

import (
	"fmt"
	"strings"
)

// 支持的语言代码
var validLangCodes = map[string]string{
	"zh": "中文",
	"en": "英文",
	"ja": "日文",
	"ko": "韩文",
	"fr": "法文",
	"de": "德文",
	"es": "西班牙文",
	"ru": "俄文",
	"it": "意大利文",
	"pt": "葡萄牙文",
	"ar": "阿拉伯文",
}

// ValidateLanguagePair 验证语言对是否有效
func ValidateLanguagePair(src, tgt string) error {
	if !IsValidLangCode(src) {
		return fmt.Errorf("无效的源语言代码: %s", src)
	}
	if !IsValidLangCode(tgt) {
		return fmt.Errorf("无效的目标语言代码: %s", tgt)
	}
	if src == tgt {
		return fmt.Errorf("源语言和目标语言不能相同")
	}
	return nil
}

// IsValidLangCode 检查语言代码是否有效
func IsValidLangCode(code string) bool {
	code = strings.ToLower(code)
	_, ok := validLangCodes[code]
	return ok
}

// LangCodeToName 将语言代码转换为语言名称
func LangCodeToName(code string) string {
	code = strings.ToLower(code)
	if name, ok := validLangCodes[code]; ok {
		return name
	}
	return code
}
