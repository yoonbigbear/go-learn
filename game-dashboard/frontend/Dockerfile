# 1. 빌드 단계
FROM node:18-alpine AS builder
WORKDIR /app
COPY ./game-dashboard/frontend/package.json ./game-dashboard/frontend/package-lock.json ./
RUN npm install
COPY /game-dashboard/frontend/ .
RUN npm run build

# 2. 실행 단계 (Nginx)
FROM nginx:alpine
# 빌드된 결과물을 Nginx 폴더로 복사
COPY --from=builder /app/dist /usr/share/nginx/html
# 위에서 만든 설정 파일 덮어쓰기
COPY ./game-dashboard/frontend/nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
