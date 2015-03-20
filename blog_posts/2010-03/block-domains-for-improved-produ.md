Id: 63001
Title: Block domains for improved productivity
Tags: note
Date: 2010-03-03T21:05:54-08:00
Format: Markdown
--------------
Through the magic of `/etc/hosts` I blacklisted many news sites I visit
compulsively. So far looks like a good idea in battling my growing
Internet ADD.

Here’s my list:

<code>\
66.135.33.106 engadget.com www.engadget.com news.ycombinator.com\
66.135.33.106 techmeme.com www.techmeme.com techcrunch.com\
66.135.33.106 www.imdb.com imdb.com daringfireball.net www.hulu.com\
66.135.33.106 news.google.com www.tvsquad.comn twitter.com\
</code>

I could have used 127.0.0.1 - 66.135.33.106 is someone else’s server
that displays a message that a domain was blocked via `/etc/hosts`,
which is nice if you forgot about blocking. On the downside,
66.135.33.106 is often down.
