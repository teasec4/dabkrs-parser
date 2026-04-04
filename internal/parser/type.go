package parser

type Entry struct {
	Headword         string // 三比西河
	Pinyin           string // sānbǐxīhé
	PinyinNormalized string // sanbixihe
	Meanings         []Meaning
}

type Meaning struct {
	Level int // m1, m2, m3
	Text  string
	Tags  []Tag
	Order int // порядок значения
}

type Tag struct {
	Type  string // i, p, ref, ex
	Value string
}
