FROM node:18-alpine AS builder
COPY tsconfig.json index.ts package.json package-lock.json /root/
RUN cd /root && npm i && npx tsc

FROM node:18-alpine AS deps
COPY package.json package-lock.json /root/
RUN cd /root && npm i --omit=dev

FROM node:18-alpine
COPY --from=builder /root/index.js /root/index.js.map /root/
COPY --from=deps /root/node_modules /root/node_modules/
WORKDIR /root
ENTRYPOINT node
CMD index.js