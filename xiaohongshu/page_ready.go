package xiaohongshu

import (
	"time"

	"github.com/go-rod/rod"
	"github.com/sirupsen/logrus"

	"github.com/xpzouying/xiaohongshu-mcp/humanize"
)

// feedReadyTimeout 等待详情页关键容器的上限，超时不算失败。
const feedReadyTimeout = 8 * time.Second

// feedReadySelectors 详情页已出结果的判据：正常笔记容器，或 checkPageAccessible 认的错误容器。
const feedReadySelectors = ".interact-container, .note-scroller, " +
	".access-wrapper, .error-wrapper, .not-found-wrapper, .blocked-wrapper"

// feedDataTimeout 提速时等详情页数据的上限；到点没等到就退回照常的等法。
const feedDataTimeout = 6 * time.Second

// feedDataJS 读帖要的数据到没到："note" 帖子进了 __INITIAL_STATE__，评论也到了（列表里有了，或这篇本来就没有评论）；
// "error" 出了错误页（后面 checkPageAccessible 会报）；"" 还没到。
const feedDataJS = `(id) => {
	if (document.querySelector('.access-wrapper, .error-wrapper, .not-found-wrapper, .blocked-wrapper')) return "error";
	const d = window.__INITIAL_STATE__?.note?.noteDetailMap?.[id];
	if (!d || !d.note || !d.note.noteId) return "";
	const c = d.comments || {};
	if (Array.isArray(c.list) && c.list.length > 0) return "note";
	if (c.firstRequestFinish === true) return "note";
	const n = d.note.interactInfo ? String(d.note.interactInfo.commentCount || "") : "";
	return n === "0" ? "note" : "";
}`

// waitFeedDataReady 读帖用：提速（XHS_PACE_SCALE 小于 1）时不等整页 load（图片也算在里面，常要好几秒），
// 数据一到就读；没开提速、或者到点没等到，照常走 waitFeedPageReady。
// 返回帖子的数据是不是已经确认到了（到了就不用再等错误页出不出来）。
func waitFeedDataReady(page *rod.Page, feedID string) bool {
	if humanize.Fast() {
		deadline := time.Now().Add(feedDataTimeout)
		for time.Now().Before(deadline) {
			if res, err := page.Eval(feedDataJS, feedID); err == nil && res.Value.Str() != "" {
				return res.Value.Str() == "note"
			}
			time.Sleep(150 * time.Millisecond)
		}
		logrus.Warnf("详情页数据未在 %s 内到位，退回等整页加载", feedDataTimeout)
	}
	waitFeedPageReady(page)
	return false
}

// waitFeedPageReady 等待 feed 详情页可用：load 事件 + 关键容器出现，不要求整页 DOM 静止。
func waitFeedPageReady(page *rod.Page) {
	page.MustWaitLoad()

	if _, err := page.Timeout(feedReadyTimeout).Element(feedReadySelectors); err != nil {
		// 没等到也继续，后续 Element 各自带轮询
		logrus.Warnf("详情页关键容器未在 %s 内出现，继续尝试: %v", feedReadyTimeout, err)
	}
}
