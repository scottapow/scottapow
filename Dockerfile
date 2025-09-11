FROM node:lts-bullseye AS runtime
WORKDIR /

COPY . .

RUN npm install
RUN npm run build

ENV HOST=0.0.0.0
ENV PORT=8080
EXPOSE 8080
CMD ["HOST=0.0.0.0", "node", "./dist/server/entry.mjs"]