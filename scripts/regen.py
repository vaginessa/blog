#!/usr/bin/env python

"""
Re-generate html files from markdown (.md) files.
"""

import os, codecs, markdown

def ext(path):
    return os.path.splitext(path)[1].lower()

# returns full paths of files in a given directory, potentially recursively,
# potentially filtering file names by filter_func (which takes file path as
# an argument)
def list_files_g(d, filter_func=None, recur=False):
    to_visit = [d]
    while len(to_visit) > 0:
        d = to_visit.pop(0)
        for f in os.listdir(d):
            path = os.path.join(d, f)
            isdir = os.path.isdir(path)
            if isdir:
                if recur:
                    to_visit.append(path)
            else:
                if filter_func != None:
                    if filter_func(path):
                        yield path
                else:
                    yield path

# generator => array
def list_files(d, filter_func=None, recur=False):
    return [path for path in list_files_g(d, filter_func, recur)]

def is_markdown_file(path):
	return ext(path) in [".md"]

def read_file_utf8(path):
	with codecs.open(path, "r", "utf8") as fo:
		s = fo.read()
	return s

def write_file_utf8(path, s):
	with codecs.open(path, "w", "utf8") as fo:
		s = fo.write(s)

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
