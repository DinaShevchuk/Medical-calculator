package ds

type DrugListSerializer struct {
    Drug
    IsMine bool `json:"is_mine"`
}

type DrugFeedSerializer struct {
    Drug
    IsLiked bool `json:"is_liked"`
    LikeCount int64 `json:"like_count"`
}

type DrugFullSerializer struct {
    Drug
    Creator User `json:"creator"`
}
