Id: 705001
Title: How to fix cygwin ssl certificate
Tags: note,ssh,cygwin
Date: 2011-10-13T15:45:07-07:00
Format: Markdown
--------------
After updating cygwin, you might get the following error when doing
something related to ssh:

<code>\
error: SSL certificate problem, verify that the CA cert is OK.\
Details: error:14090086:SSL
routines:SSL3\_GET\_SERVER\_CERTIFICATE:certificate\
verify failed while accessing <something>\
</code>

To fix it:\
<code>\
cd /usr/ssl/certs\
curl http://curl.haxx.se/ca/cacert.pem | awk
‘split\_after==1{n++;split\_after=0} /——~~END CERTIFICATE——~~/
{split\_after=1} {print \> “cert” n “.pem”}’\
c\_rehash\
</code>
