package ds

// Список — краткая информация + флаг «создано текущим пользователем»
type DrugListSerializer struct {
    Drug
    IsMine bool `json:"is_mine"`
}

// Лента — краткая информация + флаг «лайкнул ли текущий пользователь»
type DrugFeedSerializer struct {
    Drug
    IsLiked bool `json:"is_liked"`
    LikeCount int64 `json:"like_count"`
}

// Подробная информация — с вложенным создателем
type DrugFullSerializer struct {
    Drug
    Creator User `json:"creator"`
}
