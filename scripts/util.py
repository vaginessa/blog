import os, codecs

def read_file_utf8(path):
	with codecs.open(path, "r", "utf8") as fo:
		s = fo.read()
	return s

def write_file_utf8(path, s):
	with codecs.open(path, "w", "utf8") as fo:
		s = fo.write(s)

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

def delete_file(path):
    if os.path.exists(path):
        os.remove(path)
