package flow

type serviceDependency struct {
	Id              string `json:"id"`
	InteractionType string `json:"interaction_type"`
}

type PathItem struct {
	Current string
	Next    []string
}

type FlowPath struct {
	FlowID int        `json:"flow_id"`
	Path   []PathItem `json:"path"`
}
