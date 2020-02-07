from ctypes import *
from ctypes import cdll

class go_string(Structure):
    _fields_ = [
        ("p", c_char_p),
        ("n", c_int)]

lib = cdll.LoadLibrary('./pychtseg.so')

def callchtseg(eng, conn, istr):
	char_str = c_char_p(istr.encode('utf-8'))
	size_str = len(istr.encode('utf-8'))
	inp = go_string(char_str, size_str)
	in_eng = go_string(c_char_p(eng.encode('utf-8')), len(eng))
	in_conn = go_string(c_char_p(conn.encode('utf-8')), len(conn))
	lib.Getchtseg.restype = c_char_p
	x = lib.Getchtseg(in_eng, in_conn, inp)
	y = json.loads(x.decode('utf-8'))
	print(y)
	

callchtseg('sqlite3', '../db/chtseg.db', '現貨附發票 Raspberry Pi 樹莓派專用 USB電腦遙控器 帶無線鼠標無線鍵盤功能 萬能PC/紅外線遙控器')
