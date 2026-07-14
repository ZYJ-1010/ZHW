FROM node:20-alpine

WORKDIR /app

COPY package.json server.js ./
COPY scripts ./scripts
COPY src ./src

RUN node scripts/build.js

ENV ADMIN_WEB_PORT=5173
ENV NODE_ENV=production

EXPOSE 5173

CMD ["node", "server.js"]
