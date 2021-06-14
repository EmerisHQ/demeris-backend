FROM node:14

WORKDIR /usr/src/app

COPY package.json .
RUN yarn

COPY . .
RUN yarn build

ENV HOST=0.0.0.0
ENV PORT=80
ENV CNS_URL=http://cns-server:8000
ENV BASE_URL="/"

EXPOSE ${PORT}

CMD [ "yarn", "start" ]