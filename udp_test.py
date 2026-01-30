import socket

# --- 설정 ---
SERVER_IP = "127.0.0.1"
SERVER_PORT = 7015
MESSAGE = b"Hello, Game Server!" # 보낼 메시지 (bytes 형태)
# ------------

# 1. UDP 소켓 생성
sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
print(f"UDP 소켓을 생성했습니다.")

# 2. 서버로 메시지 전송
try:
    print(f"'{SERVER_IP}:{SERVER_PORT}' 주소로 메시지 전송 시도...")
    sock.sendto(MESSAGE, (SERVER_IP, SERVER_PORT))
    print(f"✅ 성공: '{MESSAGE.decode()}' 메시지를 성공적으로 보냈습니다.")
    print("   (서버로부터 응답이 없는 것은 정상일 수 있습니다)")

except Exception as e:
    print(f"❌ 실패: 메시지 전송 중 에러 발생 - {e}")

finally:
    # 3. 소켓 닫기
    sock.close()
    print("소켓을 닫았습니다.")