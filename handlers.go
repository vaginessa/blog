package main

import "net/http"

func InitHttpHandlers() {
	http.Handle("/", makeTimingHandler(handleMainPage))
	http.HandleFunc("/favicon.ico", handleFavicon)
	http.HandleFunc("/robots.txt", handleRobotsTxt)
	http.HandleFunc("/contactme.html", handleContactme)
	http.HandleFunc("/logs", handleLogs)
	http.HandleFunc("/timings", handleTimings)
	http.HandleFunc("/oauthtwittercb", handleOauthTwitterCallback)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/logout", handleLogout)

	http.Handle("/app/crashsubmit", makeTimingHandler(handleCrashSubmit))
	http.Handle("/app/crashes", makeTimingHandler(handleCrashes))
	http.Handle("/app/crashesrss", makeTimingHandler(handleCrashesRss))
	http.Handle("/app/crashshow", makeTimingHandler(handleCrashShow))
	// TODO: I stopped pointing people to FeedBurner feed on 2013-05-22
	// At some point I should delete /feedburner.xml, which is a source data
	// for FeedBurner
	http.Handle("/feedburner.xml", makeTimingHandler(handleAtom))
	http.Handle("/atom.xml", makeTimingHandler(handleAtom))
	http.Handle("/atom-all.xml", makeTimingHandler(handleAtomAll))
	http.Handle("/archives.html", makeTimingHandler(handleArchives))
	http.Handle("/software", makeTimingHandler(handleSoftware))
	http.Handle("/software/", makeTimingHandler(handleSoftware))
	http.Handle("/extremeoptimizations/", makeTimingHandler(handleExtremeOpt))
	http.Handle("/article/", makeTimingHandler(handleArticle))
	http.Handle("/kb/", makeTimingHandler(handleArticle))
	http.Handle("/blog/", makeTimingHandler(handleArticle))
	http.Handle("/forum_sumatra/", makeTimingHandler(forumRedirect))
	http.Handle("/articles/", makeTimingHandler(handleArticles))
	http.Handle("/tag/", makeTimingHandler(handleTag))
	http.Handle("/static/", makeTimingHandler(handleStatic))
	http.Handle("/css/", makeTimingHandler(handleCss))
	http.Handle("/js/", makeTimingHandler(handleJs))
	http.Handle("/gfx/", makeTimingHandler(handleGfx))
	http.Handle("/markitup/", makeTimingHandler(handleMarkitup))
	http.Handle("/djs/", makeTimingHandler(handleDjs))
	http.Handle("/metrics", makeTimingHandler(handleMetrics))
	http.HandleFunc("/ws", serveWs)
	http.HandleFunc("/tool/protobufs-online-decoder", handleProtobufDecoder)
}
