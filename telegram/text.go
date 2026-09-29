package telegram

import "strings"

func runeCount(s string) int { return len([]rune(s)) }

func runeSlice(s string, start, end int) string {
	r := []rune(s)
	if start >= len(r) {
		return ""
	}
	if end > len(r) {
		end = len(r)
	}
	return string(r[start:end])
}

func chunkMessage(text string, limit int) []string {
	if runeCount(text) <= limit {
		return []string{text}
	}
	paras := strings.Split(text, "\n\n")
	var chunks []string
	var cur string
	curLen := 0
	flushCur := func() {
		if cur != "" {
			chunks = append(chunks, cur)
			cur = ""
			curLen = 0
		}
	}
	for _, ps := range paras {
		psLen := runeCount(ps)
		sepLen := 0
		if curLen > 0 {
			sepLen = 2
		}
		if curLen+sepLen+psLen <= limit {
			if cur != "" {
				cur += "\n\n" + ps
			} else {
				cur = ps
			}
			curLen += sepLen + psLen
		} else {
			if cur != "" {
				flushCur()
			}
			if psLen <= limit {
				cur = ps
				curLen = psLen
			} else {
				lines := strings.Split(ps, "\n")
				var lcur string
				lcurLen := 0
				for _, ls := range lines {
					lsLen := runeCount(ls)
					sep := 0
					if lcurLen > 0 {
						sep = 1
					}
					if lcurLen+sep+lsLen <= limit {
						if lcur != "" {
							lcur += "\n" + ls
						} else {
							lcur = ls
						}
						lcurLen += sep + lsLen
					} else {
						if lcur != "" {
							chunks = append(chunks, lcur)
							lcur = ""
							lcurLen = 0
						}
						// hard cut on rune boundaries
						for runeCount(ls) > limit {
							chunks = append(chunks, runeSlice(ls, 0, limit))
							ls = runeSlice(ls, limit, runeCount(ls))
						}
						lcur = ls
						lcurLen = runeCount(ls)
					}
				}
				if lcur != "" {
					cur = lcur
					curLen = lcurLen
				}
			}
		}
	}
	if cur != "" {
		chunks = append(chunks, cur)
	}
	return chunks
}

// escapeHTML escapes user-supplied text for Telegram's HTML parse mode.
func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
