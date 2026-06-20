package domain

type ResourceKind string

const (
	ResourceKindContainer ResourceKind = "container"
	ResourceKindImage     ResourceKind = "image"
	ResourceKindVolume    ResourceKind = "volume"
	ResourceKindNetwork   ResourceKind = "network"
)

type ConnectionStatus struct {
	OK      bool
	Message string
}

type ResourceSummary struct {
	Kind    ResourceKind
	ID      string
	Name    string
	Raw     map[string]any
	Columns map[string]string
	Group   string
	IsGroup bool
}

func (r ResourceSummary) ShortID() string {
	if len(r.ID) <= 12 {
		return r.ID
	}

	return r.ID[:12]
}
