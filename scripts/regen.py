#!/usr/bin/env python

"""
Re-generate html files from markdown (.md) files.
"""

import markdown
from util import read_file_utf8, write_file_utf8, list_files, ext

def is_markdown_file(path):
	return ext(path) in [".md"]

def md_to_html(src, dst):
		s = read_file_utf8(src)
		html = markdown.markdown(s)
		write_file_utf8(dst, html)

def main():
	md_files = list_files("www", is_markdown_file, recur=True)
	for md_file in md_files:
		html_file = md_file[:-2] + "html"
		md_to_html(md_file, html_file)

if __name__ == "__main__":
	main()
