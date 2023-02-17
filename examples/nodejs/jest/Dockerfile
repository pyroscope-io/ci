FROM node:18.4

COPY package.json ./
COPY yarn.lock ./
RUN yarn install --frozen-lockfile

COPY . .

#ENTRYPOINT ["yarn", "test"]
