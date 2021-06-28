FROM node:14 as builder

WORKDIR /usr/src/app

COPY package.json .
RUN yarn

COPY . .
RUN yarn generate

FROM nginx:alpine
COPY --from=builder /usr/src/app/dist /usr/share/nginx/html/admin