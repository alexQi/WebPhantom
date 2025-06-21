package douyin

import "fmt"

const (
	ELE_AFTER = "following-sibling"
	ELE_BEFOR = "preceding-sibling"
)

const FollowButton = "//button[@data-e2e='user-info-follow-btn']"
const MesageButton = "%s/%s::button[1]"
const CurMsgButton = FollowButton + "/parent::*/button[not(@data-e2e='user-info-follow-btn')]"
const InputElement = "//div[@role='textbox'][@contenteditable='true']"

const BUILD_VISIBLE_CONTAINER = `
	var checkItems = document.evaluate(
		"%s",
		document, null, XPathResult.ORDERED_NODE_SNAPSHOT_TYPE, null
	);
	var optContainer
	for (let i = 0; i < checkItems.snapshotLength; i++) {
		if (checkItems.snapshotItem(i).getBoundingClientRect().x>0){
			optContainer = checkItems.snapshotItem(i).parentElement
		}
	}
`

const GET_ACTION_BUTTON_RECT = `
(function() {
	var el = document.evaluate(
		"%s",
		optContainer, null, XPathResult.FIRST_ORDERED_NODE_TYPE, null
	).singleNodeValue;
	if (el) {
		var rect = el.getBoundingClientRect();
		return {
			bottom: rect.bottom,
			top: rect.top,
			left: rect.left,
			right: rect.right,
			width: rect.width,
			height: rect.height,
			x: rect.x,
			y: rect.y
		};
	} else {
		return {}
	}
})()
`

const SUBSCRIBE_BUTTON_CLICK = `
	var subscribeButton = document.evaluate(
		"%s",
		optContainer, null, XPathResult.FIRST_ORDERED_NODE_TYPE, null
	).singleNodeValue;
	if (subscribeButton) {
		subscribeButton.click();
	} else {
		console.log('Button not found');
	}
`

const MESSAGE_BUTTON_CLICK = `
	var msgButton = document.evaluate(
		"%s",
		optContainer, null, XPathResult.FIRST_ORDERED_NODE_TYPE, null
	).singleNodeValue;
	if (msgButton) {
		msgButton.click();
	} else {
		console.log('msgButton not found');
	}
`

const MESSAGE_RESULT_CHECK = `
	document.querySelector("#messageContent div[data-e2e='msg-item-content'] div").querySelectorAll(":scope > div").length===1
`

const MESSAGE_FAILED_REASON = `
	var failReasonNode = document.querySelector("#messageContent div[data-e2e='msg-item-content']").parentElement.parentElement.nextElementSibling;failReasonNode!=null?failReasonNode.innerText:''
`

const PAGE_ROLL_NEXT_CLICK = `
	document.querySelector("div[data-e2e='video-switch-next-arrow'] span").click()
`

// SearchSortType 表示搜索排序类型
type SearchSortType int

// SearchChannelType 表示搜索频道类型
type SearchChannelType string

const (
	SearchChannelGeneral SearchChannelType = "aweme_general"   // 综合
	SearchChannelVideo   SearchChannelType = "aweme_video_web" // 视频
	SearchChannelUser    SearchChannelType = "aweme_user_web"  // 用户
	SearchChannelLive    SearchChannelType = "aweme_live"      // 直播
)

const (
	SearchSortGeneral  SearchSortType = 0 // 综合排序
	SearchSortMostLike SearchSortType = 1 // 最多点赞
	SearchSortLatest   SearchSortType = 2 // 最新发布
)

// PublishTimeType 表示发布时间类型
type PublishTimeType int

const (
	PublishTimeUnlimited PublishTimeType = 0   // 不限
	PublishTimeOneDay    PublishTimeType = 1   // 一天内
	PublishTimeOneWeek   PublishTimeType = 7   // 一周内
	PublishTimeSixMonth  PublishTimeType = 180 // 半年内
)

// String 方法用于获取 SearchChannelType 的字符串表示
func (s SearchChannelType) String() string {
	return string(s)
}

// String 方法用于获取 SearchSortType 的字符串表示
func (s SearchSortType) String() string {
	return fmt.Sprintf("%d", s)
}

// String 方法用于获取 PublishTimeType 的字符串表示
func (p PublishTimeType) String() string {
	return fmt.Sprintf("%d", p)
}

const (
	// 抖音相关常量
	DOUYIN_INDEX_URL        = "https://www.douyin.com"
	DOUYIN_API_URL          = DOUYIN_INDEX_URL
	DOUYIN_FIXED_USER_AGENT = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36"
	//DOUYIN_FIXED_USER_AGENT = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36"
	DOUYIN_MS_TOKEN_REQ_URL = "https://mssdk.bytedance.com/web/r/token?ms_appid=6383"
	DOUYIN_WEBID_REQ_URL    = "https://mcs.zijieapi.com/webid?aid=6383&sdk_version=5.1.18_zip&device_platform=web"

	// 抖音视频和笔记类型
	DOUYIN_VIDEO_TYPE            = 0
	DOUYIN_NOTE_TYPE             = 68
	DOUYIN_MS_TOKEN_REQ_STR_DATA = `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`
)
