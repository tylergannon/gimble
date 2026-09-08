import os
import socket

server = socket.socket()
server.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
server.bind(("127.0.0.1", int(os.environ["PORT"])))
server.listen(8)

while True:
    connection, _ = server.accept()
    connection.recv(4096)
    connection.sendall(
        b"HTTP/1.1 200 OK\r\n"
        b"Content-Length: 2\r\n"
        b"Connection: close\r\n\r\n"
        b"ok"
    )
    connection.close()
