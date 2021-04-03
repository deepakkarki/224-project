import socket

# pip3 install http-parser
from http_parser.http import HttpStream
from http_parser.reader import SocketReader

class RequestManager:
  def __init__(self, host="127.0.0.1", port=8080):
    self.host = host
    self.port = port
    self.sock = None

  def __enter__(self):
    self.sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    self.sock.connect((self.host, self.port))
    return self

  def __exit__(self, type, value, traceback):
    self.sock.close()

  def send(self, data):
    self.sock.sendall(data)

  def recv(self, size=4096):
    return self.sock.recv(size)

  # other option is read. sock.send(b"1") == 0 or BrokenPipeError => conn dead
  def is_socket_closed(self):
    #NOTE : this will get rid of all the data in the buffer
    try:
      while True:
        # either reads data available or returns "" if conn is closed. if conn
        # is still open and theres an attempt to block, BlockingIOError is thrown
        data = self.sock.recv(8096, socket.MSG_DONTWAIT)
        if len(data) == 0:
          return True
    except BlockingIOError:
      return False  # socket is open and reading from it would block
    except ConnectionResetError:
      return True  # socket was closed for some other reason
    return False

  def send_get(self, url="/", headers={}):
    headers["Host"] = headers.get("Host", self.host)
    req = "GET %s HTTP/1.1\r\n" %url
    for key in headers:
      req += key + ": "+ headers[key] + "\r\n"
    req += "\r\n"
    self.sock.sendall(bytes(req, "ascii"))

  def read_get(self):
    reader = SocketReader(self.sock)
    resp = HttpStream(reader)
    status_code = resp.status_code()
    headers = dict(resp.headers())
    body = resp.body_string()
    return {"status_code": status_code, "headers": headers, "body": body}

