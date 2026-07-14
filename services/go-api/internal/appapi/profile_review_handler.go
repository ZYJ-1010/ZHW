package appapi

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"zhw-mini/services/go-api/internal/common/httpx"
	"zhw-mini/services/go-api/internal/games"
	"zhw-mini/services/go-api/internal/reviews"
)

type profileReviewReply struct {
	ReviewID  int64     `json:"reviewId"`
	Content   string    `json:"content"`
	RepliedAt time.Time `json:"repliedAt"`
}

func (s *Server) profileReviewList(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	items := s.profileReviewItems(userID)

	httpx.OK(w, map[string]interface{}{
		"score":        profileReviewScore(items),
		"pendingCount": countPendingProfileReviews(items),
		"stats":        profileReviewStats(items),
		"reviews":      items,
		"templates":    profileReviewTemplates(),
	})
}

func (s *Server) profileReviewDetail(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	reviewID, ok := idFromAdminPath(w, r.URL.Path, "/api/app/profile/service-center/reviews/", "")
	if !ok {
		return
	}
	item, found := s.findProfileReviewItem(userID, reviewID)
	if !found {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "评价不存在")
		return
	}

	httpx.OK(w, map[string]interface{}{
		"review":       item,
		"templates":    profileReviewTemplates(),
		"templateRows": templateRows(profileReviewTemplates()),
		"history":      profileReviewHistory(item),
	})
}

func (s *Server) routeProfileReviewPost(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.HasSuffix(r.URL.Path, "/reply"):
		s.replyProfileReview(w, r)
	case strings.HasSuffix(r.URL.Path, "/like"):
		s.likeProfileReview(w, r)
	case strings.HasSuffix(r.URL.Path, "/actions"):
		s.profileReviewActions(w, r)
	default:
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "review action not found")
	}
}

func (s *Server) replyProfileReview(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	reviewID, ok := idFromAdminPath(w, r.URL.Path, "/api/app/profile/service-center/reviews/", "/reply")
	if !ok {
		return
	}
	if _, found := s.findProfileReviewItem(userID, reviewID); !found {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "评价不存在")
		return
	}
	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidationError, "请求参数错误")
		return
	}
	content := strings.TrimSpace(req.Content)
	if content == "" || len([]rune(content)) > 200 {
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeValidationError, "回复内容需为1-200字")
		return
	}

	reply := profileReviewReply{ReviewID: reviewID, Content: content, RepliedAt: time.Now()}
	s.reviewReplyMu.Lock()
	s.reviewReplies[reviewID] = reply
	s.reviewReplyMu.Unlock()
	s.recordBehavior(userID, "reply_profile_review", "review", reviewID, nil)

	httpx.OK(w, map[string]interface{}{
		"reviewId":   strconv.FormatInt(reviewID, 10),
		"reply":      profileReviewReplyDTO(reply),
		"statusType": "replied",
	})
}

func (s *Server) likeProfileReview(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	reviewID, ok := idFromAdminPath(w, r.URL.Path, "/api/app/profile/service-center/reviews/", "/like")
	if !ok {
		return
	}
	if _, found := s.findProfileReviewItem(userID, reviewID); !found {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "评价不存在")
		return
	}
	liked := s.toggleProfileReviewLike(userID, reviewID)
	action := "unlike_profile_review"
	if liked {
		action = "like_profile_review"
	}
	s.recordBehavior(userID, action, "review", reviewID, nil)
	httpx.OK(w, map[string]interface{}{
		"reviewId": strconv.FormatInt(reviewID, 10),
		"liked":    liked,
	})
}

func (s *Server) profileReviewActions(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireIdentityUser(w, r)
	if !ok {
		return
	}
	reviewID, ok := idFromAdminPath(w, r.URL.Path, "/api/app/profile/service-center/reviews/", "/actions")
	if !ok {
		return
	}
	item, found := s.findProfileReviewItem(userID, reviewID)
	if !found {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "评价不存在")
		return
	}
	statusType, _ := item["statusType"].(string)
	actions := []map[string]interface{}{
		{"key": "detail", "text": "查看关联局", "route": item["detailUrl"]},
	}
	if statusType != "replied" {
		actions = append(actions, map[string]interface{}{"key": "reply", "text": "回复评价", "route": "/pages/profile/service-center/manage/review-reply/index?id=" + strconv.FormatInt(reviewID, 10)})
	}
	if reportURL, _ := item["reportUrl"].(string); reportURL != "" {
		actions = append(actions, map[string]interface{}{"key": "intervention", "text": "申请平台介入", "route": reportURL})
	}
	s.recordBehavior(userID, "open_profile_review_actions", "review", reviewID, map[string]interface{}{"count": len(actions)})
	httpx.OK(w, map[string]interface{}{
		"reviewId": strconv.FormatInt(reviewID, 10),
		"actions":  actions,
	})
}

