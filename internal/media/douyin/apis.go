package douyin

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// GetUserInfo 获取用户信息
func (c *DouYinApiClient) GetUserInfo(secUserID string) (map[string]interface{}, error) {
	queryParams := map[string]string{
		"sec_user_id": secUserID,
	}
	headers, err := c.getHeaders()
	if err != nil {
		return nil, err
	}
	headers["Referer"] = "https://www.douyin.com/user/" + secUserID + "?from_tab_name=main"
	resp, err := c.fetch("/aweme/v1/web/user/profile/other/", queryParams, &CallRequestParams{
		NeedSign: true,
		Headers:  headers,
	})
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	err = json.Unmarshal(resp, &data)
	return data, err
}

// SearchInfoByKeyword 关键字搜索
func (c *DouYinApiClient) SearchInfoByKeyword(params *SearchParams) (*SearchResponse, error) {
	queryParams := map[string]string{
		"search_channel":       params.SearchChannel.String(),
		"search_id":            params.SearchId,
		"keyword":              params.Keyword,
		"offset":               fmt.Sprintf("%d", params.Offset),
		"count":                fmt.Sprintf("%d", params.Count),
		"enable_history":       "1",
		"search_source":        "normal_search",
		"query_correct_type":   "1",
		"is_filter_search":     "0",
		"from_group_id":        "",
		"need_filter_settings": "1",
		"list_type":            "multi",
	}
	// 条件判断
	if params.SortType != SearchSortGeneral || params.PublishTimeType != PublishTimeUnlimited {
		queryParams["is_filter_search"] = "1"
		queryParams["search_source"] = "tab_search"
		queryParams["sort_type"] = params.SortType.String()
		queryParams["publish_time"] = params.PublishTimeType.String()
	}

	headers, err := c.getHeaders()
	if err != nil {
		return nil, err
	}
	headers["Referer"] = "https://www.douyin.com/root/search/" + url.QueryEscape(params.Keyword) + "?type=video"
	resp, err := c.fetch("/aweme/v1/web/search/item/", queryParams, &CallRequestParams{
		NeedSign: true,
		Headers:  headers,
	})
	if err != nil {
		return nil, err
	}

	result := &SearchResponse{}
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetVideoByID 获取视频详情
func (c *DouYinApiClient) GetVideoByID(awemeID string) (map[string]interface{}, error) {
	params := map[string]string{"aweme_id": awemeID}
	headers, err := c.getHeaders()
	if err != nil {
		return nil, err
	}
	if _, ok := headers["origin"]; ok {
		delete(headers, "origin")
	}
	resp, err := c.fetch("/aweme/v1/web/aweme/detail/", params, &CallRequestParams{
		NeedSign: true,
		Headers:  headers,
	})
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	err = json.Unmarshal(resp, &data)
	return data, err
}

// GetAwemeComments 获取视频评论
func (c *DouYinApiClient) GetAwemeComments(awemeID string, cursor int, sourceKeyword string) (*CommentResponse, error) {
	queryParams := map[string]string{
		"aweme_id":  awemeID,
		"cursor":    fmt.Sprintf("%d", cursor),
		"count":     "20",
		"item_type": "0",
	}
	headers, err := c.getHeaders()
	if err != nil {
		return nil, err
	}
	headers["Referer"] = "https://www.douyin.com/search/" + url.QueryEscape(sourceKeyword) + "?aid=3a3cec5a-9e27-4040-b6aa-ef548c2c1138&publish_time=0&sort_type=0&source=search_history&type=general"

	resp, err := c.fetch("/aweme/v1/web/comment/list/", queryParams, &CallRequestParams{
		NeedSign: true,
		Headers:  headers,
	})
	if err != nil {
		return nil, err
	}
	result := &CommentResponse{}
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetUserPosts 获取用户的所有视频
func (c *DouYinApiClient) GetUserPosts(secUserID string, maxCursor string) ([]map[string]interface{}, error) {
	queryParams := map[string]string{
		"sec_user_id":  secUserID,
		"max_cursor":   maxCursor,
		"count":        "44",
		"locate_query": "false",
		"verifyFp":     c.verifyParams.VerifyFp,
		"fp":           c.verifyParams.VerifyFp,
	}
	resp, err := c.fetch("/aweme/v1/web/aweme/post/", queryParams, &CallRequestParams{
		NeedSign: true,
	})
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, err
	}

	awemeList, ok := result["aweme_list"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("aweme_list is not an array")
	}

	parsedList := make([]map[string]interface{}, len(awemeList))
	for i, d := range awemeList {
		parsedList[i] = d.(map[string]interface{})
	}
	return parsedList, nil
}
