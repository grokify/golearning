package dmmcutil

type ClusterSet struct {
	ClustersMap map[int]Cluster
}

type Cluster struct {
	CenterDocID int
	DocIDs      []int
}

type DocumentSet struct {
	Language   string
	Documents  []Document
	ClusterSet ClusterSet
}

type Document struct {
	ID      int
	Title   string
	Body    string
	Tokens  []string
	Vectors []int
}

// IDs meets `dmmclust.Document` interface requirements.
func (d *Document) IDs() []int {
	return d.Vectors
}

// Len meets `dmmclust.Document` interface requirements.
func (d *Document) Len() int {
	return len(d.Vectors)
}
