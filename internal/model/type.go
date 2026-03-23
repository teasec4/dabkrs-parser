package model

type Entry struct{
	Headword string // 中国
	Pinyin string // zhongguo
	Meanings []Meaning
}

type Meaning struct{
	Level int // m1 m2 m3 
	Text string
	Tags []Tag
}

type Tag struct{
	Type string // i, p, ref, ex
	Value string
}