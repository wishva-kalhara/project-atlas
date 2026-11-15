package data

// Structs matching the JSON contract in `.data-workflow-fb-schedule-in/metadata.json`

type SchedulingInfo struct {
    IsSuccess bool `json:"isSuccess"`
    Tried     bool `json:"tried"`
}

type Post struct {
    Caption        string         `json:"caption"`
    VideoPath      string         `json:"videoPath"`
    SchedulingInfo SchedulingInfo `json:"schdulingInfo"`
}

type FbReelsData struct {
    FbPageSlug string `json:"fbPageSlug"`
    PageName   string `json:"pageName"`
    Posts      []Post `json:"posts"`
}
