package translator

// Translator 翻译器接口
type Translator interface {
	// Translate 翻译文本
	// sourceLang: 源语言代码
	// targetLang: 目标语言代码
	// text: 要翻译的文本
	Translate(sourceLang, targetLang, text string) (string, error)
}
