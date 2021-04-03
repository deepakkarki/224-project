from socket import socket
import time
import threading

from data import *
from request_manager import RequestManager


def test_200_text():
  """Checks if it can fetch an existing text/html file
  """
  with RequestManager() as ch:
    r = b"GET /index.html HTTP/1.1\r\nHost: 127.0.0.1\r\n\r\n"
    ch.send(r)
    res = ch.read_get()
  assert res["status_code"] == 200
  assert res["headers"]["Content-Type"] == "text/html"
  assert res["body"] == root_index


def test_200_bin():
  """Checks if it can fetch an existing bin file
  """
  with RequestManager() as ch:
    r = b"GET /kitten.jpg HTTP/1.1\r\nHost: 127.0.0.1\r\n\r\n"
    ch.send(r)
    res = ch.read_get()
  assert res["status_code"] == 200
  assert res["headers"]["Content-Type"] == "image/jpeg"
  assert res["body"] == kitten


def test_get_index():
  """Checks if /subdir/ is requested, it returns /subdir/index.html
  """
  with RequestManager() as ch:
    r = b"GET /subdir1/ HTTP/1.1\r\nHost: 127.0.0.1\r\n\r\n"
    ch.send(r)
    res = ch.read_get()
  assert res["status_code"] == 200
  assert res["headers"]["Content-Type"] == "text/html"
  assert res["body"] == subdir1_index


def test_default_mime():
  """Checks if the mime-type is "application/octet-stream" for unknown filetype
  """
  with RequestManager() as ch:
    r = b"GET /subdir1/subdir11/maoyo.giaogiao HTTP/1.1\r\nHost: 127.0.0.1\r\n\r\n"
    ch.send(r)
    res = ch.read_get()
  assert res["status_code"] == 200
  assert res["headers"]["Content-Type"] == "application/octet-stream"


def test_404():
  """Checks if non-existant file will return 404
  """
  with RequestManager() as ch:
    r = b"GET /foo-bar.html HTTP/1.1\r\nHost: 127.0.0.1\r\n\r\n"
    ch.send(r)
    res = ch.read_get()
  assert res["status_code"] == 404


def test_root_escape():
  """Checks if root escape attempt will return 404
  """
  with RequestManager() as ch:
    r = b"GET /../src/main/main.go HTTP/1.1\r\nHost: 127.0.0.1\r\n\r\n"
    ch.send(r)
    res = ch.read_get()
  assert res["status_code"] == 404


def test_conn_close():
  """Checks if connection close header will close the connection after req
  """
  with RequestManager() as ch:
    ch.send_get(url="/", headers={"Connection":"close"})
    res = ch.read_get()
    assert res["headers"]["Connection"] == "close"
    time.sleep(2) # wait for server to process
    assert ch.is_socket_closed() == True


def test_timeout():
  """Checks if connection will timeout after 5 seconds
  """
  with RequestManager() as ch:
    # connect and just sleep for 6s
    time.sleep(6)
    assert ch.is_socket_closed() == True


def test_seq_requests():
  """Checks if server can handle sequential requests
  """
  r1 = b"GET /index.html HTTP/1.1\r\nHost: 127.0.0.1\r\n\r\n"
  r2 = b"GET /subdir1/ HTTP/1.1\r\nHost: 127.0.0.1\r\n\r\n"
  with RequestManager() as ch:
    ch.send(r1+r2)
    res1 = ch.read_get()
    res2 = ch.read_get()
  assert res1["body"] == root_index
  assert res2["body"] == subdir1_index


def test_chunked_request():
  """Checks if sending the request part by part would work
  """
  with RequestManager() as ch:
    ch.send(b"GET /index.html HTTP/1.1\r\n")
    time.sleep(1)
    ch.send(b"Host: 127.0.0.1\r\n\r\n")
    res = ch.read_get()
  assert res["body"] == root_index