func (s *Server) profileReviewItems(userID int64) []map[string]interface{} {
	items := make([]map[string]interface{}, 0)
	for _, review := range s.reviews.AllReviews() {
		if review.TargetUserID != userID {
			continue
		}
		game, err := s.games.Get(review.GameID)
		if err != nil {
			continue
		}
		items = append(items, s.profileReviewItem(review, game))
	}
	return items
}

func (s *Server) findProfileReviewItem(userID int64, reviewID int64) (map[string]interface{}, bool) {
	for _, review := range s.reviews.AllReviews() {
		if review.ID != reviewID || review.TargetUserID != userID {
			continue
		}
		game, err := s.games.Get(review.GameID)
		if err != nil {
			return nil, false
		}
		return s.profileReviewItem(review, game), true
	}
	return nil, false
}

func (s *Server) profileReviewItem(review reviews.Review, game games.Game) map[string]interface{} {
	userName := s.inGameDisplayName(review.ReviewerUserID, "玩家")
	reply, replied := s.profileReviewReply(review.ID)
	statusType := "pending"
	if replied {
		statusType = "replied"
	} else if review.Score <= 2 {
		statusType = "critical"
	} else if review.Score <= 3 {
		statusType = "neutral"
	}
	item := map[string]interface{}{
		"id":             strconv.FormatInt(review.ID, 10),
		"reviewId":       review.ID,
		"gameId":         review.GameID,
		"reviewerUserId": review.ReviewerUserID,
		"targetUserId":   review.TargetUserID,
		"user":           userName,
		"realName":       userName,
		"displayName":    userName,
		"avatar":         avatarTextForName(userName, review.ReviewerUserID),
		"avatarText":     avatarTextForName(userName, review.ReviewerUserID),
		"avatarClass":    profileReviewAvatarClass(review.Score),
		"time":           review.CreatedAt.Format("01-02"),
		"timeText":       review.CreatedAt.Format("01-02 15:04"),
		"rating":         starsText(review.Score),
		"score":          review.Score,
		"title":          game.Title,
		"content":        defaultString(review.Content, "用户未填写文字评价"),
		"tags":           review.Tags,
		"reply":          "",
		"liked":          s.profileReviewLiked(review.TargetUserID, review.ID),
		"statusType":     statusType,
		"orderNo":        serviceOrderID(review.GameID),
		"amount":         serviceAmountText(successFundAmount(game)),
		"detailUrl":      "/pages/game/detail/index?id=" + strconv.FormatInt(review.GameID, 10),
		"reportUrl": "/pages/profile/system-management/report-center/index?gameId=" +
			strconv.FormatInt(review.GameID, 10) +
			"&targetUserId=" + strconv.FormatInt(review.ReviewerUserID, 10) +
			"&targetName=" + url.QueryEscape(userName) +
			"&reviewId=" + strconv.FormatInt(review.ID, 10),
	}
	if review.Score <= 2 {
		item["alert"] = "低分评价需及时回复，必要时可申请平台介入"
	}
	if replied {
		item["reply"] = reply.Content
		item["replyInfo"] = profileReviewReplyDTO(reply)
	}
	return item
}

func (s *Server) toggleProfileReviewLike(userID int64, reviewID int64) bool {
	s.reviewReplyMu.Lock()
	defer s.reviewReplyMu.Unlock()
	if s.reviewLikes == nil {
		s.reviewLikes = make(map[int64]map[int64]bool)
	}
	if s.reviewLikes[userID] == nil {
		s.reviewLikes[userID] = make(map[int64]bool)
	}
	next := !s.reviewLikes[userID][reviewID]
	s.reviewLikes[userID][reviewID] = next
	return next
}

func (s *Server) profileReviewLiked(userID int64, reviewID int64) bool {
	s.reviewReplyMu.RLock()
	defer s.reviewReplyMu.RUnlock()
	return s.reviewLikes != nil && s.reviewLikes[userID] != nil && s.reviewLikes[userID][reviewID]
}

