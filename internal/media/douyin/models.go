package douyin

type CallRequestParams struct {
	NeedSign bool
	Headers  map[string]string
}

type UserInfo struct {
	UID      string `json:"uid"`
	Nickname string `json:"nickname"`
}

// VerifyParams 结构体
type VerifyParams struct {
	MsToken  string `json:"ms_token"`
	WebID    string `json:"webid"`
	VerifyFp string `json:"verify_fp"`
	SVWebID  string `json:"s_v_web_id"`
}

// Cookie 结构体
type Cookie struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// 签名请求
type SignRequest struct {
	URI         string
	UserAgent   string
	Cookies     string
	QueryParams string
}

type FetchResult struct {
	Body []byte
}

// Resp 定义响应结构体
type PongResp struct {
	ID                 string `json:"id"`
	CreateTime         string `json:"create_time"`
	LastTime           string `json:"last_time"`
	UserUID            string `json:"user_uid"`
	UserUIDType        int    `json:"user_uid_type"`
	FirebaseInstanceID string `json:"firebase_instance_id"`
	UserAgent          string `json:"user_agent"`
	BrowserName        string `json:"browser_name"`
}

type SearchParams struct {
	SearchId        string
	SearchChannel   SearchChannelType
	PublishTimeType PublishTimeType
	SortType        SearchSortType
	Keyword         string
	Offset          int
	Count           int
}

// 定义作者信息结构体
type Author struct {
	UID                    string `json:"uid"`
	ShortId                string `json:"short_id"`
	UniqueId               string `json:"unique_id"`
	Nickname               string `json:"nickname"`
	AvatarThumb            Avatar `json:"avatar_thumb"`
	Signature              string `json:"signature"`
	EnterpriseVerifyReason string `json:"enterprise_verify_reason"`
	SecUID                 string `json:"sec_uid"`
}

// 定义视频信息结构体
type Video struct {
	PlayAddr VideoURL `json:"play_addr"`
	Cover    Avatar   `json:"cover"`
	Duration int      `json:"duration"`
}

// 定义视频 URL 结构体
type VideoURL struct {
	URLList []string `json:"url_list"`
}

// 定义统计信息结构体
type Statistics struct {
	CommentCount int64 `json:"comment_count"`
	DiggCount    int64 `json:"digg_count"`
	ShareCount   int64 `json:"share_count"`
	CollectCount int64 `json:"collect_count"`
}

// 定义 Aweme 结构体
type Aweme struct {
	AwemeID    string     `json:"aweme_id"`
	AwemeType  int        `json:"aweme_type"`
	Desc       string     `json:"desc"`
	CreateTime int64      `json:"create_time"`
	Author     Author     `json:"author"`
	Video      Video      `json:"video"`
	Statistics Statistics `json:"statistics"`
}

// 定义主 JSON 结构体
type SearchResponse struct {
	StatusCode int `json:"status_code"`
	Data       []struct {
		AwemeInfo Aweme `json:"aweme_info"`
	} `json:"data"`
	Extra struct {
		Now        int64    `json:"now"`
		Logid      string   `json:"logid"`
		FatalItems []string `json:"fatal_item_ids"`
	} `json:"extra"`
	SearchNilInfo struct {
		SearchNilType string `json:"search_nil_type"`
		SearchNilItem string `json:"search_nil_item"`
		IsLoadMore    string `json:"is_load_more"`
		TextType      int    `json:"text_type"`
	} `json:"search_nil_info"`
}

type CommentResponse struct {
	StatusCode           int         `json:"status_code"`
	Comments             []Comment   `json:"comments"`
	Cursor               int         `json:"cursor"`
	HasMore              int         `json:"has_more"`
	ReplyStyle           int         `json:"reply_style"`
	Total                int         `json:"total"`
	Extra                Extra       `json:"extra"`
	LogPB                LogPB       `json:"log_pb"`
	HotsoonFilteredCount int         `json:"hotsoon_filtered_count"`
	UserCommented        int         `json:"user_commented"`
	FastResponseComment  FastComment `json:"fast_response_comment"`
	CommentConfig        interface{} `json:"comment_config"`
	GeneralCommentConfig interface{} `json:"general_comment_config"`
	ShowManagementEntry  int         `json:"show_management_entry_point"`
	CommentCommonData    string      `json:"comment_common_data"`
	FoldedCommentCount   int         `json:"folded_comment_count"`
}

type Comment struct {
	CID               string      `json:"cid"`
	Text              string      `json:"text"`
	AwemeID           string      `json:"aweme_id"`
	CreateTime        int64       `json:"create_time"`
	DiggCount         int         `json:"digg_count"`
	Status            int         `json:"status"`
	User              User        `json:"user"`
	ReplyID           string      `json:"reply_id"`
	UserDigged        int         `json:"user_digged"`
	ReplyComment      []Comment   `json:"reply_comment"`
	LabelText         string      `json:"label_text"`
	LabelType         int         `json:"label_type"`
	ReplyCommentTotal int         `json:"reply_comment_total"`
	ReplyToReplyID    string      `json:"reply_to_reply_id"`
	IsAuthorDigged    bool        `json:"is_author_digged"`
	StickPosition     int         `json:"stick_position"`
	UserBuried        bool        `json:"user_buried"`
	LabelList         interface{} `json:"label_list"`
	IsHot             bool        `json:"is_hot"`
	TextMusicInfo     interface{} `json:"text_music_info"`
	ImageList         interface{} `json:"image_list"`
	IsNoteComment     int         `json:"is_note_comment"`
	IPLabel           string      `json:"ip_label"`
	CanShare          bool        `json:"can_share"`
	ItemCommentTotal  int         `json:"item_comment_total"`
	Level             int         `json:"level"`
	VideoList         interface{} `json:"video_list"`
	SortTags          string      `json:"sort_tags"`
	IsUserTendToReply bool        `json:"is_user_tend_to_reply"`
	ContentType       int         `json:"content_type"`
	IsFolded          bool        `json:"is_folded"`
	EnterFrom         string      `json:"enter_from"`
}

type User struct {
	UID             string `json:"uid"`
	ShortID         string `json:"short_id"`
	Nickname        string `json:"nickname"`
	Signature       string `json:"signature"`
	AvatarLarger    Avatar `json:"avatar_larger"`
	AvatarThumb     Avatar `json:"avatar_thumb"`
	AvatarMedium    Avatar `json:"avatar_medium"`
	IsVerified      bool   `json:"is_verified"`
	FollowStatus    int    `json:"follow_status"`
	AwemeCount      int    `json:"aweme_count"`
	FollowingCount  int    `json:"following_count"`
	FollowerCount   int    `json:"follower_count"`
	FavoritingCount int    `json:"favoriting_count"`
	TotalFavorited  int    `json:"total_favorited"`
	IsBlock         bool   `json:"is_block"`
	Region          string `json:"region"`
	UniqueID        string `json:"unique_id"`
	SecUID          string `json:"sec_uid"`
	AccountRegion   string `json:"account_region"`
}

type Avatar struct {
	URI     string   `json:"uri"`
	URLList []string `json:"url_list"`
	Width   int      `json:"width"`
	Height  int      `json:"height"`
}

type Extra struct {
	Now          int         `json:"now"`
	FatalItemIDs interface{} `json:"fatal_item_ids"`
}

type LogPB struct {
	ImprID string `json:"impr_id"`
}

type FastComment struct {
	ConstantResponseWords []string `json:"constant_response_words"`
	TimedResponseWords    []string `json:"timed_response_words"`
}
