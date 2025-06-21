package douyin

import (
	"encoding/json"
	"fmt"
	"noctua/internal/media/douyin"
	"noctua/internal/model"
	"time"
)

type DouyinDataSaver struct {
}

func (d *DouyinDataSaver) HandleMedia(aweme douyin.Aweme, taskId, sourceTaskId, source string) error {
	// 处理视频数据入库
	var awemeUrl string
	if aweme.AwemeType == douyin.DOUYIN_NOTE_TYPE {
		awemeUrl = fmt.Sprintf("%s/note/%s", douyin.DOUYIN_INDEX_URL, aweme.AwemeID)
	} else {
		awemeUrl = fmt.Sprintf("%s/video/%s", douyin.DOUYIN_INDEX_URL, aweme.AwemeID)
	}
	modelMedia := &model.CrawlMedia{
		MediaCode:      "douyin",
		TaskId:         taskId,
		SourceTaskId:   sourceTaskId,
		MediaID:        aweme.AwemeID,
		Type:           aweme.AwemeType,
		Title:          aweme.Desc[:min(1024, len(aweme.Desc))],
		Description:    aweme.Desc,
		SecUID:         aweme.Author.SecUID,
		Nickname:       aweme.Author.Nickname,
		CreateTime:     time.Unix(aweme.CreateTime, 0).In(time.Local),
		LikedCount:     aweme.Statistics.DiggCount,
		CommentCount:   aweme.Statistics.CommentCount,
		ShareCount:     aweme.Statistics.ShareCount,
		CollectedCount: aweme.Statistics.CollectCount,
		URL:            awemeUrl,
		Source:         source,
	}
	err := modelMedia.UpsertModel()
	if err != nil {
		return err
	}
	var avatarUrl = ""
	if len(aweme.Author.AvatarThumb.URLList) > 0 {
		avatarUrl = aweme.Author.AvatarThumb.URLList[0]
	}
	modelCrawlUser := &model.CrawlUser{
		MediaCode:    "douyin",
		TaskId:       taskId,
		SourceTaskId: sourceTaskId,
		Source:       source,
		SecUID:       aweme.Author.SecUID,
		ShortUserID:  aweme.Author.ShortId,
		UserUniqueID: aweme.Author.UniqueId,
		Nickname:     aweme.Author.Nickname,
		Avatar:       avatarUrl,
		Signature:    aweme.Author.Signature,
	}
	err = modelCrawlUser.UpsertModel()
	if err != nil {
		return err
	}
	return nil
}

func (d *DouyinDataSaver) HandleComment(comment douyin.Comment, taskId, sourceTaskId, source string) error {
	pictures, _ := json.Marshal(comment.ImageList)
	modelCrawlComment := &model.CrawlComment{
		MediaCode:       "douyin",
		TaskId:          taskId,
		SourceTaskId:    sourceTaskId,
		MediaID:         comment.AwemeID,
		SecUID:          comment.User.SecUID,
		Nickname:        comment.User.Nickname,
		Location:        comment.IPLabel,
		CommentID:       comment.CID,
		Content:         comment.Text,
		CreateTime:      time.Unix(comment.CreateTime, 0).In(time.Local),
		SubCommentCount: comment.ReplyCommentTotal,
		ParentCommentID: comment.ReplyID, // 父评论ID为空
		LikeCount:       comment.DiggCount,
		Pictures:        string(pictures),
		Source:          source,
	}
	err := modelCrawlComment.UpsertModel()
	if err != nil {
		return err
	}
	var avatarUrl = ""
	if len(comment.User.AvatarThumb.URLList) > 0 {
		avatarUrl = comment.User.AvatarThumb.URLList[0]
	}
	modelCrawlUser := &model.CrawlUser{
		MediaCode:    "douyin",
		TaskId:       taskId,
		SourceTaskId: sourceTaskId,
		Source:       source,
		SecUID:       comment.User.SecUID,
		ShortUserID:  comment.User.ShortID,
		UserUniqueID: comment.User.UniqueID,
		Nickname:     comment.User.Nickname,
		Avatar:       avatarUrl,
		Signature:    comment.User.Signature,
		Location:     comment.IPLabel,
	}
	err = modelCrawlUser.UpsertModel()
	if err != nil {
		return err
	}
	return nil
}

func (d *DouyinDataSaver) HandleUser(user douyin.User, taskId, sourceTaskId, source string) error {
	// TODU: 待调试查询用户信息
	return nil
}