func (s *Server) profileReviewReply(reviewID int64) (profileReviewReply, bool) {
	s.reviewReplyMu.RLock()
	defer s.reviewReplyMu.RUnlock()
	reply, ok := s.reviewReplies[reviewID]
	return reply, ok
}

func profileReviewScore(items []map[string]interface{}) map[string]interface{} {
	total := len(items)
	sum := 0
	good := 0
	breakdownCount := map[int]int{}
	for _, item := range items {
		score, _ := item["score"].(int)
		sum += score
		if score >= 4 {
			good++
		}
		breakdownCount[score]++
	}
	overall := "0.0"
	goodRate := "0%"
	stars := starsText(0)
	if total > 0 {
		overall = strconv.FormatFloat(float64(sum)/float64(total), 'f', 1, 64)
		goodRate = strconv.Itoa(good*100/total) + "%"
		stars = starsText(minInt(5, maxInt(1, roundedScore(sum, total))))
	}
	return map[string]interface{}{
		"overall":  overall,
		"total":    strconv.Itoa(total),
		"goodRate": goodRate,
		"stars":    stars,
		"breakdown": []map[string]interface{}{
			scoreBreakdownRow("5星", breakdownCount[5], total),
			scoreBreakdownRow("4星", breakdownCount[4], total),
			scoreBreakdownRow("3星", breakdownCount[3], total),
		},
	}
}

func profileReviewStats(items []map[string]interface{}) []map[string]interface{} {
	total := len(items)
	pending := countPendingProfileReviews(items)
	replied := total - pending
	good := 0
	for _, item := range items {
		if score, _ := item["score"].(int); score >= 4 {
			good++
		}
	}
	replyRate := "0%"
	goodRate := "0%"
	if total > 0 {
		replyRate = strconv.Itoa(replied*100/total) + "%"
		goodRate = strconv.Itoa(good*100/total) + "%"
	}
	return []map[string]interface{}{
		{"value": strconv.Itoa(total), "label": "近30天新增评价"},
		{"value": replyRate, "label": "回复率"},
		{"value": goodRate, "label": "好评率"},
		{"value": "2小时", "label": "平均响应"},
	}
}

func countPendingProfileReviews(items []map[string]interface{}) int {
	count := 0
	for _, item := range items {
		if item["statusType"] == "pending" || item["statusType"] == "critical" || item["statusType"] == "neutral" {
			count++
		}
	}
	return count
}

func profileReviewTemplates() []string {
	return []string{"感谢您的认可！", "期待下次合作", "有问题随时联系", "我们会继续努力", "感谢反馈，已改进"}
}

func templateRows(templates []string) [][]string {
	rows := make([][]string, 0)
	for index, template := range templates {
		if index%2 == 0 {
			rows = append(rows, []string{template})
			continue
		}
		rows[len(rows)-1] = append(rows[len(rows)-1], template)
	}
	return rows
}

func profileReviewHistory(item map[string]interface{}) []map[string]interface{} {
	history := []map[string]interface{}{
		{"role": item["user"].(string) + "（玩家）", "avatar": item["avatar"], "time": item["timeText"], "content": item["content"], "side": "user"},
	}
	if reply, _ := item["reply"].(string); strings.TrimSpace(reply) != "" {
		history = append(history, map[string]interface{}{"role": "我（行家）", "avatar": "我", "time": "已回复", "content": reply, "side": "expert"})
	}
	return history
}

func profileReviewReplyDTO(reply profileReviewReply) map[string]interface{} {
	return map[string]interface{}{
		"content":   reply.Content,
		"repliedAt": reply.RepliedAt.Format(time.RFC3339),
		"timeText":  reply.RepliedAt.Format("01-02 15:04"),
	}
}

func profileReviewAvatarClass(score int) string {
	if score <= 2 {
		return "red"
	}
	if score <= 3 {
		return "indigo"
	}
	return "green"
}

func starsText(score int) string {
	if score < 0 {
		score = 0
	}
	if score > 5 {
		score = 5
	}
	return strings.Repeat("★", score) + strings.Repeat("☆", 5-score)
}

func scoreBreakdownRow(label string, count int, total int) map[string]interface{} {
	percent := 0
	if total > 0 {
		percent = count * 100 / total
	}
	return map[string]interface{}{"label": label, "percent": strconv.Itoa(percent) + "%", "style": "width: " + strconv.Itoa(percent) + "%;"}
}

func roundedScore(sum int, total int) int {
	if total == 0 {
		return 0
	}
	return int(float64(sum)/float64(total) + 0.5)
}
