package ds

type DrugListSerializer struct {
    Dosage_drugs
    IsMine bool `json:"is_mine"`
}

type DrugFeedSerializer struct {
    Dosage_drugs
    IsLiked bool `json:"is_liked"`
    LikeCount int64 `json:"like_count"`
}

type DrugFullSerializer struct {
    Dosage_drugs
    Creator User `json:"creator"`
}

type FeedResponse struct {
    Data       []DrugFeedSerializer `json:"data"`
    TotalCount int64                 `json:"total_count"`
    Limit      int                   `json:"limit"`
    Offset     int                   `json:"offset"`
}
