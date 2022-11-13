package dmmcutil

import "github.com/go-nlp/dmmclust"

type ClusterMeta struct {
	ID          int
	DocIDs      []int
	CenterDocID int
	Cluster     dmmclust.Cluster
}
