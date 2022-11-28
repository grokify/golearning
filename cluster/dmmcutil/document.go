package dmmcutil

import (
	"github.com/go-nlp/dmmclust"
	"github.com/grokify/mogo/type/maputil"
)

type Documents []dmmclust.Document

func (docs Documents) IDs() []int {
	wordIDs := map[int]int{}
	for _, doc := range docs {
		for _, id := range doc.IDs() {
			wordIDs[id]++
		}
	}
	return maputil.IntKeys(wordIDs)
}
