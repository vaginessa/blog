Id: 230
Title: apache: enabling mod_rewrite
Tags: unix
Date: 2008-03-13T16:03:47-07:00
Format: Markdown
--------------
**Enabling mod\_rewrite in Apache**:

`LoadModule rewrite_module modules/mod_rewrite.so`

mod\_rewrite:

<code>\
RewriteEngine On\
RewriteBase /\
RewriteRule \^api/(.\*)\$ http://127.0.0.1:8080/api/\$1 [P]\
LoadModule rewrite\_module modules/mod\_rewrite.so\
LoadModule proxy\_module modules/mod\_proxy.so\
LoadModule proxy\_http\_module modules/mod\_proxy\_http.so\
Listen 80\
NameVirtualHost 127.0.0.1\
<VirtualHost 127.0.0.1>\
 ServerName localhost\
 DocumentRoot /var/www\
</VirtualHost>\
</code>
