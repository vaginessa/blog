from google.appengine.ext import db
import main
import codecs, hashlib, os, os.path

g_data_dir = os.path.join("..", "..", "blogimported")

assert os.path.exists(g_data_dir)

def touni(v):
	if not isinstance(v, basestring):
		return unicode(v)
	if isinstance(v, unicode): return v
	return unicode(v, "utf-8")

def kv(k, v):
	return u"%s: %s" % (touni(k), touni(v))

def sep(): return u""

def long2ip(val):
	slist = []
	for x in range(0,4):
		slist.append(str(int(val >> (24 - (x * 8)) & 0xFF)))
	return ".".join(slist)

def create_dir(path):
	if not os.path.exists(path):
		os.makedirs(path)

def write_bin_to_file(path, data):
	open(path, "wb").write(data)

def write_str_to_file(path, data):
	write_bin_to_file(path, data.encode("utf8"))

def read_bin_from_file(path):
	if not os.path.exists(path): return None
	return open(path, "rb").read()

def read_str_from_file(path):
	return read_bin_from_file(path)

def save_msg_sha1_to_dir(msg, dir):
	data = msg.encode("utf8")
	m = hashlib.sha1()
	m.update(data)
	sha1 = m.hexdigest()
	file_dir = os.path.join(g_data_dir, dir, sha1[:2], sha1[2:4])
	create_dir(file_dir)
	file_path = os.path.join(file_dir, sha1)
	if not os.path.exists(file_path):
		write_bin_to_file(file_path, data)
	return sha1

def save_msg_sha1(msg):
	return save_msg_sha1_to_dir(msg, "blobs")

def save_msg_tmp_sha1(msg):
	return save_msg_sha1_to_dir(msg, "blobs_tmp")

def sercrash(e):
	try:
		s = touni(e.data)
	except:
		return ""
	sha1 = save_msg_tmp_sha1(s)
	lines = [
		kv("M", sha1),
		kv("On", e.created_on),
		kv("Ip", e.ip_addr),
		kv("N", e.app_name),
		kv("V", e.app_ver),
		sep(), sep()
	]
	return u"\n".join(lines)

def sertext(e):
	s = touni(e.content)
	sha1 = save_msg_sha1(s)
	lines = [
		kv("I", str(e.key().id())),
		kv("M", sha1),
		kv("On", e.published_on),
		kv("F", e.format),
		sep(), sep()
	]
	return u"\n".join(lines)

def crashes(count=-1, batch_size=501):
	file_path = os.path.join(g_data_dir, "crashes.txt")
	last_key_filepath = os.path.join(g_data_dir, "crashes_last_key_id.txt")

	last_key_id = read_str_from_file(last_key_filepath)
	if None == last_key_id:
		print("Loading crashes from the beginning")
		entities = main.CrashReports.all().fetch(batch_size)
	else:
		last_key = db.Key.from_path('CrashReports', long(last_key_id))
		print("Loading crashes from key %s" % last_key_id)
		entities = main.CrashReports.all().filter('__key__ >', last_key).fetch(batch_size)

	last_key = None
	if len(entities) == 0:
		print("There are no new crashes")

	f = open(file_path, "a")
	n = 0
	while entities:
		for e in entities:
			last_key = e.key()
			s = sercrash(e)
			f.write(s.encode("utf8"))
			n += 1
			if count > 0 and n > count:
				break
			if n % 100 == 0:
				print("%d crashes" % n)
			if count > 0 and n >= count:
				entities = None
				break
		if entities is None:
			break
		entities = main.CrashReports.all().filter('__key__ >', last_key).fetch(batch_size)
	f.close()
	if last_key != None:
		print("New last crashes key id: %d" % last_key.id())
		write_str_to_file(last_key_filepath, str(last_key.id()))

def texts(count=-1, batch_size=501):
	file_path = os.path.join(g_data_dir, "texts.txt")
	last_key_filepath = os.path.join(g_data_dir, "texts_last_key_id.txt")

	last_key_id = read_str_from_file(last_key_filepath)
	if None == last_key_id:
		print("Loading texts from the beginning")
		entities = main.TextContent.all().fetch(batch_size)
	else:
		last_key = db.Key.from_path('TextContent', long(last_key_id))
		print("Loading texts from key %s" % last_key_id)
		entities = main.TextContent.all().filter('__key__ >', last_key).fetch(batch_size)

	last_key = None
	if len(entities) == 0:
		print("There are no new texts")

	f = open(file_path, "a")
	n = 0
	while entities:
		for e in entities:
			last_key = e.key()
			s = sertext(e)
			f.write(s.encode("utf8"))
			n += 1
			if count > 0 and n > count:
				break
			if n % 100 == 0:
				print("%d texts" % n)
			if count > 0 and n >= count:
				entities = None
				break
		if entities is None:
			break
		entities = main.TextContent.all().filter('__key__ >', last_key).fetch(batch_size)
	f.close()
	if last_key != None:
		print("New last text key id: %d" % last_key.id())
		write_str_to_file(last_key_filepath, str(last_key.id()))

def serarticle(e):
	keys = [str(k.id()) for k in e.previous_versions]
	lines = [
		kv("P1", e.permalink),
		kv("P2", e.permalink2),
		kv("P?", e.is_public),
		kv("D?", e.is_deleted),
		kv("T", touni(e.title)),
		kv("TG", u",".join(e.tags)),
		kv("V", u",".join(keys)),
		sep(), sep()
	]
	return u"\n".join(lines)

def articles(count=-1, batch_size=501):
	file_path = os.path.join(g_data_dir, "articles.txt")
	last_key_filepath = os.path.join(g_data_dir, "articles_last_key_id.txt")

	last_key_id = read_str_from_file(last_key_filepath)
	if None == last_key_id:
		print("Loading articles from the beginning")
		entities = main.Article.all().fetch(batch_size)
	else:
		last_key = db.Key.from_path('Article', long(last_key_id))
		print("Loading articles from key %s" % last_key_id)
		entities = main.Article.all().filter('__key__ >', last_key).fetch(batch_size)

	last_key = None
	if len(entities) == 0:
		print("There are no new articles")

	f = open(file_path, "a")
	n = 0
	while entities:
		for e in entities:
			last_key = e.key()
			s = serarticle(e)
			f.write(s.encode("utf8"))
			n += 1
			if count > 0 and n > count:
				break
			if n % 100 == 0:
				print("%d articles" % n)
			if count > 0 and n >= count:
				entities = None
				break
		if entities is None:
			break
		entities = main.Article.all().filter('__key__ >', last_key).fetch(batch_size)
	f.close()
	if last_key != None:
		print("New last articles key id: %d" % last_key.id())
		write_str_to_file(last_key_filepath, str(last_key.id()))

def do():
	articles()
	texts()
	crashes()

