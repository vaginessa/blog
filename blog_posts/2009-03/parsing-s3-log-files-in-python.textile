Id: 13010
Title: Parsing s3 log files in python
Tags: aws,python
Date: 2009-03-08T16:54:37-07:00
Format: Markdown
--------------
After we [took care of compacting s3
logs](/article/Compacting-s3-aws-logs.html), it’s time to parse them. s3
log format is [well
documented](http://docs.amazonwebservices.com/AmazonS3/latest/index.html?LogFormat.html)
and can be parsed with a simple regular expression.

It took me some time to craft it, so here it is. You can also view the
source
[here](http://code.google.com/p/kjk/source/browse/trunk/scripts/test_parse_s3_log.py)

<code>\
\#!/usr/bin/env python\
import re

s3\_line\_logpats = r’(\\S+) (\\S+) \\[(.\*?)\\] (\\S+) (\\S+) ‘ \\\
 r’(\\S+) (\\S+) (\\S+) “([\^”]+)" ‘ \\\
 r’(\\S+) (\\S+) (\\S+) (\\S+) (\\S+) (\\S+) ‘ \\\
 r’“([\^”]*)" “([\^”]*)"’

s3\_line\_logpat = re.compile(s3\_line\_logpats)

(S3\_LOG\_BUCKET\_OWNER, S3\_LOG\_BUCKET, S3\_LOG\_DATETIME,
S3\_LOG\_IP,\
S3\_LOG\_REQUESTOR\_ID, S3\_LOG\_REQUEST\_ID, S3\_LOG\_OPERATION,
S3\_LOG\_KEY,\
S3\_LOG\_HTTP\_METHOD\_URI\_PROTO, S3\_LOG\_HTTP\_STATUS,
S3\_LOG\_S3\_ERROR,\
S3\_LOG\_BYTES\_SENT, S3\_LOG\_OBJECT\_SIZE, S3\_LOG\_TOTAL\_TIME,\
S3\_LOG\_TURN\_AROUND\_TIME, S3\_LOG\_REFERER, S3\_LOG\_USER\_AGENT) =
range(17)

s3\_names = (“bucket\_owner”, “bucket”, “datetime”, “ip”,
“requestor\_id”,\
“request\_id”, “operation”, “key”, “http\_method\_uri\_proto”,
“http\_status”,\
“s3\_error”, “bytes\_sent”, “object\_size”, “total\_time”,
“turn\_around\_time”,\
“referer”, “user\_agent”)

def parse\_s3\_log\_line(line):\
 match = s3\_line\_logpat.match(line)\
 result = [match.group(1+n) for n in range(17)]\
 return result

def dump\_parsed\_s3\_line(parsed):\
 for (name, val) in zip(s3\_names, parsed):\
 print("%s: s" (name, val))

def test():\
 l = r’607c4573f2972c26aff39f7e56ff0490881a35c19b9bf94072cbab8c3219f948
kjkpub [06/Mar/2009:23:13:28 +0000] 41.221.20.231
65a011a29cdf8ec533ec3d1ccaae921c C46E93FF2E865AC1 REST.GET.OBJECT
sumatrapdf/rel/SumatraPDF-0.9.1.zip “GET
/sumatrapdf/rel/SumatraPDF-0.9.1.zip HTTP/1.1” 206 - 43457 1003293 697
611 “http://kjkpub.s3.amazonaws.com/sumatrapdf/rel/” “Mozilla/4.0
(compatible; MSIE 6.0; Windows NT 5.1)”’\
 parsed = parse\_s3\_log\_line(l)\
 dump\_parsed\_s3\_line(parsed)

if *name* == [*main*]()\
 test()\
</code>

This snippet only shows how to break one s3 log line into its components
(`parse_s3_log_line`). Some work is needed to build upon this for
parsing log files and extracting useful information out of them. For
doing that, I recommend techniques described in [Generator Tricks for
Systems Programmers](http://www.dabeaz.com/generators-uk/)