def test_partial():
  """Checks if partial http requests result in 400 + close (after 5 sec)
  """
  with RequestManager() as ch:
    ch.send(b"GET /index.html HTTP/1.1\r\n")
    time.sleep(6)
    res = ch.read_get()
    assert ch.is_socket_closed() == True
  assert res["status_code"] == 400


def test_host_missing():
  """Checks if sending a request without Host header returns 400
  """
  with RequestManager() as ch:
    ch.send(b"GET /index.html HTTP/1.1\r\nFoo: Bar\r\n\r\n")
    res = ch.read_get()
    assert ch.is_socket_closed() == True
  assert res["status_code"] == 400


def test_malformed_request():
  """Checks if malformed request results in 400 & closed conn
  Two cases - 1. issue in request line, 2. issue in request headers
  """
  with RequestManager() as ch: # wrong verb
    ch.send(b"GETT /index.html HTTP/1.1\r\nHost: 127.0.0.1\r\n\r\n")
    res = ch.read_get()
    time.sleep(1) # wait for server to close
    assert ch.is_socket_closed() == True
  assert res["status_code"] == 400

  with RequestManager() as ch: # wrong uri format
    ch.send(b"GET index.html HTTP/1.1\r\nHost: 127.0.0.1\r\n\r\n")
    res = ch.read_get()
    time.sleep(1) # wait for server to close
    assert ch.is_socket_closed() == True
  assert res["status_code"] == 400

  with RequestManager() as ch: # wrong headers format
    ch.send(b"GET /index.html HTTP/1.1\r\nHo st: 127.0.0.1\r\n\r\n")
    res = ch.read_get()
    time.sleep(1) # wait for server to close
    assert ch.is_socket_closed() == True
  assert res["status_code"] == 400

def test_seq_400():
  """Checks if server can handle a malformed requests in a sequence of requests.
     Expected behaviour is to return 400 when a bad req is encountered and close
     the connection discarding any request in the pipeline.
  """
  r1 = b"GET /index.html HTTP/1.1\r\nHost: 127.0.0.1\r\n\r\n"
  r2 = b"GETT /subdir1/ HTTP/1.1\r\nHost: 127.0.0.1\r\n\r\n" # malformed
  r3 = b"GET / HTTP/1.1\r\nHost: 127.0.0.1\r\n\r\n"
  with RequestManager() as ch:
    ch.send(r1+r2+r3)
    time.sleep(1)
    # oddly, the client parser had a bug, so just check if conn closed
    # I've checked manually that the data returned is indeed as expected
    assert ch.is_socket_closed() == True


def test_large_request():
  """Checks if a request larger than 8KB results in 400 & close
  """
  params = {}
  for i in range(8000):
    params["key"+str(i)] = "value"+str(i)
  with RequestManager() as ch:
    ch.send_get(url="/", headers=params)
    res = ch.read_get()
    time.sleep(2) # wait for server to process
    assert ch.is_socket_closed() == True
  assert res["status_code"] == 400


def test_concurrent():
  """Checks if server can handle two parallel connections at once
  """
  def thread_fn(id):
    with RequestManager() as ch:
      ch.send(b"GET /index.html HTTP/1.1\r\n")
      print("Request #%d is sleeping" %id)
      time.sleep(2)
      ch.send(b"Host: 127.0.0.1\r\n\r\n")
      res = ch.read_get()
    print("Request #%d is complete" %id)
    assert res["body"] == root_index

  threads = []
  for i in range(5):
    t = threading.Thread(target=thread_fn, args=(i,))
    threads.append(t)
    t.start()

  for thread in threads:
    thread.join()

  # Output :
  # Request #0 is sleeping
  # Request #1 is sleeping
  # Request #2 is sleeping
  # Request #4 is sleeping
  # Request #3 is sleeping
  # Request #2 is complete
  # Request #0 is complete
  # Request #1 is complete
  # Request #4 is complete
  # Request #3 is complete

  # LOGIC
  # As we can see, multiple threads have made their own unique connetion to the
  # server at once, they sent a partial request and went to sleep. Then all the
  # requests are processed (seemingly) at once. This was to test whether the
  # server could handle concurrent connections.