FROM node:lts-bullseye AS runtime
WORKDIR /

COPY . .

RUN npm install
RUN npm run build

ENV HOST=scottpowell.dev
ENV PORT=3000
EXPOSE 3000
CMD ["node", "./dist/server/entry.mjs"]